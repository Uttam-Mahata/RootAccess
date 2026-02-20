package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
)

type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
}

func NewAnalyticsHandler(analyticsService *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

// GetPlatformAnalytics returns platform-wide statistics for admin
// @Summary Get platform analytics
// @Description Retrieve global statistics about users, teams, challenges, and submissions.
// @Tags Analytics
// @Produce json
// @Success 200 {object} models.AdminAnalytics
// @Failure 403 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/analytics [get]
func (h *AnalyticsHandler) GetPlatformAnalytics(c *gin.Context) {
	analytics, err := h.analyticsService.GetPlatformAnalytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, analytics)
}
