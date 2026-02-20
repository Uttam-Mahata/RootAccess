//go:build production
// +build production

package routes

import (
	"github.com/gin-gonic/gin"
)

func registerSwagger(r *gin.Engine) {
	// Swagger is disabled in production builds
}
