package gocache

import (
	"fmt"

	"github.com/pkg/errors"
)

type Cachable interface {
	CacheKey() (string, error)
}

type CacheExpire interface {
	CacheExpire() int64
}

type CacheCoder interface {
	CacheCoder() Coder
}

type CacheSubkeys interface {
	CacheSubkeys() []string
}

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

func (c *Cache) cacheCoder(x Cachable) Coder {
	if coder, ok := x.(CacheCoder); ok {
		return coder.CacheCoder()
	}

	return c.coder
}

func (c *Cache) cacheSubkeys(x Cachable) []string {
	if sub, ok := x.(CacheSubkeys); ok {
		return sub.CacheSubkeys()
	}

	return nil
}

func (c *Cache) cacheExpire(x Cachable, expires ...int64) int64 {
	if len(expires) > 0 {
		return expires[0]
	}

	if ex, ok := x.(CacheExpire); ok {
		return ex.CacheExpire()
	}

	return 0
}

func (c *Cache) Save(x Cachable, expires ...int64) error {
	key, err := x.CacheKey()
	if err != nil {
		return errors.Wrap(err, "primary key is invalid")
	}

	coder := c.cacheCoder(x)
	data, err := coder.Marshal(x)
	if err != nil {
		return errors.Wrap(err, "marshal failed")
	}

	pairs := Pairs{key: data}
	if subkeys := c.cacheSubkeys(x); len(subkeys) > 0 {
		keyData := []byte(key)
		for _, subkey := range subkeys {
			pairs.Set(subkey, keyData)
		}
	}

	if err := c.store.Save(pairs, c.cacheExpire(x, expires...)); err != nil {
		return errors.Wrap(err, "save data failed")
	}

	return nil
}

func getPrimaryKey(pairs Pairs) string {
	for _, v := range pairs {
		if len(v) > 0 {
			return string(v)
		}
	}

	return ""
}

func (c *Cache) Load(x Cachable) error {
	key, err := x.CacheKey()
	if err != nil {
		subkeys := c.cacheSubkeys(x)
		if len(subkeys) == 0 {
			return fmt.Errorf("key is not available neither primary key nor subkeys")
		}

		pairs, err := c.store.Get(subkeys...)
		if err != nil {
			return errors.Wrap(err, "load key from subkeys failed")
		}

		if key = getPrimaryKey(pairs); len(key) == 0 {
			return CacheMiss
		}
	}

	pairs, err := c.store.Get(key)
	if err != nil {
		return errors.Wrap(err, "load data by primary key failed")
	}

	if data, ok := pairs.Get(key); ok {
		coder := c.cacheCoder(x)
		if err := coder.Unmarshal(data, x); err != nil {
			return errors.Wrap(err, "unmarshal data failed")
		}

		return nil
	}

	return CacheMiss
}

func (c *Cache) Clean(x Cachable) error {
	key, err := x.CacheKey()
	if err != nil {
		return errors.Wrap(err, "primary key is invalid")
	}

	keys := append(c.cacheSubkeys(x), key)
	return c.store.Delete(keys...)
}

// errors
var CacheMiss = fmt.Errorf("cache miss")
