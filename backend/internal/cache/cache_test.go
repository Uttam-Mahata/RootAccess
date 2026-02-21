package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	if err := c.Set(ctx, "k1", "v1", 5*time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	val, err := c.Get(ctx, "k1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "v1" {
		t.Errorf("got %q, want %q", val, "v1")
	}
}

func TestMemoryCache_Miss(t *testing.T) {
	c := NewMemoryCache()
	_, err := c.Get(context.Background(), "nonexistent")
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss, got %v", err)
	}
}

func TestMemoryCache_Expiry(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	// Set with a very short TTL
	if err := c.Set(ctx, "exp", "val", 1*time.Millisecond); err != nil {
		t.Fatalf("Set: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	_, err := c.Get(ctx, "exp")
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss after expiry, got %v", err)
	}
}

func TestMemoryCache_Del(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	_ = c.Set(ctx, "a", "1", 5*time.Minute)
	_ = c.Set(ctx, "b", "2", 5*time.Minute)

	if err := c.Del(ctx, "a", "b"); err != nil {
		t.Fatalf("Del: %v", err)
	}
	_, err := c.Get(ctx, "a")
	if err != ErrCacheMiss {
		t.Error("expected miss after delete")
	}
	_, err = c.Get(ctx, "b")
	if err != ErrCacheMiss {
		t.Error("expected miss after delete")
	}
}

func TestMemoryCache_Overwrite(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	_ = c.Set(ctx, "k", "old", 5*time.Minute)
	_ = c.Set(ctx, "k", "new", 5*time.Minute)
	val, err := c.Get(ctx, "k")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "new" {
		t.Errorf("got %q, want %q", val, "new")
	}
}

// Compile-time interface compliance checks.
var _ CacheProvider = (*MemoryCache)(nil)
var _ CacheProvider = (*RedisCache)(nil)
