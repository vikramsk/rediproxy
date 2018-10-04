package service

import (
	"github.com/vikramsk/rediproxy/pkg/cache"
)

type cacheProxy struct {
	lruCache      cache.Cacher
	backingClient cache.Getter
}

// NewCacheProxy initializes the primary cache proxy service.
// It accepts the interfaces for the backing cache store and
// the in memory cache.
func NewCacheProxy(c cache.Getter, lc cache.Cacher) cache.Getter {
	return &cacheProxy{
		backingClient: c,
		lruCache:      lc,
	}
}

// Get returns the value for a given key.
// It looks for the key in the in-memory cache. If it
// doesn't find it there, it fetches the data from
// the backing cache store.
func (cp *cacheProxy) Get(key string) (string, error) {
	// lookup key in the in-memory cache.
	val, err := cp.lruCache.Get(key)
	if err == nil {
		return val, nil
	}

	// lookup key in the backing store.
	val, err = cp.backingClient.Get(key)
	if err != nil {
		return "", err
	}

	// add key to in-memory cache
	cp.lruCache.Set(key, val)
	return val, nil
}
