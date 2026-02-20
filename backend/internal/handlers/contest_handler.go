package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
)

type ContestHandler struct {
	contestService *services.ContestService
}

func NewContestHandler(contestService *services.ContestService) *ContestHandler {
	return &ContestHandler{
		contestService: contestService,
	}
}

// GetContestStatus returns the current contest status (public)
// @Summary Get contest status
// @Description Retrieve the current status and basic configuration of the contest.
// @Tags Contest
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /contest/status [get]
func (h *ContestHandler) GetContestStatus(c *gin.Context) {
	status, config, err := h.contestService.GetContestStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"status": string(status),
	}

	if config != nil {
		response["title"] = config.Title
		response["start_time"] = config.StartTime
		response["end_time"] = config.EndTime
		response["is_active"] = config.IsActive
		response["is_paused"] = config.IsPaused
		response["scoreboard_visibility"] = config.ScoreboardVisibility
		if config.FreezeTime != nil {
			response["freeze_time"] = config.FreezeTime
			response["is_frozen"] = config.IsScoreboardFrozen()
		}
	}

	c.JSON(http.StatusOK, response)
}

// UpdateContestRequest represents the request body for updating contest config
type UpdateContestRequest struct {
	Title                string  `json:"title" binding:"required"`
	StartTime            string  `json:"start_time" binding:"required"`
	EndTime              string  `json:"end_time" binding:"required"`
	FreezeTime           *string `json:"freeze_time"`
	IsActive             bool    `json:"is_active"`
	IsPaused             bool    `json:"is_paused"`
	ScoreboardVisibility string  `json:"scoreboard_visibility"`
}

// UpdateContestConfig updates the contest configuration (admin only)
// @Summary Update contest configuration
// @Description Update the global contest timing and visibility settings.
// @Tags Admin Contest
// @Accept json
// @Produce json
// @Param request body UpdateContestRequest true "Contest configuration"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/contest [put]
func (h *ContestHandler) UpdateContestConfig(c *gin.Context) {
	var req UpdateContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format, use RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format, use RFC3339"})
		return
	}

	var freezeTime *time.Time
	if req.FreezeTime != nil && *req.FreezeTime != "" {
		ft, err := time.Parse(time.RFC3339, *req.FreezeTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid freeze_time format, use RFC3339"})
			return
		}
		freezeTime = &ft
	}

	// Validate scoreboard visibility
	visibility := req.ScoreboardVisibility
	if visibility != "" && visibility != "public" && visibility != "private" && visibility != "hidden" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scoreboard_visibility must be 'public', 'private', or 'hidden'"})
		return
	}

	config, err := h.contestService.UpdateContestConfig(req.Title, startTime, endTime, freezeTime, req.IsActive, req.IsPaused, visibility)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Contest configuration updated",
		"config":  config,
	})
}

// GetContestConfig returns the full contest config (admin)
// @Summary Get contest configuration
// @Description Retrieve the full contest configuration for administrative purposes.
// @Tags Admin Contest
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security ApiKeyAuth
// @Router /admin/contest [get]
func (h *ContestHandler) GetContestConfig(c *gin.Context) {
	config, err := h.contestService.GetContestConfig()
	if err != nil {
		// No config yet
		c.JSON(http.StatusOK, gin.H{
			"message": "No contest configured yet",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"config": config,
		"status": string(config.GetStatus()),
	})
}
