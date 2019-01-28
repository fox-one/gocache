package redis

import (
	"github.com/fox-one/gocache"
	"github.com/gomodule/redigo/redis"
)

type hashStore struct {
	pool *redis.Pool
	key  string
}

// NewHashStore new redis hash store
// TODO bugfix
//  Multi Read ALL will always return empty
func NewHashStore(pool *redis.Pool, key string) gocache.Store {
	if pool == nil {
		panic("pool is nil")
	}

	return &hashStore{
		pool: pool,
		key:  key,
	}
}

// Save save items
func (s *hashStore) Save(pairs gocache.Pairs, expire int64) error {
	if len(pairs) == 0 {
		return nil
	}

	conn := s.pool.Get()
	defer conn.Close()

	paras := append([]interface{}{s.key}, pairs.Spread()...)

	_, err := conn.Do("HMSET", paras...)
	if err != nil {
		return err
	}

	if expire > 0 {
		_, err = conn.Do("EXPIRE", s.key, expire)
	}
	return err
}

func (s *hashStore) getAll() (gocache.Pairs, error) {
	conn := s.pool.Get()
	defer conn.Close()

	data, err := redis.ByteSlices(conn.Do("HGETALL", s.key))
	if err != nil {
		return nil, err
	}

	pairs := gocache.Pairs{}
	for idx := 0; idx < len(data); idx += 2 {
		pairs.Set(string(data[idx]), data[idx+1])
	}
	return pairs, nil
}

// Get get items
func (s *hashStore) Get(fields ...string) (gocache.Pairs, error) {
	if len(fields) == 0 {
		return s.getAll()
	}

	conn := s.pool.Get()
	defer conn.Close()

	params := make([]interface{}, len(fields)+1)
	params[0] = s.key
	for idx, field := range fields {
		params[idx+1] = field
	}

	data, err := redis.ByteSlices(conn.Do("HMGET", params...))
	if err != nil {
		return nil, err
	}

	pairs := gocache.Pairs{}
	for idx, field := range fields {
		if v := data[idx]; len(v) > 0 {
			pairs.Set(field, data[idx])
		}
	}

	return pairs, nil
}

// Delete delete items
func (s *hashStore) Delete(fields ...string) error {
	coon := s.pool.Get()
	defer coon.Close()

	if len(fields) == 0 {
		_, err := coon.Do("DEL", s.key)
		return err
	}

	params := make([]interface{}, len(fields)+1)
	params[0] = s.key
	for idx, field := range fields {
		params[idx+1] = field
	}

	_, err := coon.Do("HDEL", params...)
	return err
}

// Exists exists
func (s *hashStore) Exists(field string) (bool, error) {
	coon := s.pool.Get()
	defer coon.Close()

	return redis.Bool(coon.Do("HEXISTS", s.key, field))
}
