package utils

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RespondWithError sends a sanitized error response to the client
// while logging the full error details internally.
func RespondWithError(c *gin.Context, code int, message string, err error) {
	if err != nil {
		log.Printf("[ERROR] %s: %v", message, err)
	}
	
	// In production, we should avoid sending the raw error message
	// for 500 Internal Server Errors.
	responseMessage := message
	if code == http.StatusInternalServerError && responseMessage == "" {
		responseMessage = "An unexpected error occurred. Please try again later."
	}

	c.JSON(code, gin.H{"error": responseMessage})
}
