// +build js,wasm

package main

import (
	"errors"
	"fmt"
	"sync"
	"syscall/js"

	"github.com/genjidb/genji"
	"github.com/genjidb/genji.js/src/bindings/simpleengine"
	"github.com/genjidb/genji/document"
)

func main() {
	wrapper := GenjiWrapper{
		dbs: make(map[int]*genji.DB),
	}

	js.Global().Set("openDB", js.FuncOf(runWithCallback(func(inputs []js.Value) (interface{}, error) {
		return wrapper.OpenDB()
	})))
	js.Global().Set("dbExec", js.FuncOf(runWithCallback(func(inputs []js.Value) (interface{}, error) {
		if len(inputs) < 3 {
			return nil, errors.New("missing arguments")
		}

		return nil, wrapper.Exec(inputs[0].Int(), inputs[1].String(), jsArrayToSlice(inputs[2]))
	})))
	js.Global().Set("dbQuery", js.FuncOf(func(this js.Value, inputs []js.Value) interface{} {
		if len(inputs) < 3 {
			return errors.New("missing arguments")
		}

		callback := inputs[len(inputs)-1:][0]
		inputs = inputs[:len(inputs)-1]

		err := wrapper.Query(inputs[0].Int(), func(m map[string]interface{}) error {
			callback.Invoke(nil, m)
			return nil
		}, inputs[1].String(), jsArrayToSlice(inputs[2]))
		if err != nil {
			callback.Invoke(err.Error(), nil)
			return nil
		}

		callback.Invoke(nil, nil)
		return nil
	}))

	select {}
}

// GenjiWrapper is a type that allows opening databases and running queries on them.
// It keeps a reference on all opened databases.
type GenjiWrapper struct {
	dbs map[int]*genji.DB

	lastid int
	m      sync.RWMutex
}

// OpenDB opens a database and returns an id that can be used by the Javascript code to select
// the right database when running a query.
func (w *GenjiWrapper) OpenDB() (int, error) {
	db, err := genji.New(simpleengine.NewEngine())
	if err != nil {
		return 0, err
	}

	w.m.Lock()
	defer w.m.Unlock()
	w.lastid++
	w.dbs[w.lastid] = db
	return w.lastid, nil
}

// Exec calls Genji's db.Exec method on the database identified by the id passed as first argument.
func (w *GenjiWrapper) Exec(id int, query string, args []js.Value) error {
	db, ok := w.dbs[id]
	if !ok {
		return fmt.Errorf("unknown database id %d", id)
	}

	params, err := jsValuesToParams(args)
	if err != nil {
		return err
	}

	return db.Exec(query, params...)
}

// Query calls Genji's db.Query method.
func (w *GenjiWrapper) Query(id int, cb func(m map[string]interface{}) error, query string, args []js.Value) error {
	db, ok := w.dbs[id]
	if !ok {
		return fmt.Errorf("unknown database id %d", id)
	}

	params, err := jsValuesToParams(args)
	if err != nil {
		return err
	}

	res, err := db.Query(query, params...)
	if err != nil {
		return err
	}
	defer res.Close()

	return res.Iterate(func(d document.Document) error {
		m := make(map[string]interface{})

		err := d.Iterate(func(f string, v document.Value) error {
			m[f] = v.V
			return nil
		})
		if err != nil {
			return err
		}

		return cb(m)
	})
}

func runWithCallback(fn func(inputs []js.Value) (interface{}, error)) func(this js.Value, inputs []js.Value) interface{} {
	return func(this js.Value, inputs []js.Value) interface{} {
		callback := inputs[len(inputs)-1:][0]
		ret, err := fn(inputs[:len(inputs)-1])
		if err != nil {
			callback.Invoke(err.Error(), nil)
		} else {
			callback.Invoke(nil, ret)
		}
		return nil
	}
}
