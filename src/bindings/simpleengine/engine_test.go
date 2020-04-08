package simpleengine_test

import (
	"testing"

	"github.com/asdine/genji.js/src/bindings/simpleengine"
	"github.com/asdine/genji/engine"
	"github.com/asdine/genji/engine/enginetest"
)

func builder() (engine.Engine, func()) {
	ng := simpleengine.NewEngine()
	return ng, func() { ng.Close() }
}

func Testsimpleengine(t *testing.T) {
	enginetest.TestSuite(t, builder)
}

func BenchmarksimpleengineStorePut(b *testing.B) {
	enginetest.BenchmarkStorePut(b, builder)
}

func BenchmarksimpleengineStoreScan(b *testing.B) {
	enginetest.BenchmarkStoreScan(b, builder)
}
