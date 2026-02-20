package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HintHandler struct {
	hintService *services.HintService
}

func NewHintHandler(hintService *services.HintService) *HintHandler {
	return &HintHandler{
		hintService: hintService,
	}
}

// GetHints returns hints for a challenge with reveal status
// @Summary Get hints for a challenge
// @Description Retrieve hints for a specific challenge, including whether they have been revealed by the user or their team.
// @Tags Challenges
// @Produce json
// @Param id path string true "Challenge ID"
// @Success 200 {array} services.HintResponse
// @Failure 401 {object} map[string]string
// @Security ApiKeyAuth
// @Router /challenges/{id}/hints [get]
func (h *HintHandler) GetHints(c *gin.Context) {
	challengeID := c.Param("id")
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	hints, err := h.hintService.GetHintsForChallenge(challengeID, userID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	if hints == nil {
		hints = []services.HintResponse{}
	}

	c.JSON(http.StatusOK, hints)
}

// RevealHint reveals a specific hint for a challenge
// @Summary Reveal a hint
// @Description Deduct points and reveal the content of a specific hint.
// @Tags Challenges
// @Produce json
// @Param id path string true "Challenge ID"
// @Param hintId path string true "Hint ID"
// @Success 200 {object} services.HintResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security ApiKeyAuth
// @Router /challenges/{id}/hints/{hintId}/reveal [post]
func (h *HintHandler) RevealHint(c *gin.Context) {
	challengeID := c.Param("id")
	hintID := c.Param("hintId")
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	hint, err := h.hintService.RevealHint(challengeID, hintID, userID)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, hint)
}
