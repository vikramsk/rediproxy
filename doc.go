// Package rediproxy provides an HTTP proxy service that
// sits in front of Redis and provides the following:
// - HTTP GET endpoint to retrieve values.
// - Local LRU cache to persist data with a defined TTL. This cache sits in front of a backing Redis instance.
package rediproxy
