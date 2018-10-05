package cache

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func key(i int) string {
	return fmt.Sprintf("key%d", i)
}

func value(i int) string {
	return fmt.Sprintf("value%d", i)
}

func TestGetOldest_EvictOldest(t *testing.T) {
	lc := NewLRUCache(100, time.Minute*1)

	// this is done in order to ensure that
	// we skip the lazy promotion flow.
	c := lc.(*lruCache)
	c.timeWindow = 0

	for i := 0; i < 100; i++ {
		c.Set(key(i), value(i))
	}

	// key0 being the first key that was added,
	// should still exist in the cache.
	val, err := c.Get(key(0))
	if val != value(0) || err != nil {
		t.Fatalf("expected oldest value to be found")
	}

	// The previous Get call should've moved key0
	// to the first position because of the cache hit.
	c.Set(key(101), value(101))
	val, err = c.Get(key(1))
	if val != "" || err != ErrKeyNotFound {
		t.Fatalf("expected oldest value to be evicted")
	}
}

func TestKeyExpiry(t *testing.T) {
	c := NewLRUCache(100, time.Millisecond*1)
	for i := 0; i < 100; i++ {
		c.Set(key(i), value(i))
	}

	time.Sleep(time.Millisecond * 1)

	// key0 being the first key that was added,
	// should still exist in the cache, but not
	// returned because it has expired.
	val, err := c.Get(key(0))
	if val != "" || err != ErrKeyNotFound {
		t.Fatalf("expected key to be deleted after expiry")
	}
}

func TestLazyKeyPromotion(t *testing.T) {
	c := NewLRUCache(100, time.Hour*1)
	for i := 0; i < 100; i++ {
		c.Set(key(i), value(i))
	}
	val, err := c.Get(key(0))

	c.Set(key(101), value(101))
	val, err = c.Get(key(0))
	if val != "" || err != ErrKeyNotFound {
		// this should happen because of our cache implementation.
		// the first key won't have moved to the front because we
		// don't promote keys on every Get call.
		// check defaultWindow in lru.go for details.
		t.Fatalf("expected oldest value to be evicted")
	}
}

func TestStaleDataCleanup(t *testing.T) {
	lc := NewLRUCache(100, time.Millisecond*1)
	for i := 0; i < 50; i++ {
		lc.Set(key(i), value(i))
	}

	time.Sleep(time.Second * 1)
	c := lc.(*lruCache)
	c.RLock()
	defer c.RUnlock()

	if len(c.lookupTable) == 50 {
		t.Fatalf("expected background worker to have cleaned up expired resource")
	}
}

func BenchmarkLRURandom(b *testing.B) {
	lc := NewLRUCache(8192, time.Hour*1)

	trace := make([]string, b.N*2)
	for i := 0; i < b.N*2; i++ {
		trace[i] = key(rand.Int() % 32768)
	}

	b.ResetTimer()

	var hit, miss int
	for i := 0; i < 2*b.N; i++ {
		if i%2 == 0 {
			lc.Set(trace[i], trace[i])
		} else {
			_, err := lc.Get(trace[i])
			if err != nil {
				miss++
			} else {
				hit++
			}
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}

// Frequent benchmark works by frequently
func BenchmarkLRUFrequent(b *testing.B) {
	lc := NewLRUCache(8192, time.Hour*1)

	unique := make(map[string]struct{}, 0)
	trace := make([]int, b.N*2)
	for i := 0; i < b.N*2; i++ {
		if i%2 == 0 {
			trace[i] = rand.Int() % 32768
		} else {
			trace[i] = rand.Int()
		}
	}

	for i := 0; i < b.N; i++ {
		if trace[i] <= 32768 {
			k := key(trace[i])
			if _, ok := unique[k]; !ok {
				unique[k] = struct{}{}
			} else {
				lc.Set(k, k)
			}
		}
	}

	b.ResetTimer()

	var hit, miss int
	for i := 0; i < b.N; i++ {
		_, err := lc.Get(key(trace[i]))
		if err != nil {
			miss++
		} else {
			hit++
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}
