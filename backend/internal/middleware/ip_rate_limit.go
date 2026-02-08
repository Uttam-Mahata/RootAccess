package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// IPRateLimiter tracks requests per IP address
type IPRateLimiter struct {
	attempts map[string][]time.Time
	mu       sync.RWMutex
}

// Global IP rate limiter instance
var ipRateLimiter = &IPRateLimiter{
	attempts: make(map[string][]time.Time),
}

// IPRateLimitMiddleware creates middleware that rate limits by client IP
// maxAttempts: maximum number of attempts allowed within the time window
// window: time duration for the rate limit window
func IPRateLimitMiddleware(maxAttempts int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		path := c.FullPath()

		// Create unique key for IP + endpoint
		key := fmt.Sprintf("%s:%s", ip, path)

		ipRateLimiter.mu.Lock()
		defer ipRateLimiter.mu.Unlock()

		now := time.Now()

		// Clean up old attempts outside the window
		if attempts, ok := ipRateLimiter.attempts[key]; ok {
			validAttempts := make([]time.Time, 0)
			for _, t := range attempts {
				if now.Sub(t) <= window {
					validAttempts = append(validAttempts, t)
				}
			}
			ipRateLimiter.attempts[key] = validAttempts
		}

		// Check if IP has exceeded rate limit
		attempts := ipRateLimiter.attempts[key]
		if len(attempts) >= maxAttempts {
			oldestAttempt := attempts[0]
			waitTime := window - now.Sub(oldestAttempt)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests. Please wait before trying again.",
				"retry_after": int(waitTime.Seconds()),
			})
			c.Abort()
			return
		}

		// Record this attempt
		ipRateLimiter.attempts[key] = append(ipRateLimiter.attempts[key], now)

		// Add rate limit headers
		remaining := maxAttempts - len(ipRateLimiter.attempts[key])
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxAttempts))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		c.Next()
	}
}

// CleanupExpiredIPAttempts periodically cleans up expired IP rate limit entries
func CleanupExpiredIPAttempts(window time.Duration, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		ipRateLimiter.mu.Lock()
		now := time.Now()
		for key, attempts := range ipRateLimiter.attempts {
			validAttempts := make([]time.Time, 0)
			for _, t := range attempts {
				if now.Sub(t) <= window {
					validAttempts = append(validAttempts, t)
				}
			}
			if len(validAttempts) == 0 {
				delete(ipRateLimiter.attempts, key)
			} else {
				ipRateLimiter.attempts[key] = validAttempts
			}
		}
		ipRateLimiter.mu.Unlock()
	}
}
