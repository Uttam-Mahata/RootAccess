package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/utils"
)

type LeaderboardHandler struct {
	scoreboardService *services.ScoreboardService
}

func NewLeaderboardHandler(scoreboardService *services.ScoreboardService) *LeaderboardHandler {
	return &LeaderboardHandler{scoreboardService: scoreboardService}
}

// GetCategoryLeaderboard returns scoreboard filtered by category
// @Summary Get category leaderboard
// @Description Retrieve the scoreboard filtered by a specific challenge category.
// @Tags Leaderboard
// @Produce json
// @Param category query string true "Challenge category"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /leaderboard/category [get]
func (h *LeaderboardHandler) GetCategoryLeaderboard(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category parameter required"})
		return
	}
	contestID := c.Query("contest_id")
	scoreboard, err := h.scoreboardService.GetScoreboard(contestID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"category":   category,
		"scoreboard": scoreboard,
	})
}

// GetTimeBasedLeaderboard returns scoreboard for a time period
// @Summary Get time-based leaderboard
// @Description Retrieve the scoreboard for a specific time period (24h, weekly, monthly, or all).
// @Tags Leaderboard
// @Produce json
// @Param period query string false "Time period" Enums(24h, weekly, monthly, all) default(all)
// @Success 200 {object} map[string]interface{}
// @Router /leaderboard/time [get]
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
	contestID := c.Query("contest_id")
	scoreboard, err := h.scoreboardService.GetScoreboard(contestID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"period":     period,
		"scoreboard": scoreboard,
	})
}
