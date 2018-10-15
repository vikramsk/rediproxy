package service

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/vikramsk/rediproxy/pkg/cache"
)

type redisClient struct {
	client *redis.Client
}

// NewRedisClient initializes a wrapper around the
// go-redis/redis client.
func NewRedisClient(addr string) (cache.Getter, error) {
	redisdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	_, err := redisdb.Ping().Result()
	if err != nil {
		return nil, fmt.Errorf("service: could not initialize redis client. err: %v", err)
	}

	return &redisClient{redisdb}, nil
}

// Get calls the underlying redis instance to fetch the
// data for the given key.
func (rc *redisClient) Get(key string) (string, error) {
	cmd := rc.client.Get(key)
	if cmd.Err() != nil {
		if cmd.Err() == redis.Nil {
			return "", cache.ErrKeyNotFound
		}
		return "", fmt.Errorf("service: error while reading key %s, err: %v", key, cmd.Err())
	}

	return cmd.Val(), nil
}
