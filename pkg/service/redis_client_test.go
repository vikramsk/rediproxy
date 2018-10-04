package service

import (
	"flag"
	"testing"
)

var redisURL = flag.String("redis-url", "localhost:6379", "URL for Redis")

func init() {
	flag.Parse()
}

func TestRedisConnection(t *testing.T) {
	rc, err := NewRedisClient("")
	if rc != nil || err == nil {
		t.Fatalf("expected a failure in client creation")
	}

	rc, err = NewRedisClient(*redisURL)
	if err != nil || rc == nil {
		t.Fatalf("expected client to be created")
	}
}

func TestRedisGet(t *testing.T) {
	rc, err := NewRedisClient(*redisURL)
	key := "key"
	value := "value"

	// to setup repeatable tests
	c := rc.(*redisClient)
	c.client.Del(key)
	c.client.Set(key, value, 0)

	val, err := rc.Get(key)
	if err != nil || val != value {
		t.Fatalf("redis get failed")
	}
}
