package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/services"
)

type ScoreboardHandler struct {
	scoreboardService *services.ScoreboardService
}

func NewScoreboardHandler(scoreboardService *services.ScoreboardService) *ScoreboardHandler {
	return &ScoreboardHandler{
		scoreboardService: scoreboardService,
	}
}

func (h *ScoreboardHandler) GetScoreboard(c *gin.Context) {
	scores, err := h.scoreboardService.GetScoreboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, scores)
}
