package memory

import (
	"testing"

	"github.com/fox-one/gocache"
	"github.com/fox-one/gocache/tester"
)

func TestMemory(t *testing.T) {
	for _, coder := range []gocache.Coder{gocache.Json, gocache.Jsoniter, gocache.Msgpack} {
		testWithCoder(t, coder)
	}
}

func testWithCoder(t *testing.T, coder gocache.Coder) {
	store := NewStore()
	tester.Save(t, store, coder)
}
