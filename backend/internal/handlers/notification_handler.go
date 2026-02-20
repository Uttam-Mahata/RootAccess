package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/utils"
	wsHub "github.com/Uttam-Mahata/RootAccess/backend/internal/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
	hub                *wsHub.Hub
}

func NewNotificationHandler(notificationService *services.NotificationService, hub *wsHub.Hub) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		hub:                hub,
	}
}

// CreateNotificationRequest represents the request body for creating a notification
type CreateNotificationRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Type    string `json:"type" binding:"required"` // info, warning, success, error
}

// CreateNotification handles creating a new notification (admin only)
// @Summary Create notification
// @Description Create a new platform-wide notification.
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Param request body CreateNotificationRequest true "Notification details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/notifications [post]
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
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
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Push new active notification to all connected clients immediately
	if notification.IsActive {
		h.hub.BroadcastMessage("notification:created", notification)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Notification created successfully",
		"notification": notification,
	})
}

// GetActiveNotifications returns active notifications for users
// @Summary Get active notifications
// @Description Retrieve a list of all currently active platform-wide notifications.
// @Tags Notifications
// @Produce json
// @Success 200 {array} models.Notification
// @Router /notifications [get]
func (h *NotificationHandler) GetActiveNotifications(c *gin.Context) {
	notifications, err := h.notificationService.GetActiveNotifications()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	if notifications == nil {
		notifications = []models.Notification{}
	}

	c.JSON(http.StatusOK, notifications)
}

// GetAllNotifications returns all notifications for admin
// @Summary Get all notifications
// @Description Retrieve all notifications, including inactive ones.
// @Tags Admin Notifications
// @Produce json
// @Success 200 {array} models.Notification
// @Security ApiKeyAuth
// @Router /admin/notifications [get]
func (h *NotificationHandler) GetAllNotifications(c *gin.Context) {
	notifications, err := h.notificationService.GetAllNotifications()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
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
// @Summary Update notification
// @Description Update an existing notification's details or status.
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Param request body UpdateNotificationRequest true "Updated notification details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/notifications/{id} [put]
func (h *NotificationHandler) UpdateNotification(c *gin.Context) {
	id := c.Param("id")

	var req UpdateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	err := h.notificationService.UpdateNotification(id, req.Title, req.Content, req.Type, req.IsActive)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	// Broadcast updated notification so clients can add/remove/update in real time
	if updated, fetchErr := h.notificationService.GetNotificationByID(id); fetchErr == nil {
		h.hub.BroadcastMessage("notification:updated", updated)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification updated successfully"})
}

// DeleteNotification handles deleting a notification (admin only)
// @Summary Delete notification
// @Description Permanently delete a notification.
// @Tags Admin Notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	id := c.Param("id")

	err := h.notificationService.DeleteNotification(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	// Notify clients to remove this notification immediately
	h.hub.BroadcastMessage("notification:deleted", gin.H{"id": id})

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// ToggleNotificationActive handles toggling notification active status (admin only)
// @Summary Toggle notification status
// @Description Switch a notification between active and inactive.
// @Tags Admin Notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/notifications/{id}/toggle [post]
func (h *NotificationHandler) ToggleNotificationActive(c *gin.Context) {
	id := c.Param("id")

	err := h.notificationService.ToggleNotificationActive(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	// Push toggled state: clients add or remove the notification based on is_active
	if updated, fetchErr := h.notificationService.GetNotificationByID(id); fetchErr == nil {
		h.hub.BroadcastMessage("notification:updated", updated)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification status toggled successfully"})
}
