package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
}

func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// CreateNotificationRequest represents the request body for creating a notification
type CreateNotificationRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Type    string `json:"type" binding:"required"` // info, warning, success, error
}

// CreateNotification handles creating a new notification (admin only)
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get admin user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	notification, err := h.notificationService.CreateNotification(req.Title, req.Content, req.Type, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Notification created successfully",
		"notification": notification,
	})
}

// GetActiveNotifications returns active notifications for users
func (h *NotificationHandler) GetActiveNotifications(c *gin.Context) {
	notifications, err := h.notificationService.GetActiveNotifications()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if notifications == nil {
		notifications = []models.Notification{}
	}

	c.JSON(http.StatusOK, notifications)
}

// GetAllNotifications returns all notifications for admin
func (h *NotificationHandler) GetAllNotifications(c *gin.Context) {
	notifications, err := h.notificationService.GetAllNotifications()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if notifications == nil {
		notifications = []models.Notification{}
	}

	c.JSON(http.StatusOK, notifications)
}

// UpdateNotificationRequest represents the request body for updating a notification
type UpdateNotificationRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Type     string `json:"type" binding:"required"`
	IsActive bool   `json:"is_active"`
}

// UpdateNotification handles updating a notification (admin only)
func (h *NotificationHandler) UpdateNotification(c *gin.Context) {
	id := c.Param("id")

	var req UpdateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.notificationService.UpdateNotification(id, req.Title, req.Content, req.Type, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification updated successfully"})
}

// DeleteNotification handles deleting a notification (admin only)
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	id := c.Param("id")

	err := h.notificationService.DeleteNotification(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// ToggleNotificationActive handles toggling notification active status (admin only)
func (h *NotificationHandler) ToggleNotificationActive(c *gin.Context) {
	id := c.Param("id")

	err := h.notificationService.ToggleNotificationActive(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification status toggled successfully"})
}
