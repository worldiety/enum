// Copyright 2024 Torben Schinke. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

import (
	"fmt"
	"github.com/worldiety/enum"
	"reflect"
)

func interfaceDecoder(d *decodeState, v reflect.Value) error {
	// we have a huge problem here, because we cannot simply stream parse the tokens,
	// because the payload may be declared BEFORE any tag occurs (at least for adjacent and internally tagged).
	// Though, externally tagged can be optimized but they are not exact either,
	// because other tags may occur as a sibling.
	tmp := map[string]RawMessage{}
	tmpV := reflect.ValueOf(tmp)
	if err := d.object(tmpV); err != nil {
		return err
	}

	t := v.Type()
	decl, ok := enum.DeclarationOf(v.Type())
	if !ok {
		d.saveError(&UnmarshalTypeError{Value: "interface", Type: t, Offset: int64(d.off)})
		d.skip()
		return nil
	}

	switch jte := decl.JSON().(type) {
	case enum.ExternallyOptions:
		return decodeExternally(d, tmp, v, decl, jte)
	case enum.AdjacentlyOptions:
		return decodeAdjacently(d, tmp, v, decl, jte)
	case enum.UntaggedOptions:
		d.saveError(&UnmarshalTypeError{Value: "untagged interfaces cannot be unmarshalled", Type: t, Offset: int64(d.off)})
		d.skip()
		return nil

	default:
		d.saveError(&UnmarshalTypeError{Value: fmt.Sprintf("unknown json tag options: %T", jte), Type: t, Offset: int64(d.off)})
		d.skip()
		return nil
	}

}

func decodeAdjacently(d *decodeState, obj map[string]RawMessage, v reflect.Value, decl enum.Declaration, jte enum.AdjacentlyOptions) error {
	kindTag, ok := unquote(obj[jte.Tag])
	if !ok {
		return fmt.Errorf("cannot unquote json tag '%s'", jte.Tag)
	}
	variantT, ok := decl.Type(kindTag)
	if !ok {
		return fmt.Errorf("unknown type tag '%v' for declaration of '%v'", kindTag, decl.EnumType())
	}

	targetVar := reflect.New(variantT).Interface()
	if err := Unmarshal(obj[jte.Content], &targetVar); err != nil {
		return err
	}

	v.Set(reflect.ValueOf(targetVar).Elem())

	return nil
}

func decodeExternally(d *decodeState, obj map[string]RawMessage, v reflect.Value, decl enum.Declaration, jte enum.ExternallyOptions) error {
	var firstKey string
	var data RawMessage
	for k, v := range obj {
		if data != nil {
			return fmt.Errorf("invalid externally tagged object format: found extra key: %v vs %v", firstKey, v)
		}

		firstKey = k
		data = v
	}

	kindTag := firstKey
	variantT, ok := decl.Type(kindTag)
	if !ok {
		return fmt.Errorf("unknown type tag '%v' for declaration of '%v'", kindTag, decl.EnumType())
	}

	targetVar := reflect.New(variantT).Interface()
	if err := Unmarshal(data, &targetVar); err != nil {
		return err
	}

	v.Set(reflect.ValueOf(targetVar).Elem())

	return nil
}
