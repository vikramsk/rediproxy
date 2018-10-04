package cache

import "errors"

// Cacher defines the interface for a generic
// cache that supports both reads and writes.
type Cacher interface {
	Getter
	Setter
}

// Getter defines the behavior for a
// read-only store.
type Getter interface {
	Get(key string) (string, error)
}

// Setter defines the behavior for a
// write-only store.
type Setter interface {
	Set(key, value string)
}

// ErrKeyNotFound is the error returned when the
// key is not present in the store.
var ErrKeyNotFound = errors.New("cache: key not found")
