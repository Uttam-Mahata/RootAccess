package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/services"
)

type LeaderboardHandler struct {
	scoreboardService *services.ScoreboardService
}

func NewLeaderboardHandler(scoreboardService *services.ScoreboardService) *LeaderboardHandler {
	return &LeaderboardHandler{scoreboardService: scoreboardService}
}

// GetCategoryLeaderboard returns scoreboard filtered by category
func (h *LeaderboardHandler) GetCategoryLeaderboard(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category parameter required"})
		return
	}
	// For now, get the full scoreboard - category filtering can be enhanced later
	scoreboard, err := h.scoreboardService.GetScoreboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"category":   category,
		"scoreboard": scoreboard,
	})
}

// GetTimeBasedLeaderboard returns scoreboard for a time period
func (h *LeaderboardHandler) GetTimeBasedLeaderboard(c *gin.Context) {
	period := c.DefaultQuery("period", "all")
	var since time.Time
	now := time.Now()
	switch period {
	case "24h":
		since = now.Add(-24 * time.Hour)
	case "weekly":
		since = now.Add(-7 * 24 * time.Hour)
	case "monthly":
		since = now.Add(-30 * 24 * time.Hour)
	default:
		since = time.Time{}
	}

	_ = since // Will be used for filtered queries
	scoreboard, err := h.scoreboardService.GetScoreboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"period":     period,
		"scoreboard": scoreboard,
	})
}
