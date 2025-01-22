// Copyright 2024 Torben Schinke. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

import (
	"fmt"
	"github.com/worldiety/enum"
	"reflect"
)

type codecMode int

const (
	done codecMode = iota
	fallback
)

func interfaceEncoder(e *encodeState, v reflect.Value, opts encOpts) {
	decl, ok := enum.DeclarationOf(v.Type())
	if !ok {
		//e.error(fmt.Errorf("json: undeclared interface type %v", v.Type()))
		// fallback to default stdlib behavior
		if v.IsNil() {
			e.WriteString("null")
			return
		}
		e.reflectValue(v.Elem(), opts)

		return
	}

	if v.IsNil() {
		if decl.NoZero() {
			e.error(fmt.Errorf("json: nil is not allowed for interface %v", v.Type()))
		}

		e.WriteString("null")
		return
	}

	switch jte := decl.JSON().(type) {
	case enum.ExternallyOptions:
		encodeExternally(e, v, opts, decl, jte)
	case enum.AdjacentlyOptions:
		encodeAdjacently(e, v, opts, decl, jte)
	case enum.InternallyOptions:
		encodeInternally(e, v, opts, decl, jte)
	case enum.UntaggedOptions:
		// this is one-way and like the original implementation
		e.reflectValue(v, opts)
	default:
		e.error(fmt.Errorf("json: unsupported JSON option type %T", jte))
	}

}

func encodeExternally(e *encodeState, v reflect.Value, opts encOpts, decl enum.Declaration, jsonOpts enum.ExternallyOptions) {
	e.WriteByte('{')
	externalName, ok := decl.Name(v.Elem().Type())
	if !ok {
		e.error(fmt.Errorf("json: undeclared external type name for interface variant type '%T'.'%v'", v.Type(), v.Elem().Type()))
	}

	e.Write(appendString(e.AvailableBuffer(), externalName, opts.quoted))
	e.WriteByte(':')
	e.reflectValue(v.Elem(), opts)
	e.WriteByte('}')
}

func encodeAdjacently(e *encodeState, v reflect.Value, opts encOpts, decl enum.Declaration, jsonOpts enum.AdjacentlyOptions) {
	e.WriteByte('{')
	e.Write(appendString(e.AvailableBuffer(), jsonOpts.Tag, opts.quoted))
	e.WriteByte(':')

	externalName, ok := decl.Name(v.Elem().Type())
	if !ok {
		e.error(fmt.Errorf("json: undeclared external type name for interface variant type '%T'.'%v'", v.Type(), v.Elem().Type()))
	}

	e.Write(appendString(e.AvailableBuffer(), externalName, opts.quoted))

	e.WriteByte(',')

	e.Write(appendString(e.AvailableBuffer(), jsonOpts.Content, opts.quoted))
	e.WriteByte(':')
	e.reflectValue(v.Elem(), opts)
	e.WriteByte('}')
}

func encodeInternally(e *encodeState, v reflect.Value, opts encOpts, decl enum.Declaration, jsonOpts enum.InternallyOptions) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fallthrough
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fallthrough
	case reflect.Float32, reflect.Float64:
		fallthrough
	case reflect.String, reflect.Bool:
		e.error(fmt.Errorf("cannot encode non-object kind value as internally tagged: %v", v))
	default:
		// nothing
	}

	// write the normal object
	e.reflectValue(v.Elem(), opts)

	// step back
	e.Truncate(e.Len() - 1)

	// append the type tag
	e.WriteByte(',')
	e.Write(appendString(e.AvailableBuffer(), jsonOpts.Tag, opts.quoted))
	e.WriteByte(':')

	externalName, ok := decl.Name(v.Elem().Type())
	if !ok {
		e.error(fmt.Errorf("json: undeclared external type name for interface variant type '%T'.'%v'", v.Type(), v.Elem().Type()))
	}

	e.Write(appendString(e.AvailableBuffer(), externalName, opts.quoted))

	e.WriteByte('}')
}
