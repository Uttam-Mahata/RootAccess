package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
)

type ScoreboardHandler struct {
	scoreboardService *services.ScoreboardService
	contestService    *services.ContestService
}

func NewScoreboardHandler(scoreboardService *services.ScoreboardService, contestService *services.ContestService) *ScoreboardHandler {
	return &ScoreboardHandler{
		scoreboardService: scoreboardService,
		contestService:    contestService,
	}
}

// GetScoreboard returns the current individual scoreboard
// @Summary Get individual scoreboard
// @Description Retrieve the current leaderboard for individual users, sorted by points.
// @Tags Scoreboard
// @Produce json
// @Success 200 {array} services.UserScore
// @Failure 403 {object} map[string]string
// @Router /scoreboard [get]
func (h *ScoreboardHandler) GetScoreboard(c *gin.Context) {
	// Check scoreboard visibility
	if h.contestService != nil {
		config, err := h.contestService.GetContestConfig()
		if err == nil && config != nil && config.ScoreboardVisibility == "hidden" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Scoreboard is currently hidden"})
			return
		}
	}

	scores, err := h.scoreboardService.GetScoreboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, scores)
}

// GetTeamScoreboard returns the current team scoreboard
// @Summary Get team scoreboard
// @Description Retrieve the current leaderboard for teams, sorted by points.
// @Tags Scoreboard
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]string
// @Router /scoreboard/teams [get]
func (h *ScoreboardHandler) GetTeamScoreboard(c *gin.Context) {
	// Check scoreboard visibility
	if h.contestService != nil {
		config, err := h.contestService.GetContestConfig()
		if err == nil && config != nil && config.ScoreboardVisibility == "hidden" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Scoreboard is currently hidden"})
			return
		}
	}

	scores, err := h.scoreboardService.GetTeamScoreboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"teams": scores,
	})
}

// GetTeamStatistics returns team score progression statistics
// @Summary Get team statistics
// @Description Retrieve team score progression over time for charts
// @Tags Scoreboard
// @Produce json
// @Param days query int false "Number of days (default: 30, max: 90)"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]string
// @Router /scoreboard/teams/statistics [get]
func (h *ScoreboardHandler) GetTeamStatistics(c *gin.Context) {
	// Check scoreboard visibility
	if h.contestService != nil {
		config, err := h.contestService.GetContestConfig()
		if err == nil && config != nil && config.ScoreboardVisibility == "hidden" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Scoreboard is currently hidden"})
			return
		}
	}

	days := 30
	if daysParam := c.Query("days"); daysParam != "" {
		if parsedDays := parseInt(daysParam); parsedDays > 0 {
			days = parsedDays
		}
	}

	stats, err := h.scoreboardService.GetTeamScoreProgression(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"progressions": stats,
	})
}

func parseInt(s string) int {
	var result int
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		} else {
			return 0
		}
	}
	return result
}
