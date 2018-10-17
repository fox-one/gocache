package redis

import (
	"testing"
	"time"

	"github.com/fox-one/gocache"
	"github.com/fox-one/gocache/tester"
	"github.com/gomodule/redigo/redis"
)

func getPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:   100,
		MaxActive: 300, // max number of connections
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "localhost:6379", redis.DialConnectTimeout(3*time.Second))
			if err != nil {
				panic(err)
			}
			return c, err
		},
	}
}

func TestRedis(t *testing.T) {
	for _, coder := range []gocache.Coder{gocache.Json, gocache.Jsoniter, gocache.Msgpack} {
		testWithCoder(t, coder)
	}
}

func testWithCoder(t *testing.T, coder gocache.Coder) {
	pool := getPool()
	defer pool.Close()

	store := NewStoreWithPool(pool)
	tester.Save(t, store, coder)
}
