package memory

import (
	"time"

	"github.com/fox-one/gocache"
	"github.com/patrickmn/go-cache"
)

type store struct {
	cache *cache.Cache
}

func NewStore() gocache.Store {
	return &store{
		cache: cache.New(cache.NoExpiration, time.Minute*10),
	}
}

func (s *store) Save(pairs gocache.Pairs, expire int64) error {
	d := cache.NoExpiration
	if expire > 0 {
		d = time.Duration(expire) * time.Second
	}

	for k, data := range pairs {
		s.cache.Set(k, data, d)
	}

	return nil
}

func (s *store) Get(keys ...string) (gocache.Pairs, error) {
	pairs := make(gocache.Pairs)
	for _, k := range keys {
		if data, ok := s.cache.Get(k); ok {
			pairs.Set(k, data.([]byte))
		}
	}

	return pairs, nil
}

func (s *store) Delete(keys ...string) error {
	for _, key := range keys {
		s.cache.Delete(key)
	}

	return nil
}

func (s *store) Exists(key string) (bool, error) {
	_, ok := s.cache.Get(key)
	return ok, nil
}
