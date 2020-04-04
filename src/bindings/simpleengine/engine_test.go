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

func TestMemoryEngine(t *testing.T) {
	enginetest.TestSuite(t, builder)
}

func BenchmarkMemoryEngineStorePut(b *testing.B) {
	enginetest.BenchmarkStorePut(b, builder)
}

func BenchmarkMemoryEngineStoreScan(b *testing.B) {
	enginetest.BenchmarkStoreScan(b, builder)
}
