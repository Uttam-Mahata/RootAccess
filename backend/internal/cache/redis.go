package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements CacheProvider backed by a Redis client.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a RedisCache wrapping the given Redis client.
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisCache) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Client returns the underlying Redis client for advanced operations
// that require direct Redis access (e.g., OAuth state management).
func (c *RedisCache) Client() *redis.Client {
	return c.client
}
