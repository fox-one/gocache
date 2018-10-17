package memcache

import (
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/fox-one/gocache"
	"github.com/fox-one/gocache/tester"
)

func getClient() *memcache.Client {
	return memcache.New("localhost:11211")
}

func TestMemcached(t *testing.T) {
	for _, coder := range []gocache.Coder{gocache.Json, gocache.Jsoniter, gocache.Msgpack} {
		testWithCoder(t, coder)
	}
}

func testWithCoder(t *testing.T, coder gocache.Coder) {
	client := getClient()
	store := NewStoreWithClient(client)
	tester.Save(t, store, coder)
}
