package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementHandler struct {
	achievementService *services.AchievementService
}

func NewAchievementHandler(achievementService *services.AchievementService) *AchievementHandler {
	return &AchievementHandler{achievementService: achievementService}
}

// GetMyAchievements returns achievements for the current user
// @Summary Get user achievements
// @Description Retrieve all achievements earned by the authenticated user.
// @Tags Achievements
// @Produce json
// @Success 200 {array} models.Achievement
// @Failure 401 {object} map[string]string
// @Security ApiKeyAuth
// @Router /achievements/me [get]
func (h *AchievementHandler) GetMyAchievements(c *gin.Context) {
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
	achievements, err := h.achievementService.GetUserAchievements(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, achievements)
}
