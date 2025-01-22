// Copyright 2024 Torben Schinke. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package enum

import (
	"fmt"
	"iter"
	"reflect"
	"slices"
	"sync"
)

type JSON interface {
	jsonOption()
}

type AdjacentlyOptions struct {
	Tag     string // name of the tag key in json
	Content string // name of the content key in json
}

func (AdjacentlyOptions) jsonOption() {}

type InternallyOptions struct {
	Tag string // name of the tag key in json
}

func (InternallyOptions) jsonOption() {}

type ExternallyOptions struct {
}

func (ExternallyOptions) jsonOption() {}

type UntaggedOptions struct {
}

func (UntaggedOptions) jsonOption() {}

type Declaration struct {
	ifaceT   reflect.Type
	variants []reflect.Type
	cfg      enumCfg
}

func (d Declaration) NoZero() bool {
	return d.cfg.NoZero()
}

func (d Declaration) Variants() iter.Seq[reflect.Type] {
	return slices.Values(d.variants)
}

func (d Declaration) EnumType() reflect.Type {
	return d.ifaceT
}

func (d Declaration) JSON() JSON {
	return d.cfg.jsonOpts
}

func (d Declaration) Name(t reflect.Type) (string, bool) {
	s, ok := d.cfg.fromTypeToName[t]
	return s, ok
}

func (d Declaration) Type(name string) (reflect.Type, bool) {
	t, ok := d.cfg.fromNameToType[name]
	return t, ok
}

var globalDeclContext = &declContext{
	decls: make(map[reflect.Type]Declaration),
}

type declContext struct {
	mutex sync.RWMutex
	decls map[reflect.Type]Declaration
}

func DeclarationFor[Interface any]() (Declaration, bool) {
	return DeclarationOf(reflect.TypeFor[Interface]())
}

func DeclarationOf(t reflect.Type) (Declaration, bool) {
	globalDeclContext.mutex.RLock()
	defer globalDeclContext.mutex.RUnlock()

	v, ok := globalDeclContext.decls[t]

	return v, ok
}

// Declare specifies at runtime a sum type based on a (marker) interface type. The actual members are defined
// through a (anonymous) function type, whose callbacks parameters define the allowed member types. This declaration
// must ever occur once per Interface type and as early as possible, thus probably at package level.
// Afterward, the [Enumeration] can be used for an exhaustive type switch and the interface type can
// transparently be used with the included json encoder and decoder package.
//
// MatchFn must look like
//
//	func(func(TypA), func(TypB), func(TypC))
func Declare[Interface any, MatchFn any](opts ...Option) Enumeration[Interface, MatchFn] {
	ifaceT := reflect.TypeFor[Interface]()
	if ifaceT.Kind() != reflect.Interface {
		panic(fmt.Errorf("expected interface but got %v", ifaceT))
	}

	globalDeclContext.mutex.RLock()
	if _, ok := globalDeclContext.decls[ifaceT]; ok {
		globalDeclContext.mutex.RUnlock()
		panic(fmt.Errorf("interface already declared: %v", ifaceT))
	}
	globalDeclContext.mutex.RUnlock()

	if ifaceT.NumMethod() == 0 {
		panic(fmt.Errorf("interface has no methods"))
	}

	fnType := reflect.TypeFor[MatchFn]()
	if fnType.Kind() != reflect.Func {
		panic("type parameter MatchFn must be a function")
	}

	var enum Enumeration[Interface, MatchFn]
	enum.fnSwitchType = fnType
	enum.enumCfg.fromTypeToName = make(map[reflect.Type]string)
	enum.enumCfg.fromNameToType = make(map[string]reflect.Type)

	for _, opt := range opts {
		opt.apply(&enum.enumCfg)
	}

	if enum.jsonOpts == nil {
		Externally().apply(&enum.enumCfg)
	}

	for i := 0; i < fnType.NumIn(); i++ {
		argT := fnType.In(i) // e.g. something like func(enum_test.Dollar)
		if argT.Kind() != reflect.Func {
			panic(fmt.Errorf("branch must be of type func: %v", argT))
		}

		if argT.NumOut() != 0 {
			panic(fmt.Errorf("branch must not declare output parameter: %v", argT))
		}

		if argT.NumIn() != 1 {
			panic(fmt.Errorf("branch must declare exact one input parameter: %v", argT))
		}

		defaultCase := false
		enumVariantT := argT.In(0) // e.g. any
		if i == fnType.NumIn()-1 && !enum.NoZero() {
			defaultCase = true
			if enumVariantT != reflect.TypeFor[any]() {
				panic(fmt.Errorf("last branch must declare func(any) but found: %v", enumVariantT))
			}
		} else if isInvalidKind(enumVariantT) {
			panic(fmt.Errorf("branch input parameter is of invalid kind: %v", argT))
		}

		if !defaultCase {
			ifaceT := reflect.TypeFor[Interface]()
			if !enumVariantT.AssignableTo(ifaceT) {
				panic(fmt.Errorf("branch input parameter must be assignable to Interface: %v is not a %v", enumVariantT, ifaceT))
			}

			enum.variants = append(enum.variants, enumVariantT)
		}

		name := enumVariantT.Name()
		if _, ok := enum.fromNameToType[name]; !ok {
			enum.fromNameToType[name] = enumVariantT
			enum.fromTypeToName[enumVariantT] = name
		}

		//fmt.Println(enumVariantT)
	}

	globalDeclContext.mutex.Lock()
	defer globalDeclContext.mutex.Unlock()

	globalDeclContext.decls[ifaceT] = Declaration{
		ifaceT:   ifaceT,
		variants: enum.variants,
		cfg:      enum.enumCfg,
	}

	return enum
}

func isInvalidKind(vType reflect.Type) bool {
	// we disallow also pointers here, because it may cause multiple unwanted side effects
	//  - we want to encourage an immutable programming style
	//  - spreading pointers in (json) serialization breaks invariants by definition, because content is flattened
	//  - intention of using a pointer as an optional marker is not clear, use an optional type instead
	//  - causing additional indirections and escapes are unnecessary for optional cases
	return vType.Kind() == reflect.Ptr ||

		// channels are not serializable by definition
		vType.Kind() == reflect.Chan ||

		// functions are not serializable by definition
		vType.Kind() == reflect.Func
}
