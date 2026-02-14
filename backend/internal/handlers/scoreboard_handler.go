package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/services"
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
