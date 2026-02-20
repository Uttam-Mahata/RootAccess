package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
)

// EventTokenMiddleware validates the X-Event-Token header and stores the resolved
// event in the Gin context under the key "event". Routes that require an event
// scope should apply this middleware.
func EventTokenMiddleware(orgService *services.OrganizationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Event-Token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-Event-Token header required"})
			c.Abort()
			return
		}

		event, err := orgService.ValidateEventToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired event token"})
			c.Abort()
			return
		}

		// Make the resolved event available to downstream handlers
		c.Set("event", event)
		c.Set("event_id", event.ID.Hex())
		c.Set("org_id", event.OrgID.Hex())

		c.Next()
	}
}
