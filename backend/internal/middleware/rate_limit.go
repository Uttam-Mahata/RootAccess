package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter tracks submission attempts per user per challenge
type RateLimiter struct {
	attempts map[string][]time.Time
	mu       sync.RWMutex
}

// Global rate limiter instance
var flagSubmitLimiter = &RateLimiter{
	attempts: make(map[string][]time.Time),
}

// RateLimitMiddleware creates a middleware that limits requests per user per challenge
// maxAttempts: maximum number of attempts allowed within the time window
// window: time duration for the rate limit window
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

		// Create unique key for user+challenge combination
		key := userID.(string) + ":" + challengeID

		flagSubmitLimiter.mu.Lock()
		defer flagSubmitLimiter.mu.Unlock()

		now := time.Now()

		// Clean up old attempts outside the window
		if attempts, ok := flagSubmitLimiter.attempts[key]; ok {
			validAttempts := make([]time.Time, 0)
			for _, t := range attempts {
				if now.Sub(t) <= window {
					validAttempts = append(validAttempts, t)
				}
			}
			flagSubmitLimiter.attempts[key] = validAttempts
		}

		// Check if user has exceeded rate limit
		attempts := flagSubmitLimiter.attempts[key]
		if len(attempts) >= maxAttempts {
			// Calculate time until next attempt is allowed
			oldestAttempt := attempts[0]
			waitTime := window - now.Sub(oldestAttempt)
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many attempts. Please wait before trying again.",
				"retry_after": int(waitTime.Seconds()),
				"message":     "Rate limit exceeded for this challenge",
			})
			c.Abort()
			return
		}

		// Record this attempt
		flagSubmitLimiter.attempts[key] = append(flagSubmitLimiter.attempts[key], now)

		// Add rate limit headers
		remaining := maxAttempts - len(flagSubmitLimiter.attempts[key])
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxAttempts))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		c.Next()
	}
}

// CleanupExpiredAttempts periodically cleans up expired rate limit entries
// Call this in a goroutine during application startup
func CleanupExpiredAttempts(window time.Duration, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		flagSubmitLimiter.mu.Lock()
		now := time.Now()
		for key, attempts := range flagSubmitLimiter.attempts {
			validAttempts := make([]time.Time, 0)
			for _, t := range attempts {
				if now.Sub(t) <= window {
					validAttempts = append(validAttempts, t)
				}
			}
			if len(validAttempts) == 0 {
				delete(flagSubmitLimiter.attempts, key)
			} else {
				flagSubmitLimiter.attempts[key] = validAttempts
			}
		}
		flagSubmitLimiter.mu.Unlock()
	}
}
