package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChallengeHandler struct {
	challengeService *services.ChallengeService
}

func NewChallengeHandler(challengeService *services.ChallengeService) *ChallengeHandler {
	return &ChallengeHandler{
		challengeService: challengeService,
	}
}

type CreateChallengeRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description" binding:"required"`
	Category    string   `json:"category" binding:"required"`
	Points      int      `json:"points" binding:"required"`
	Flag        string   `json:"flag" binding:"required"`
	Files       []string `json:"files"`
}

func (h *ChallengeHandler) CreateChallenge(c *gin.Context) {
	var req CreateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	challenge := &models.Challenge{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Points:      req.Points,
		Flag:        req.Flag,
		Files:       req.Files,
	}

	if err := h.challengeService.CreateChallenge(challenge); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Challenge created successfully"})
}

func (h *ChallengeHandler) GetAllChallenges(c *gin.Context) {
	challenges, err := h.challengeService.GetAllChallenges()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, challenges)
}

func (h *ChallengeHandler) GetChallengeByID(c *gin.Context) {
	id := c.Param("id")
	challenge, err := h.challengeService.GetChallengeByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Challenge not found"})
		return
	}

	c.JSON(http.StatusOK, challenge)
}

type SubmitFlagRequest struct {
	Flag string `json:"flag" binding:"required"`
}

func (h *ChallengeHandler) SubmitFlag(c *gin.Context) {
	challengeID := c.Param("id")
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Convert interface{} to string then ObjectID
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	var req SubmitFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isCorrect, err := h.challengeService.SubmitFlag(userID, challengeID, req.Flag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if isCorrect {
		c.JSON(http.StatusOK, gin.H{"message": "Flag correct!", "correct": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Flag incorrect", "correct": false})
	}
}
