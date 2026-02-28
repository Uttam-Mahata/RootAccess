package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// localLimiter is an in-memory fallback for when Redis is unavailable.
type localLimiter struct {
	attempts map[string][]time.Time
	mu       sync.Mutex
}

var flagSubmitLimiter = &localLimiter{
	attempts: make(map[string][]time.Time),
}

// RateLimitMiddleware limits requests per user per challenge.
// Uses Redis (distributed) with in-memory fallback for local dev.
func RateLimitMiddleware(maxAttempts int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		challengeID := c.Param("id")
		if challengeID == "" {
			c.Next()
			return
		}

		key := userID.(string) + ":" + challengeID

		if rdb := getRateLimitRedis(); rdb != nil {
			limited, remaining, retryAfter := checkRedisRateLimit(rdb, "rl:flag:"+key, maxAttempts, window)
			setRateLimitHeaders(c, maxAttempts, remaining)
			if limited {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "Too many attempts. Please wait before trying again.",
					"retry_after": retryAfter,
					"message":     "Rate limit exceeded for this challenge",
				})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// In-memory fallback (local dev / no Redis)
		limited, remaining, retryAfter := checkLocalRateLimit(flagSubmitLimiter, key, maxAttempts, window)
		setRateLimitHeaders(c, maxAttempts, remaining)
		if limited {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many attempts. Please wait before trying again.",
				"retry_after": retryAfter,
				"message":     "Rate limit exceeded for this challenge",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// --- Shared helpers (used by both rate_limit.go and ip_rate_limit.go) ---

// getRateLimitRedis returns the General Redis client for rate limiting, or nil.
func getRateLimitRedis() *redis.Client {
	if database.Registry != nil {
		return database.Registry.General
	}
	return nil
}

// checkRedisRateLimit increments a fixed-window counter in Redis.
// Each window gets its own key that auto-expires via TTL.
// Returns (isLimited, remaining, retryAfterSeconds).
func checkRedisRateLimit(rdb *redis.Client, key string, maxAttempts int, window time.Duration) (bool, int, int) {
	ctx := context.Background()
	windowSec := int64(window.Seconds())
	bucket := time.Now().Unix() / windowSec
	redisKey := fmt.Sprintf("%s:%d", key, bucket)

	count, err := rdb.Incr(ctx, redisKey).Result()
	if err != nil {
		// Redis error — fail open (allow the request)
		return false, maxAttempts, 0
	}

	// Set TTL on first increment so the key auto-expires
	if count == 1 {
		rdb.Expire(ctx, redisKey, window+time.Second)
	}

	if int(count) > maxAttempts {
		windowStart := bucket * windowSec
		elapsed := time.Now().Unix() - windowStart
		retryAfter := windowSec - elapsed
		if retryAfter < 1 {
			retryAfter = 1
		}
		return true, 0, int(retryAfter)
	}

	return false, maxAttempts - int(count), 0
}

// checkLocalRateLimit is the in-memory sliding-window fallback.
// Returns (isLimited, remaining, retryAfterSeconds).
func checkLocalRateLimit(limiter *localLimiter, key string, maxAttempts int, window time.Duration) (bool, int, int) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	now := time.Now()

	// Prune expired entries
	if attempts, ok := limiter.attempts[key]; ok {
		valid := attempts[:0]
		for _, t := range attempts {
			if now.Sub(t) <= window {
				valid = append(valid, t)
			}
		}
		limiter.attempts[key] = valid
	}

	attempts := limiter.attempts[key]
	if len(attempts) >= maxAttempts {
		waitTime := window - now.Sub(attempts[0])
		if waitTime < time.Second {
			waitTime = time.Second
		}
		return true, 0, int(waitTime.Seconds())
	}

	limiter.attempts[key] = append(limiter.attempts[key], now)
	remaining := maxAttempts - len(limiter.attempts[key])
	return false, remaining, 0
}

func setRateLimitHeaders(c *gin.Context, limit, remaining int) {
	c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
	c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
}
