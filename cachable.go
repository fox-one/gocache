package gocache

type Cachable interface {
	CacheKey() (string, error)
}

type cacheCoder interface {
	CacheCoder() Coder
}

func (c *Cache) cacheCoder(v Cachable) Coder {
	if coder, ok := v.(cacheCoder); ok {
		return coder.CacheCoder()
	}

	return c.coder
}

type cacheSubkeyer interface {
	CacheSubkeys() []string
}

func (c *Cache) cacheSubkeys(v Cachable) []string {
	if sub, ok := v.(cacheSubkeyer); ok {
		return sub.CacheSubkeys()
	}

	return nil
}

type cacheExpirer interface {
	CacheExpire() int64
}

func (c *Cache) cacheExpire(v Cachable, expires ...int64) int64 {
	if len(expires) > 0 {
		return expires[0]
	}

	if ex, ok := v.(cacheExpirer); ok {
		return ex.CacheExpire()
	}

	return 0
}
