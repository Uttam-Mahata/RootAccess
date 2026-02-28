package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var ipLocalLimiter = &localLimiter{
	attempts: make(map[string][]time.Time),
}

// IPRateLimitMiddleware limits requests per client IP per endpoint.
// Uses Redis (distributed) with in-memory fallback for local dev.
func IPRateLimitMiddleware(maxAttempts int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		path := c.FullPath()
		key := fmt.Sprintf("%s:%s", ip, path)

		if rdb := getRateLimitRedis(); rdb != nil {
			limited, remaining, retryAfter := checkRedisRateLimit(rdb, "rl:ip:"+key, maxAttempts, window)
			setRateLimitHeaders(c, maxAttempts, remaining)
			if limited {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "Too many requests. Please wait before trying again.",
					"retry_after": retryAfter,
				})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// In-memory fallback (local dev / no Redis)
		limited, remaining, retryAfter := checkLocalRateLimit(ipLocalLimiter, key, maxAttempts, window)
		setRateLimitHeaders(c, maxAttempts, remaining)
		if limited {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests. Please wait before trying again.",
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
