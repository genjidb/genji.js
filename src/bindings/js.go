// +build js,wasm

package main

import (
	"errors"
	"syscall/js"

	"github.com/genjidb/genji/document"
)

func jsValuesToParams(values []js.Value) ([]interface{}, error) {
	params := make([]interface{}, len(values))
	for i, value := range values {
		vv, err := jsValueToParam(value)
		if err != nil {
			return nil, err
		}
		params[i] = vv
	}

	return params, nil
}

func jsValueToParam(value js.Value) (interface{}, error) {
	var v interface{}

	switch value.Type() {
	case js.TypeBoolean:
		v = value.Bool()
	case js.TypeString:
		v = value.String()
	case js.TypeNumber:
		v = value.Float()
	case js.TypeObject:
		keys := value.Get("_keys")
		// array
		if keys.IsUndefined() {
			vb := document.NewValueBuffer()
			for i := 0; i < value.Length(); i++ {
				p, err := jsValueToParam(value.Index(i))
				if err != nil {
					return nil, err
				}
				vv, err := document.NewValue(p)
				if err != nil {
					return nil, err
				}

				vb = vb.Append(vv)
			}

			return vb, nil
		}

		// object
		object := value.Get("_object")
		fb := document.NewFieldBuffer()
		for i := 0; i < keys.Length(); i++ {
			k := keys.Index(i).String()

			p, err := jsValueToParam(object.Get(k))
			if err != nil {
				return nil, err
			}

			vv, err := document.NewValue(p)
			if err != nil {
				return nil, err
			}

			fb.Add(k, vv)
		}
		v = fb
	default:
		return nil, errors.New("incompatible value " + value.String())
	}

	return v, nil
}

func jsArrayToSlice(a js.Value) []js.Value {
	values := make([]js.Value, a.Length())

	for i := 0; i < a.Length(); i++ {
		values[i] = a.Index(i)
	}

	return values
}

func genjiValuesToJs(v document.Value) (interface{}, error) {
	switch v.Type {
	case document.ArrayValue:
		var arr []interface{}
		err := v.V.(document.Array).Iterate(func(i int, value document.Value) error {
			vv, err := genjiValuesToJs(value)
			if err != nil {
				return err
			}
			arr = append(arr, vv)
			return nil
		})
		return arr, err
	case document.DocumentValue:
		m := make(map[string]interface{})
		err := v.V.(document.Document).Iterate(func(field string, value document.Value) error {
			var err error
			m[field], err = genjiValuesToJs(value)
			return err
		})
		return m, err
	}

	return v.V, nil
}
