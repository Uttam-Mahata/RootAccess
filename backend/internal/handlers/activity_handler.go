package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActivityHandler struct {
	activityService *services.ActivityService
}

func NewActivityHandler(activityService *services.ActivityService) *ActivityHandler {
	return &ActivityHandler{activityService: activityService}
}

// GetMyActivity returns activity stats for the current user
// @Summary Get user activity
// @Description Retrieve activity statistics, solve history, and progress for the authenticated user.
// @Tags Activity
// @Produce json
// @Success 200 {object} models.UserActivity
// @Failure 401 {object} map[string]string
// @Security ApiKeyAuth
// @Router /activity/me [get]
func (h *ActivityHandler) GetMyActivity(c *gin.Context) {
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
	activity, err := h.activityService.GetUserActivity(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, activity)
}
