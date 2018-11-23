package gocache

import (
	"errors"
)

type Coder interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type Pairs map[string][]byte

func (pairs Pairs) IsEmpty() bool {
	return len(pairs) == 0
}

func (pairs Pairs) Set(k string, v []byte) {
	pairs[k] = v
}

func (pairs Pairs) Get(k string) ([]byte, bool) {
	v, ok := pairs[k]
	return v, ok
}

func (pairs Pairs) Spread() []interface{} {
	args := make([]interface{}, 0, len(pairs)*2)
	for k, v := range pairs {
		args = append(args, k, v)
	}

	return args
}

type Store interface {
	Save(pairs Pairs, expire int64) error
	Get(keys ...string) (Pairs, error)
	Delete(keys ...string) error
	Exists(key string) (bool, error)
}

// errors
var ErrCacheMiss = errors.New("cache miss")

type Cache struct {
	store Store
	// default coder used when no coder is specified
	coder Coder
}

func New(store Store, defaultCoder Coder) *Cache {
	return &Cache{
		store: store,
		coder: defaultCoder,
	}
}

func (c *Cache) Write(key string, v interface{}, expires ...int64) error {
	data, err := c.coder.Marshal(v)
	if err != nil {
		return err
	}

	var exoire int64 = 0
	if len(expires) > 0 {
		exoire = expires[0]
	}

	return c.store.Save(Pairs{key: data}, exoire)
}

func (c *Cache) Read(key string, v interface{}) error {
	pairs, err := c.store.Get(key)
	if err != nil {
		return err
	}

	data, ok := pairs.Get(key)
	if !ok {
		return ErrCacheMiss
	}

	err = c.coder.Unmarshal(data, v)
	return err
}

func (c *Cache) Delete(key string) error {
	return c.store.Delete(key)
}

func (c *Cache) Save(v Cachable, expires ...int64) error {
	key, err := v.CacheKey()
	if err != nil {
		return err
	}

	coder := c.cacheCoder(v)
	data, err := coder.Marshal(v)
	if err != nil {
		return err
	}

	pairs := Pairs{key: data}
	if subkeys := c.cacheSubkeys(v); len(subkeys) > 0 {
		keyData := []byte(key)
		for _, subkey := range subkeys {
			pairs.Set(subkey, keyData)
		}
	}

	if err := c.store.Save(pairs, c.cacheExpire(v, expires...)); err != nil {
		return err
	}

	return nil
}

func dumpPrimaryKey(pairs Pairs) string {
	for _, v := range pairs {
		if len(v) > 0 {
			return string(v)
		}
	}

	return ""
}

func (c *Cache) Load(v Cachable) error {
	key, err := v.CacheKey()
	if err != nil {
		subkeys := c.cacheSubkeys(v)
		if len(subkeys) == 0 {
			return errors.New("key is not available neither primary key nor subkeys")
		}

		pairs, err := c.store.Get(subkeys...)
		if err != nil {
			return err
		}

		if key = dumpPrimaryKey(pairs); len(key) == 0 {
			return ErrCacheMiss
		}
	}

	pairs, err := c.store.Get(key)
	if err != nil {
		return err
	}

	if data, ok := pairs.Get(key); ok {
		coder := c.cacheCoder(v)
		if err := coder.Unmarshal(data, v); err != nil {
			return err
		}

		return nil
	}

	return ErrCacheMiss
}

func (c *Cache) Clean(v Cachable) error {
	key, err := v.CacheKey()
	if err != nil {
		return err
	}

	keys := append(c.cacheSubkeys(v), key)
	return c.store.Delete(keys...)
}
