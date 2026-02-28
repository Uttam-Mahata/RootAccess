package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestIPRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Reset the global rate limiter for testing
	ipRateLimiter = &IPRateLimiter{
		attempts: make(map[string][]time.Time),
	}

	maxAttempts := 3
	window := 1 * time.Minute

	middleware := IPRateLimitMiddleware(maxAttempts, window)

	t.Run("allows requests under limit", func(t *testing.T) {
		for i := 0; i < maxAttempts; i++ {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/auth/login", nil)
			c.Request.RemoteAddr = "192.168.1.1:1234"

			middleware(c)

			if w.Code == http.StatusTooManyRequests {
				t.Errorf("Request %d should not be rate limited", i+1)
			}
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/auth/login", nil)
		c.Request.RemoteAddr = "192.168.1.1:1234"

		middleware(c)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("Request should be rate limited, got status %d", w.Code)
		}
	})

	t.Run("different IPs have separate limits", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/auth/login", nil)
		c.Request.RemoteAddr = "10.0.0.1:1234"

		middleware(c)

		if w.Code == http.StatusTooManyRequests {
			t.Error("Different IP should not be rate limited")
		}
	})
}
