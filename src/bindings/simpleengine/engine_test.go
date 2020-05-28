package simpleengine_test

import (
	"testing"

	"github.com/genjidb/genji.js/src/bindings/simpleengine"
	"github.com/genjidb/genji/engine"
	"github.com/genjidb/genji/engine/enginetest"
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
