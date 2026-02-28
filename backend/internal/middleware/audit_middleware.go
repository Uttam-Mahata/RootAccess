package middleware

import (
	"os"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/gin-gonic/gin"
)

// getClientIP returns the client IP from the request, with Lambda API Gateway fallback when ClientIP() is empty.
func getClientIP(c *gin.Context) string {
	ip := c.ClientIP()
	if ip != "" {
		return ip
	}
	// Fallback for Lambda: adapter stores RequestContext in the request's context
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		if gwCtx, ok := core.GetAPIGatewayContextFromContext(c.Request.Context()); ok && gwCtx.Identity.SourceIP != "" {
			return gwCtx.Identity.SourceIP
		}
	}
	return ip
}

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

		auditService.Log(userID, usernameStr, action, resource, details, getClientIP(c))
	}
}
