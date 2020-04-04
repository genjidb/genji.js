// +build js,wasm

package main

import (
	"syscall/js"

	"github.com/asdine/genji"
	"github.com/asdine/genji/document"
)

//go:generate cp $GOROOT/misc/wasm/wasm_exec.js .

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
	db, err := genji.New(nil)
	if err != nil {
		return nil, err
	}

	id++
	openedDBs[id] = db
	return id, nil
}

func dbExec(inputs []js.Value) (interface{}, error) {
	db := openedDBs[inputs[0].Int()]
	return nil, db.Exec(inputs[1].String())
}

func dbQuery(this js.Value, inputs []js.Value) interface{} {
	callback := inputs[len(inputs)-1:][0]

	db := openedDBs[inputs[0].Int()]
	res, err := db.Query(inputs[1].String())
	if err != nil {
		callback.Invoke(nil, err)
		return nil
	}

	defer res.Close()

	err = res.Iterate(func(d document.Document) error {
		var m map[string]interface{}

		err := document.MapScan(d, &m)
		if err != nil {
			return err
		}

		callback.Invoke(m, nil)
		return nil
	})
	if err != nil {
		callback.Invoke(nil, err)
	}
	return nil
}
