package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))
		usernameStr, _ := username.(string)

		action := method + " " + c.FullPath()
		resource := c.FullPath()
		details := "Path: " + c.Request.URL.Path

		auditService.Log(userID, usernameStr, action, resource, details, c.ClientIP())
	}
}
