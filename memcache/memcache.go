package memcache

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/fox-one/gocache"
)

type store struct {
	client *memcache.Client
}

func NewStoreWithClient(client *memcache.Client) gocache.Store {
	return &store{client: client}
}

func (s *store) Save(pairs gocache.Pairs, expire int64) error {
	for k, v := range pairs {
		it := &memcache.Item{
			Key:        k,
			Value:      v,
			Expiration: int32(expire),
		}

		if err := s.client.Set(it); err != nil {
			return err
		}
	}

	return nil
}

func (s *store) Get(keys ...string) (gocache.Pairs, error) {
	items, err := s.client.GetMulti(keys)
	if err != nil {
		return nil, err
	}

	pairs := make(gocache.Pairs, len(items))
	for k, item := range items {
		pairs.Set(k, item.Value)
	}

	return pairs, nil
}

func (s *store) Delete(keys ...string) error {
	for _, key := range keys {
		err := s.client.Delete(key)
		if err != nil && err != memcache.ErrCacheMiss {
			return err
		}
	}

	return nil
}

func (s *store) Exists(key string) (bool, error) {
	item, err := s.client.Get(key)
	if err == memcache.ErrCacheMiss {
		err = nil
	}

	return item != nil, err
}
