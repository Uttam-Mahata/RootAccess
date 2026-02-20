package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WriteupHandler struct {
	writeupService     *services.WriteupService
	contestAdminService *services.ContestAdminService
}

func NewWriteupHandler(writeupService *services.WriteupService) *WriteupHandler {
	return &WriteupHandler{
		writeupService: writeupService,
	}
}

func NewWriteupHandlerWithContestAdmin(writeupService *services.WriteupService, contestAdminService *services.ContestAdminService) *WriteupHandler {
	return &WriteupHandler{
		writeupService:      writeupService,
		contestAdminService: contestAdminService,
	}
}

type CreateWriteupRequest struct {
	Content       string `json:"content" binding:"required"`
	ContentFormat string `json:"content_format"` // "markdown" or "html"
}

// CreateWriteup handles creating a new writeup for a challenge
// @Summary Submit a writeup
// @Description Submit a writeup for a challenge that the user has already solved.
// @Tags Writeups
// @Accept json
// @Produce json
// @Param id path string true "Challenge ID"
// @Param request body CreateWriteupRequest true "Writeup content"
// @Success 201 {object} models.Writeup
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /challenges/{id}/writeups [post]
func (h *WriteupHandler) CreateWriteup(c *gin.Context) {
	challengeID := c.Param("id")
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	username, _ := c.Get("username")

	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	var req CreateWriteupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate and set default format
	contentFormat := req.ContentFormat
	if contentFormat == "" {
		contentFormat = "markdown" // Default to markdown
	}
	if contentFormat != "markdown" && contentFormat != "html" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content_format must be 'markdown' or 'html'"})
		return
	}

	writeup, err := h.writeupService.CreateWriteup(userID, username.(string), challengeID, req.Content, contentFormat)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Writeup submitted for review",
		"writeup": writeup,
	})
}

// GetWriteups returns approved writeups for a challenge (hidden until contest has ended)
// @Summary Get writeups for a challenge
// @Description Retrieve all approved writeups for a specific challenge.
// @Tags Writeups
// @Produce json
// @Param id path string true "Challenge ID"
// @Success 200 {array} models.Writeup
// @Security ApiKeyAuth
// @Router /challenges/{id}/writeups [get]
func (h *WriteupHandler) GetWriteups(c *gin.Context) {
	challengeID := c.Param("id")

	// Hide writeups until the challenge's contest has ended
	if h.contestAdminService != nil {
		ended, err := h.contestAdminService.HasContestEndedForChallenge(challengeID, time.Now())
		if err != nil || !ended {
			c.JSON(http.StatusOK, []models.Writeup{})
			return
		}
	}

	writeups, err := h.writeupService.GetWriteupsByChallenge(challengeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if writeups == nil {
		writeups = []models.Writeup{}
	}

	c.JSON(http.StatusOK, writeups)
}

// GetMyWriteups returns the current user's writeups
// @Summary Get my writeups
// @Description Retrieve all writeups submitted by the authenticated user.
// @Tags Writeups
// @Produce json
// @Success 200 {array} models.Writeup
// @Security ApiKeyAuth
// @Router /writeups/my [get]
func (h *WriteupHandler) GetMyWriteups(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	writeups, err := h.writeupService.GetMyWriteups(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if writeups == nil {
		writeups = []models.Writeup{}
	}

	c.JSON(http.StatusOK, writeups)
}

// GetAllWriteups returns all writeups for admin
// @Summary Get all writeups
// @Description Retrieve all writeups, including pending and rejected ones. Optionally filter by team_id query parameter.
// @Tags Admin Writeups
// @Produce json
// @Param team_id query string false "Filter by team ID"
// @Success 200 {array} models.Writeup
// @Security ApiKeyAuth
// @Router /admin/writeups [get]
func (h *WriteupHandler) GetAllWriteups(c *gin.Context) {
	teamID := c.Query("team_id")
	writeups, err := h.writeupService.GetAllWriteups(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if writeups == nil {
		writeups = []models.Writeup{}
	}

	c.JSON(http.StatusOK, writeups)
}

type UpdateWriteupStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateWriteupStatus updates writeup approval status (admin)
// @Summary Update writeup status
// @Description Approve or reject a writeup.
// @Tags Admin Writeups
// @Accept json
// @Produce json
// @Param id path string true "Writeup ID"
// @Param request body UpdateWriteupStatusRequest true "New status"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/writeups/{id}/status [put]
func (h *WriteupHandler) UpdateWriteupStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateWriteupStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.writeupService.UpdateWriteupStatus(id, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Writeup status updated"})
}

// DeleteWriteup deletes a writeup (admin)
// @Summary Delete writeup
// @Description Permanently delete a writeup.
// @Tags Admin Writeups
// @Param id path string true "Writeup ID"
// @Success 200 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/writeups/{id} [delete]
func (h *WriteupHandler) DeleteWriteup(c *gin.Context) {
	id := c.Param("id")

	if err := h.writeupService.DeleteWriteup(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Writeup deleted"})
}

// UpdateWriteup allows authors to edit their writeup
// @Summary Update writeup content
// @Description Modify the content or format of a writeup. Only the author can perform this action.
// @Tags Writeups
// @Accept json
// @Produce json
// @Param id path string true "Writeup ID"
// @Param request body CreateWriteupRequest true "New content"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /writeups/{id} [put]
func (h *WriteupHandler) UpdateWriteup(c *gin.Context) {
	id := c.Param("id")
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	var req CreateWriteupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate and set default format
	contentFormat := req.ContentFormat
	if contentFormat == "" {
		contentFormat = "markdown" // Default to markdown
	}
	if contentFormat != "markdown" && contentFormat != "html" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content_format must be 'markdown' or 'html'"})
		return
	}

	if err := h.writeupService.UpdateWriteupContent(id, userID, req.Content, contentFormat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Writeup updated"})
}

// ToggleUpvote toggles a user's upvote on a writeup
// @Summary Toggle writeup upvote
// @Description Add or remove an upvote for a writeup.
// @Tags Writeups
// @Produce json
// @Param id path string true "Writeup ID"
// @Success 200 {object} map[string]bool
// @Security ApiKeyAuth
// @Router /writeups/{id}/upvote [post]
func (h *WriteupHandler) ToggleUpvote(c *gin.Context) {
	id := c.Param("id")
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	upvoted, err := h.writeupService.ToggleUpvote(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"upvoted": upvoted})
}
