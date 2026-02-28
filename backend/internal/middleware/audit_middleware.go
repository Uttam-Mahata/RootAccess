package middleware

import (
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// AuditMiddleware creates middleware that logs admin actions
func AuditMiddleware(auditService *services.AuditLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request first
		c.Next()

		// Only log successful admin operations (POST, PUT, DELETE)
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "DELETE" {
			return
		}

		// Only log if the response was successful (2xx)
		status := c.Writer.Status()
		if status < 200 || status >= 300 {
			return
		}

		userIDStr, exists := c.Get("user_id")
		if !exists {
			return
		}

		username, _ := c.Get("username")
		userID := userIDStr.(string)
		usernameStr, _ := username.(string)

		action := method + " " + c.FullPath()
		resource := c.FullPath()
		details := "Path: " + c.Request.URL.Path

		auditService.Log(userID, usernameStr, action, resource, details, c.ClientIP())
	}
}
