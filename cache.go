package gocache

import (
	"errors"
)

// Coder coder
type Coder interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

// Pairs store pairs
type Pairs map[string][]byte

// IsEmpty is empty
func (pairs Pairs) IsEmpty() bool {
	return len(pairs) == 0
}

// Set set
func (pairs Pairs) Set(k string, v []byte) {
	pairs[k] = v
}

// Get get
func (pairs Pairs) Get(k string) ([]byte, bool) {
	v, ok := pairs[k]
	return v, ok
}

// Spread spread
func (pairs Pairs) Spread() []interface{} {
	args := make([]interface{}, 0, len(pairs)*2)
	for k, v := range pairs {
		args = append(args, k, v)
	}

	return args
}

// Store store interface
type Store interface {
	Save(pairs Pairs, expire int64) error
	Get(keys ...string) (Pairs, error)
	Delete(keys ...string) error
	Exists(key string) (bool, error)
}

// ErrCacheMiss missed
var ErrCacheMiss = errors.New("cache miss")

// Cache cache
type Cache struct {
	store Store
	// default coder used when no coder is specified
	coder Coder
}

// New new cache
func New(store Store, defaultCoder Coder) *Cache {
	return &Cache{
		store: store,
		coder: defaultCoder,
	}
}

// MultiWrite write multiple items
func (c *Cache) MultiWrite(items map[string]interface{}, expires ...int64) error {
	pairs := make(Pairs, len(items))
	for key, v := range items {
		data, err := c.coder.Marshal(v)
		if err != nil {
			return err
		}
		pairs[key] = data
	}

	var exoire int64
	if len(expires) > 0 {
		exoire = expires[0]
	}

	return c.store.Save(pairs, exoire)
}

// MultiRead read multiple items
// TODO bugfix
//  Multi Read ALL will always return empty
func (c *Cache) MultiRead(newItemFunc func() interface{}, keys ...string) ([]interface{}, error) {
	pairs, err := c.store.Get(keys...)
	if err != nil {
		return nil, err
	}

	arr := make([]interface{}, 0, len(keys))
	for _, data := range pairs {
		item := newItemFunc()
		if err := c.coder.Unmarshal(data, item); err == nil {
			arr = append(arr, item)
		}
	}

	return arr, err
}

// Write write key value
func (c *Cache) Write(key string, v interface{}, expires ...int64) error {
	data, err := c.coder.Marshal(v)
	if err != nil {
		return err
	}

	var exoire int64
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

// Delete delete keys...
func (c *Cache) Delete(keys ...string) error {
	return c.store.Delete(keys...)
}

// Save save item
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

// Load load item
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

// Clean clean items
func (c *Cache) Clean(vs ...Cachable) error {
	var keys = make([]string, 0, len(vs))
	for _, v := range vs {
		key, err := v.CacheKey()
		if err != nil {
			return err
		}

		keys = append(keys, key)
		keys = append(keys, c.cacheSubkeys(v)...)

	}
	return c.store.Delete(keys...)
}
