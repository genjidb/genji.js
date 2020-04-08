// +build js,wasm

package main

import (
	"syscall/js"

	"github.com/asdine/genji"
	"github.com/asdine/genji.js/src/bindings/simpleengine"
	"github.com/asdine/genji/document"
)

func main() {
	js.Global().Set("runDB", js.FuncOf(runWithCallback(runDB)))
	js.Global().Set("dbExec", js.FuncOf(runWithCallback(dbExec)))
	js.Global().Set("dbQuery", js.FuncOf(dbQuery))

	select {}
}

var openedDBs = map[int]*genji.DB{}

var id int

func runWithCallback(fn func(inputs []js.Value) (interface{}, error)) func(this js.Value, inputs []js.Value) interface{} {
	return func(this js.Value, inputs []js.Value) interface{} {
		callback := inputs[len(inputs)-1:][0]
		ret, err := fn(inputs[:len(inputs)-1])
		callback.Invoke(err, ret)
		return nil
	}
}

func runDB(inputs []js.Value) (interface{}, error) {
	db, err := genji.New(simpleengine.NewEngine())
	if err != nil {
		return nil, err
	}

	id++
	openedDBs[id] = db
	return id, nil
}

func dbExec(inputs []js.Value) (interface{}, error) {
	db := openedDBs[inputs[0].Int()]

	var args []interface{}
	for _, arg := range inputs[2:] {
		var v interface{}

		switch arg.Type() {
		case js.TypeBoolean:
			v = arg.Bool()
		case js.TypeString:
			v = arg.String()
		case js.TypeNumber:
			v = arg.Float()
		}
		args = append(args, v)
	}
	return nil, db.Exec(inputs[1].String(), args...)
}

func dbQuery(this js.Value, inputs []js.Value) interface{} {
	callback := inputs[len(inputs)-1:][0]
	db := openedDBs[inputs[0].Int()]

	var args []interface{}
	for _, arg := range inputs[2:] {
		var v interface{}

		switch arg.Type() {
		case js.TypeBoolean:
			v = arg.Bool()
		case js.TypeString:
			v = arg.String()
		case js.TypeNumber:
			v = arg.Float()
		}
		args = append(args, v)
	}

	res, err := db.Query(inputs[1].String(), args...)
	if err != nil {
		callback.Invoke(err, nil)
		return nil
	}

	defer res.Close()

	err = res.Iterate(func(d document.Document) error {
		m := make(map[string]interface{})

		err := d.Iterate(func(f string, v document.Value) error {
			m[f] = v.V
			return nil
		})
		if err != nil {
			return err
		}

		callback.Invoke(nil, m)
		return nil
	})
	if err != nil {
		callback.Invoke(err, nil)
		return nil
	}

	callback.Invoke(nil, nil)
	return nil
}
