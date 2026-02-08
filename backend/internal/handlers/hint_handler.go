package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/services"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if hints == nil {
		hints = []services.HintResponse{}
	}

	c.JSON(http.StatusOK, hints)
}

// RevealHint reveals a specific hint for a challenge
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, hint)
}
