// +build js,wasm

package main

import (
	"fmt"
	"syscall/js"
)

func jsValuesToParams(values []js.Value) ([]interface{}, error) {
	params := make([]interface{}, 0, len(values))
	for _, value := range values {
		var v interface{}

		switch value.Type() {
		case js.TypeBoolean:
			v = value.Bool()
		case js.TypeString:
			v = value.String()
		case js.TypeNumber:
			v = value.Float()
		default:
			return nil, fmt.Errorf("incompatible value %v", value)
		}

		params = append(params, v)
	}

	return params, nil
}

func jsArrayToSlice(a js.Value) []js.Value {
	values := make([]js.Value, a.Length())

	for i := 0; i < a.Length(); i++ {
		values[i] = a.Index(i)
	}

	return values
}
