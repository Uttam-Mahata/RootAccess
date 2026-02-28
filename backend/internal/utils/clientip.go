package utils

import (
	"os"

	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/gin-gonic/gin"
)

// GetClientIP returns the client IP from the request.
// On AWS Lambda, c.ClientIP() may be empty; this falls back to
// API Gateway's RequestContext.Identity.SourceIP.
func GetClientIP(c *gin.Context) string {
	ip := c.ClientIP()
	if ip != "" {
		return ip
	}
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		if gwCtx, ok := core.GetAPIGatewayContextFromContext(c.Request.Context()); ok && gwCtx.Identity.SourceIP != "" {
			return gwCtx.Identity.SourceIP
		}
	}
	return ip
}
