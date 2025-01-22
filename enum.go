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

var declaredEnumTypes = map[reflect.Type]enum{}
var mutex sync.RWMutex

type encodingType int

const (
	encodeExternally encodingType = iota
	encodeAdjacent
)

type enum struct {
	variants                       []reflect.Type // enumeration = declaration order
	variantTypeFromString          map[string]reflect.Type
	variantStringFromType          map[reflect.Type]string
	adjacentNameDiscriminator      string
	adjacentNameContent            string
	internallyTagNameDiscriminator string

	encoding encodingType
}

// Enumeration represents a union type declaration and enumerates the declared subtypes for a distinct interface type.
type Enumeration[Interface any, MatchFn any] struct {
	enumCfg
	variants     []reflect.Type
	fnSwitchType reflect.Type
}

func (b Enumeration[Interface, MatchFn]) Types() iter.Seq[reflect.Type] {
	return slices.Values(b.variants)
}

func (b Enumeration[Interface, MatchFn]) Switch(value Interface) MatchFn {
	fnImpl := reflect.MakeFunc(b.fnSwitchType, func(args []reflect.Value) []reflect.Value {
		// args are things like func(Euro), func(Dollar)
		ordinal := b.Ordinal(value)
		var fnArg reflect.Value
		if ordinal < 0 {
			fnArg = args[len(args)-1]
		} else {
			fnArg = args[ordinal]
		}

		if ordinal == -1 {
			fnArg.Call([]reflect.Value{reflect.New(reflect.TypeFor[any]()).Elem()})
		} else {
			fnArg.Call([]reflect.Value{reflect.ValueOf(value)})
		}

		return nil
	})

	return fnImpl.Interface().(MatchFn)

}

func (b Enumeration[Interface, MatchFn]) IsZero(value Interface) bool {
	var zeroIface Interface
	return any(value) == any(zeroIface)
}

// Ordinal returns either the zero based index of the declared types by MatchFn or -1. If the zero value
// has not been allowed in the declaration, it panics. If sealed and a non-exact type has been found, it panics and
// otherwise returns -1.
func (b Enumeration[Interface, MatchFn]) Ordinal(value Interface) int {
	isZero := b.IsZero(value)
	if b.NoZero() && isZero {
		panic(fmt.Errorf("enumeration is zero which is not allowed"))
	}

	if isZero {
		return -1
	}

	current := reflect.TypeOf(value)
	for i, variant := range b.variants {
		if variant == current {
			return i
		}
	}

	if b.Sealed() {
		panic(fmt.Errorf("enumeration is sealed and type has not been included in the enumeration: %T", value))
	}

	return -1
}
