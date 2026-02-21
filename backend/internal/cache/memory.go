package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrCacheMiss is returned when a key is not found in the cache.
var ErrCacheMiss = errors.New("cache: key not found")

type memEntry struct {
	value     string
	expiresAt time.Time
}

// MemoryCache implements CacheProvider using an in-process sync.Map.
// Suitable for single-instance deployments, local development, or as a
// fallback when Redis is unavailable (e.g., Lambda without Upstash).
type MemoryCache struct {
	data sync.Map
}

// NewMemoryCache creates a new in-memory cache.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{}
}

func (c *MemoryCache) Get(_ context.Context, key string) (string, error) {
	val, ok := c.data.Load(key)
	if !ok {
		return "", ErrCacheMiss
	}
	entry, ok := val.(memEntry)
	if !ok {
		return "", ErrCacheMiss
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		c.data.Delete(key)
		return "", ErrCacheMiss
	}
	return entry.value, nil
}

func (c *MemoryCache) Set(_ context.Context, key string, value string, ttl time.Duration) error {
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}
	c.data.Store(key, memEntry{value: value, expiresAt: expiresAt})
	return nil
}

func (c *MemoryCache) Del(_ context.Context, keys ...string) error {
	for _, k := range keys {
		c.data.Delete(k)
	}
	return nil
}
