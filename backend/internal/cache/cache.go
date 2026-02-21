package cache

import (
	"context"
	"time"
)

// CacheProvider defines the interface for cache operations.
// Implementations can be backed by Redis, in-memory stores, or any
// other key-value system.
type CacheProvider interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}
