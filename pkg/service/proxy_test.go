package service

import (
	"strings"
	"testing"

	"github.com/vikramsk/rediproxy/pkg/cache"
	"github.com/vikramsk/rediproxy/pkg/internal/mocks"
)

type mockCacher struct {
	*mocks.Getter
	*mocks.Setter
}

func cacheHit(key string) (string, error) {
	return strings.Replace(key, "key", "value", 1), nil
}

func cacheMiss(key string) (string, error) {
	return "", cache.ErrKeyNotFound
}

func cacheSet(key, value string) {
	// no op
}

func getBackingLRUMocks(
	backingGet func(string) (string, error),
	lruGet func(string) (string, error),
	lruSet func(string, string),
) (*mocks.Getter, *mockCacher) {
	mBacking := &mocks.Getter{
		GetFn: backingGet,
	}

	mLRU := &mockCacher{
		Getter: &mocks.Getter{
			GetFn: lruGet,
		},
		Setter: &mocks.Setter{
			SetFn: lruSet,
		},
	}
	return mBacking, mLRU
}

func TestInMemoryCacheHit(t *testing.T) {
	mBacking, mLRU := getBackingLRUMocks(cacheHit, cacheHit, cacheSet)
	pc := NewCacheProxy(mBacking, mLRU)
	_, err := pc.Get("key")
	if err != nil || mBacking.GetFnInvoked {
		t.Fatalf("expected the data to be returned from the lru cache itself.")
	}
}

func TestInMemoryCacheMiss_BackingHit(t *testing.T) {
	mBacking, mLRU := getBackingLRUMocks(cacheHit, cacheMiss, cacheSet)
	pc := NewCacheProxy(mBacking, mLRU)
	_, err := pc.Get("key")
	if err != nil || !mBacking.GetFnInvoked || !mLRU.SetFnInvoked {
		t.Fatalf("expected the data to be returned from the backing store and then set in the lru cache")
	}
}

func TestInMemoryCacheMiss_BackingMiss(t *testing.T) {
	mBacking, mLRU := getBackingLRUMocks(cacheMiss, cacheMiss, cacheSet)

	pc := NewCacheProxy(mBacking, mLRU)
	_, err := pc.Get("key")
	if err == nil || !mBacking.GetFnInvoked || mLRU.SetFnInvoked {
		t.Fatalf("expected the function to return after failing to retrieve data from backing store")
	}
}
