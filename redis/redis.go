package redis

import (
	"github.com/fox-one/gocache"
	"github.com/gomodule/redigo/redis"
)

type store struct {
	pool *redis.Pool
}

func NewStoreWithPool(pool *redis.Pool) gocache.Store {
	if pool == nil {
		panic("pool is nil")
	}

	return &store{pool: pool}
}

func (s *store) Save(pairs gocache.Pairs, expire int64) error {
	if len(pairs) == 0 {
		return nil
	}

	coon := s.pool.Get()
	defer coon.Close()

	if expire <= 0 {
		_, err := coon.Do("MSET", pairs.Spread()...)
		return err
	}

	if len(pairs) == 1 {
		items := pairs.Spread()
		_, err := coon.Do("SETEX", items[0], expire, items[1])
		return err
	}

	if err := coon.Send("MULTI"); err != nil {
		return err
	}

	for k, v := range pairs {
		if err := coon.Send("SETEX", k, expire, v); err != nil {
			return err
		}
	}

	_, err := coon.Do("EXEC")
	return err
}

func (s *store) Get(keys ...string) (gocache.Pairs, error) {
	params := make([]interface{}, len(keys))
	for idx, key := range keys {
		params[idx] = key
	}

	if len(params) == 0 {
		return nil, nil
	}

	coon := s.pool.Get()
	defer coon.Close()

	data, err := redis.ByteSlices(coon.Do("MGET", params...))
	if err != nil {
		return nil, err
	}

	pairs := gocache.Pairs{}
	for idx, key := range keys {
		if v := data[idx]; len(v) > 0 {
			pairs.Set(key, data[idx])
		}
	}

	return pairs, nil
}

func (s *store) Delete(keys ...string) error {
	params := make([]interface{}, len(keys))
	for idx, key := range keys {
		params[idx] = key
	}

	if len(params) == 0 {
		return nil
	}

	coon := s.pool.Get()
	defer coon.Close()

	_, err := coon.Do("DEL", params...)
	return err
}

func (s *store) Exists(key string) (bool, error) {
	coon := s.pool.Get()
	defer coon.Close()

	return redis.Bool(coon.Do("EXISTS", key))
}
