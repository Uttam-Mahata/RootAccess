package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
)

// ContestTimeMiddleware validates that the contest is running before allowing submissions
func ContestTimeMiddleware(contestService *services.ContestService) gin.HandlerFunc {
	return func(c *gin.Context) {
		status, _, err := contestService.GetContestStatus()
		if err != nil {
			// If no contest config, allow access (no time restrictions)
			c.Next()
			return
		}

		if status == models.ContestStatusNotStarted {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  "Contest has not started yet",
				"status": string(status),
			})
			c.Abort()
			return
		}

		if status == models.ContestStatusEnded {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  "Contest has ended",
				"status": string(status),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
