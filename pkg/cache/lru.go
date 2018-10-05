package cache

import (
	"container/list"
	"sync"
	"time"
)

// defaultWindowPercent is used to specify
// the min time window which needs to exist
// between moving an element from the linked
// list to the front.
// the value is a float which represents
// the percentage of the global expiry
// for the keys.
//
// example:
//
// 0.05 for an expiry of 60 min
// implies that time difference between
// successive moves of an element to the
// front will at least be 3 min.
const defaultWindowPercent = 0.05

// defaultSampleSize is used to randomly pick
// elements from the cache and evict them if
// they have expired.
const defaultSampleSize = 20

// item represents a single cache
// entry for an LRU cache.
type item struct {
	key   string
	value string

	// movedAt defines the time at
	// which the items was moved to
	// the front of the list.
	movedAt time.Time

	// expiry refers to the time
	// at which the key is invalid.
	// times are stored in UTC.
	expiry time.Time

	// element points to the item
	// in the linked list used
	// for managing the LRU chain.
	element *list.Element
}

type lruCache struct {
	// capacity is the max. size
	// of the cache.
	capacity int

	// ttl defines the ttl for
	// keys added to the cache.
	ttl time.Duration

	// timeWindow is the duration
	// between consecutive moves to the front
	// for the same element in the linked
	// list used to track LRU behavior.
	// this value is set using the
	// defaultWindowPercent
	timeWindow time.Duration

	// this is the mutex protecting the
	// lookupTable and the list.
	sync.RWMutex
	lookupTable map[string]*item
	list        *list.List
}

// NewLRUCache is used to initialize an LRU cache.
// It accepts the capacity of the cache, time to live
// for the objects in the cache.
func NewLRUCache(c int, t time.Duration) Cacher {
	tw := time.Duration(int(defaultWindowPercent * float64(t)))
	lc := &lruCache{
		capacity:    c,
		ttl:         t,
		timeWindow:  tw,
		lookupTable: make(map[string]*item, c),
		list:        list.New(),
	}
	go lc.runCleanup()
	return lc
}

// Get looks up the key in the in-memory LRU cache.
// It returns an error if the key isn't present in
// the cache.
func (lc *lruCache) Get(key string) (string, error) {
	it, move, del, err := lc.searchItem(key)
	if err != nil {
		return "", err
	} else if del {
		lc.removeItem(it)
		return "", ErrKeyNotFound
	}

	if move {
		lc.moveItemFront(it)
	}
	return it.value, nil
}

// Set adds the key value pair to the cache, ensuring
// that it adheres to the constraints on the capacity.
func (lc *lruCache) Set(k, v string) {
	i := &item{
		key:     k,
		value:   v,
		movedAt: time.Now().UTC(),
		expiry:  time.Now().UTC().Add(lc.ttl),
	}

	lc.Lock()
	defer lc.Unlock()

	if lc.isFull() {
		it := lc.list.Back().Value.(*item)
		delete(lc.lookupTable, it.key)
		lc.list.Remove(lc.list.Back())
	}

	elem := lc.list.PushFront(i)
	i.element = elem

	lc.lookupTable[k] = i
}

// searchKey looks up the key in the cache.
// It returns the following:
// 	- item for the key, if it's valid.
//  - true/false if the item needs to be moved to the front.
//  - true/false if the item needs to be deleted.
//  - error if any.
func (lc *lruCache) searchItem(key string) (*item, bool, bool, error) {
	var it *item
	var ok bool

	lc.RLock()
	defer lc.RUnlock()

	if it, ok = lc.lookupTable[key]; !ok {
		return nil, false, false, ErrKeyNotFound
	}

	now := time.Now().UTC()

	// check if item has expired
	if it.expiry.Sub(now) < 0 {
		return it, false, true, nil
	}

	// check if item needs to be moved
	if now.Sub(it.movedAt) > lc.timeWindow {
		return it, true, false, nil
	}

	return it, false, false, nil
}

// isFull checks if the lru cache has
// hit the capacity.
func (lc *lruCache) isFull() bool {
	if len(lc.lookupTable) == lc.capacity {
		return true
	}
	return false
}

// removeItem removes an item from the lookuptable
// and the list.
func (lc *lruCache) removeItem(i *item) {
	lc.Lock()
	defer lc.Unlock()
	if i.element == nil {
		return
	}
	lc.list.Remove(i.element)
	delete(lc.lookupTable, i.key)
	i.element = nil

}

// moveItemFront detaches an item from the list
// and moves it to the front.
func (lc *lruCache) moveItemFront(i *item) {
	lc.Lock()
	defer lc.Unlock()
	if i.element == nil {
		return
	}
	lc.list.Remove(i.element)
	i.movedAt = time.Now().UTC()
	elem := lc.list.PushFront(i)
	i.element = elem
}

// runCleanup is a background worker that picks
// up random keys from the cache and checks if
// they have expired. It runs the removeStaleData
// func 10 times in 1 second.
// This is inspired by the Redis EXPIRE strategy.
// https://redis.io/commands/expire#how-redis-expires-keys
func (lc *lruCache) runCleanup() {
	var keys map[string]struct{}
	for {
		if len(keys) <= 0.75*defaultSampleSize {
			keys = lc.getRandomKeys(defaultSampleSize)
		}
		lc.removeStaleData(keys)
		time.Sleep(time.Duration(1e9 / 10))
	}
}

// removeStaleData iterates through the provided keys
// and removed any expired items from the cache.
func (lc *lruCache) removeStaleData(keys map[string]struct{}) {
	for k, _ := range keys {
		it, _, del, err := lc.searchItem(k)
		// key has already been deleted
		if err != nil {
			delete(keys, k)
			continue
		}
		if del {
			delete(keys, k)
			lc.removeItem(it)
		}
	}
}

// getRandomKeys fetches random keys from the lookupTable
// equal to the count passed as input. It returns a slice.
// Note that this is randomness isn't truly random as
// the go runtime doesn't make any guarantees on the
// order of the traversal of a map.
func (lc *lruCache) getRandomKeys(s int) map[string]struct{} {
	lc.RLock()
	defer lc.RUnlock()

	keys := make(map[string]struct{}, s)
	count := 0
	for k, _ := range lc.lookupTable {
		keys[k] = struct{}{}
		count++
		if count == s {
			break
		}
	}

	return keys
}
