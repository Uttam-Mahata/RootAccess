package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/services"
	"github.com/go-ctf-platform/backend/internal/utils"
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
	Difficulty  string   `json:"difficulty" binding:"required"`
	MaxPoints   int      `json:"max_points" binding:"required"`
	MinPoints   int      `json:"min_points" binding:"required"`
	Decay       int      `json:"decay" binding:"required"`
	Flag        string   `json:"flag" binding:"required"`
	Files       []string `json:"files"`
}

func (h *ChallengeHandler) CreateChallenge(c *gin.Context) {
	var req CreateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the flag before storing
	flagHash := utils.HashFlag(req.Flag)

	challenge := &models.Challenge{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Difficulty:  req.Difficulty,
		MaxPoints:   req.MaxPoints,
		MinPoints:   req.MinPoints,
		Decay:       req.Decay,
		FlagHash:    flagHash,
		Files:       req.Files,
	}

	if err := h.challengeService.CreateChallenge(challenge); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Challenge created successfully"})
}

func (h *ChallengeHandler) UpdateChallenge(c *gin.Context) {
	id := c.Param("id")
	var req CreateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the flag before storing
	flagHash := utils.HashFlag(req.Flag)

	challenge := &models.Challenge{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Difficulty:  req.Difficulty,
		MaxPoints:   req.MaxPoints,
		MinPoints:   req.MinPoints,
		Decay:       req.Decay,
		FlagHash:    flagHash,
		Files:       req.Files,
	}

	if err := h.challengeService.UpdateChallenge(id, challenge); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Challenge updated successfully"})
}

func (h *ChallengeHandler) DeleteChallenge(c *gin.Context) {
	id := c.Param("id")

	if err := h.challengeService.DeleteChallenge(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Challenge deleted successfully"})
}

// ChallengeResponse is the response struct for challenges (for admin view)
type ChallengeAdminResponse struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Difficulty    string   `json:"difficulty"`
	MaxPoints     int      `json:"max_points"`
	MinPoints     int      `json:"min_points"`
	Decay         int      `json:"decay"`
	SolveCount    int      `json:"solve_count"`
	CurrentPoints int      `json:"current_points"`
	Files         []string `json:"files"`
}

// GetAllChallengesWithFlags returns all challenges for admin (no flag hash exposed)
func (h *ChallengeHandler) GetAllChallengesWithFlags(c *gin.Context) {
	challenges, err := h.challengeService.GetAllChallenges()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []ChallengeAdminResponse
	for _, ch := range challenges {
		result = append(result, ChallengeAdminResponse{
			ID:            ch.ID.Hex(),
			Title:         ch.Title,
			Description:   ch.Description,
			Category:      ch.Category,
			Difficulty:    ch.Difficulty,
			MaxPoints:     ch.MaxPoints,
			MinPoints:     ch.MinPoints,
			Decay:         ch.Decay,
			SolveCount:    ch.SolveCount,
			CurrentPoints: ch.CurrentPoints(),
			Files:         ch.Files,
		})
	}

	c.JSON(http.StatusOK, result)
}

// ChallengePublicResponse is the response struct for public challenge view
type ChallengePublicResponse struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Difficulty    string   `json:"difficulty"`
	MaxPoints     int      `json:"max_points"`
	CurrentPoints int      `json:"current_points"`
	SolveCount    int      `json:"solve_count"`
	Files         []string `json:"files"`
}

func (h *ChallengeHandler) GetAllChallenges(c *gin.Context) {
	challenges, err := h.challengeService.GetAllChallenges()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []ChallengePublicResponse
	for _, ch := range challenges {
		result = append(result, ChallengePublicResponse{
			ID:            ch.ID.Hex(),
			Title:         ch.Title,
			Description:   ch.Description,
			Category:      ch.Category,
			Difficulty:    ch.Difficulty,
			MaxPoints:     ch.MaxPoints,
			CurrentPoints: ch.CurrentPoints(),
			SolveCount:    ch.SolveCount,
			Files:         ch.Files,
		})
	}

	c.JSON(http.StatusOK, result)
}

func (h *ChallengeHandler) GetChallengeByID(c *gin.Context) {
	id := c.Param("id")
	challenge, err := h.challengeService.GetChallengeByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Challenge not found"})
		return
	}

	// Return public response (no flag hash)
	response := ChallengePublicResponse{
		ID:            challenge.ID.Hex(),
		Title:         challenge.Title,
		Description:   challenge.Description,
		Category:      challenge.Category,
		Difficulty:    challenge.Difficulty,
		MaxPoints:     challenge.MaxPoints,
		CurrentPoints: challenge.CurrentPoints(),
		SolveCount:    challenge.SolveCount,
		Files:         challenge.Files,
	}

	c.JSON(http.StatusOK, response)
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

	result, err := h.challengeService.SubmitFlag(userID, challengeID, req.Flag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"correct":        result.IsCorrect,
		"already_solved": result.AlreadySolved,
	}

	if result.IsCorrect {
		response["message"] = result.Message
		if response["message"] == "" {
			response["message"] = "Flag correct!"
		}
		response["points"] = result.Points
		response["solve_count"] = result.SolveCount
		if result.TeamName != "" {
			response["team_name"] = result.TeamName
		}
	} else {
		response["message"] = "Flag incorrect"
	}

	c.JSON(http.StatusOK, response)
}
