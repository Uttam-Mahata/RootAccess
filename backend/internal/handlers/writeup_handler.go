package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WriteupHandler struct {
	writeupService *services.WriteupService
}

func NewWriteupHandler(writeupService *services.WriteupService) *WriteupHandler {
	return &WriteupHandler{
		writeupService: writeupService,
	}
}

type CreateWriteupRequest struct {
	Content string `json:"content" binding:"required"`
}

// CreateWriteup handles creating a new writeup for a challenge
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

	writeup, err := h.writeupService.CreateWriteup(userID, username.(string), challengeID, req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Writeup submitted for review",
		"writeup": writeup,
	})
}

// GetWriteups returns approved writeups for a challenge
func (h *WriteupHandler) GetWriteups(c *gin.Context) {
	challengeID := c.Param("id")

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
func (h *WriteupHandler) GetAllWriteups(c *gin.Context) {
	writeups, err := h.writeupService.GetAllWriteups()
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
func (h *WriteupHandler) DeleteWriteup(c *gin.Context) {
	id := c.Param("id")

	if err := h.writeupService.DeleteWriteup(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Writeup deleted"})
}
