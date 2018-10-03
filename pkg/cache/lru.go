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
	return &lruCache{
		capacity:    c,
		ttl:         t,
		timeWindow:  tw,
		lookupTable: make(map[string]*item, c),
		list:        list.New(),
	}
}

// Get looks up the key in the in-memory LRU cache.
// It returns an error if the key isn't present in
// the cache.
func (lc *lruCache) Get(key string) (string, error) {
	it, move, del, err := lc.searchItem(key)
	if err != nil {
		return "", err
	}

	if move {
		lc.moveItemFront(it)
	} else if del {
		lc.removeItem(it)
	}
	return it.value, nil
}

// Set adds the key value pair to the cache, ensuring
// that it adheres to the constraints on the capacity.
func (lc *lruCache) Set(k, v string) {
	i := &item{
		key:    k,
		value:  v,
		expiry: time.Now().UTC().Add(lc.ttl),
	}

	lc.Lock()
	defer lc.Unlock()

	if lc.isFull() {
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
		return nil, false, true, nil
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
	if len(lc.lookupTable) >= lc.capacity {
		return true
	}
	return false
}

// removeItem removes an item from the lookuptable
// and the list.
func (lc *lruCache) removeItem(i *item) {
	lc.Lock()
	defer lc.Unlock()
	lc.list.Remove(i.element)
	delete(lc.lookupTable, i.key)

}

// moveItemFront detaches an item from the list
// and moves it to the front.
func (lc *lruCache) moveItemFront(i *item) {
	lc.Lock()
	defer lc.Unlock()
	lc.list.Remove(i.element)
	elem := lc.list.PushFront(i)
	i.element = elem
}
