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
	challengeService   *services.ChallengeService
	achievementService *services.AchievementService
	contestService     *services.ContestService
	wsHub              interface{ BroadcastMessage(string, interface{}) }
}

func NewChallengeHandler(challengeService *services.ChallengeService, achievementService *services.AchievementService, contestService *services.ContestService, wsHub interface{ BroadcastMessage(string, interface{}) }) *ChallengeHandler {
	return &ChallengeHandler{
		challengeService:   challengeService,
		achievementService: achievementService,
		contestService:     contestService,
		wsHub:              wsHub,
	}
}

type CreateChallengeRequest struct {
	Title             string        `json:"title" binding:"required"`
	Description       string        `json:"description" binding:"required"`
	DescriptionFormat string        `json:"description_format"` // "markdown" or "html"
	Category          string        `json:"category" binding:"required"`
	Difficulty        string        `json:"difficulty" binding:"required"`
	MaxPoints         int           `json:"max_points" binding:"required"`
	MinPoints         int           `json:"min_points" binding:"required"`
	Decay             int           `json:"decay" binding:"required"`
	ScoringType       string        `json:"scoring_type"`
	Flag              string        `json:"flag" binding:"required"`
	Files             []string      `json:"files"`
	Tags              []string      `json:"tags"`
	Hints             []HintRequest `json:"hints"`
}

// HintRequest represents a hint in the create/update challenge request
type HintRequest struct {
	Content string `json:"content" binding:"required"`
	Cost    int    `json:"cost" binding:"required"`
	Order   int    `json:"order"`
}

func (h *ChallengeHandler) CreateChallenge(c *gin.Context) {
	var req CreateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate and set default format
	descriptionFormat := req.DescriptionFormat
	if descriptionFormat == "" {
		descriptionFormat = "markdown" // Default to markdown for backward compatibility
	}
	if descriptionFormat != "markdown" && descriptionFormat != "html" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description_format must be 'markdown' or 'html'"})
		return
	}

	// Hash the flag before storing
	flagHash := utils.HashFlag(req.Flag)

	// Set default scoring type
	scoringType := req.ScoringType
	if scoringType == "" {
		scoringType = models.ScoringDynamic
	}

	// Build hints
	var hints []models.Hint
	for i, h := range req.Hints {
		order := h.Order
		if order == 0 {
			order = i + 1
		}
		hints = append(hints, models.Hint{
			ID:      primitive.NewObjectID(),
			Content: h.Content,
			Cost:    h.Cost,
			Order:   order,
		})
	}

	challenge := &models.Challenge{
		Title:             req.Title,
		Description:       req.Description,
		DescriptionFormat: descriptionFormat,
		Category:          req.Category,
		Difficulty:        req.Difficulty,
		MaxPoints:         req.MaxPoints,
		MinPoints:         req.MinPoints,
		Decay:             req.Decay,
		ScoringType:       scoringType,
		FlagHash:          flagHash,
		Files:             req.Files,
		Tags:              req.Tags,
		Hints:             hints,
		IsPublished:       true,
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

	// Validate and set default format
	descriptionFormat := req.DescriptionFormat
	if descriptionFormat == "" {
		descriptionFormat = "markdown" // Default to markdown for backward compatibility
	}
	if descriptionFormat != "markdown" && descriptionFormat != "html" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description_format must be 'markdown' or 'html'"})
		return
	}

	// Hash the flag before storing
	flagHash := utils.HashFlag(req.Flag)

	// Set default scoring type
	scoringType := req.ScoringType
	if scoringType == "" {
		scoringType = models.ScoringDynamic
	}

	// Build hints
	var hints []models.Hint
	for i, hint := range req.Hints {
		order := hint.Order
		if order == 0 {
			order = i + 1
		}
		hints = append(hints, models.Hint{
			ID:      primitive.NewObjectID(),
			Content: hint.Content,
			Cost:    hint.Cost,
			Order:   order,
		})
	}

	challenge := &models.Challenge{
		Title:             req.Title,
		Description:       req.Description,
		DescriptionFormat: descriptionFormat,
		Category:          req.Category,
		Difficulty:        req.Difficulty,
		MaxPoints:         req.MaxPoints,
		MinPoints:         req.MinPoints,
		Decay:             req.Decay,
		ScoringType:       scoringType,
		FlagHash:          flagHash,
		Files:             req.Files,
		Tags:              req.Tags,
		Hints:             hints,
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
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	DescriptionFormat string   `json:"description_format"`
	Category          string   `json:"category"`
	Difficulty        string   `json:"difficulty"`
	MaxPoints         int      `json:"max_points"`
	MinPoints         int      `json:"min_points"`
	Decay             int      `json:"decay"`
	ScoringType       string   `json:"scoring_type"`
	SolveCount        int      `json:"solve_count"`
	CurrentPoints     int      `json:"current_points"`
	Files             []string `json:"files"`
	Tags              []string `json:"tags"`
	HintCount         int      `json:"hint_count"`
}

// GetAllChallengesWithFlags returns all challenges for admin (no flag hash exposed).
// Query param "list=1" returns challenges without description for fast manage-tab load.
func (h *ChallengeHandler) GetAllChallengesWithFlags(c *gin.Context) {
	var challenges []models.Challenge
	var err error
	if c.Query("list") == "1" {
		challenges, err = h.challengeService.GetAllChallengesForList()
	} else {
		challenges, err = h.challengeService.GetAllChallenges()
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []ChallengeAdminResponse
	for _, ch := range challenges {
		result = append(result, ChallengeAdminResponse{
			ID:                ch.ID.Hex(),
			Title:             ch.Title,
			Description:       ch.Description, // empty when list=1 (projection excluded it)
			DescriptionFormat: ch.DescriptionFormat,
			Category:          ch.Category,
			Difficulty:        ch.Difficulty,
			MaxPoints:         ch.MaxPoints,
			MinPoints:         ch.MinPoints,
			Decay:             ch.Decay,
			ScoringType:       ch.ScoringType,
			SolveCount:        ch.SolveCount,
			CurrentPoints:     ch.CurrentPoints(),
			Files:             ch.Files,
			Tags:              ch.Tags,
			HintCount:         len(ch.Hints),
		})
	}

	c.JSON(http.StatusOK, result)
}

// ChallengePublicResponse is the response struct for public challenge view
type ChallengePublicResponse struct {
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	DescriptionFormat string   `json:"description_format"`
	Category          string   `json:"category"`
	Difficulty        string   `json:"difficulty"`
	MaxPoints         int      `json:"max_points"`
	CurrentPoints     int      `json:"current_points"`
	ScoringType       string   `json:"scoring_type"`
	SolveCount        int      `json:"solve_count"`
	Files             []string `json:"files"`
	Tags              []string `json:"tags"`
	HintCount         int      `json:"hint_count"`
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
			ID:                ch.ID.Hex(),
			Title:             ch.Title,
			Description:       ch.Description,
			DescriptionFormat: ch.DescriptionFormat,
			Category:          ch.Category,
			Difficulty:        ch.Difficulty,
			MaxPoints:         ch.MaxPoints,
			CurrentPoints:     ch.CurrentPoints(),
			ScoringType:       ch.ScoringType,
			SolveCount:        ch.SolveCount,
			Files:             ch.Files,
			Tags:              ch.Tags,
			HintCount:         len(ch.Hints),
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
		ID:                challenge.ID.Hex(),
		Title:             challenge.Title,
		Description:       challenge.Description,
		DescriptionFormat: challenge.DescriptionFormat,
		Category:          challenge.Category,
		Difficulty:        challenge.Difficulty,
		MaxPoints:         challenge.MaxPoints,
		CurrentPoints:     challenge.CurrentPoints(),
		ScoringType:       challenge.ScoringType,
		SolveCount:        challenge.SolveCount,
		Files:             challenge.Files,
		Tags:              challenge.Tags,
		HintCount:         len(challenge.Hints),
	}

	c.JSON(http.StatusOK, response)
}

type SubmitFlagRequest struct {
	Flag string `json:"flag" binding:"required"`
}

func (h *ChallengeHandler) SubmitFlag(c *gin.Context) {
	// Check if contest is paused
	if h.contestService != nil {
		status, _, err := h.contestService.GetContestStatus()
		if err == nil && status == models.ContestStatusPaused {
			c.JSON(http.StatusForbidden, gin.H{"error": "Contest is currently paused. Submissions are not accepted."})
			return
		}
	}

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

	clientIP := c.ClientIP()
	result, err := h.challengeService.SubmitFlag(userID, challengeID, req.Flag, clientIP)
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

		// Broadcast solve event via WebSocket
		if h.wsHub != nil {
			challengeObjID, _ := primitive.ObjectIDFromHex(challengeID)
			username, _ := c.Get("username")
			h.wsHub.BroadcastMessage("solve_feed", gin.H{
				"user_id":      userIDStr,
				"username":     username,
				"challenge_id": challengeID,
				"points":       result.Points,
				"solve_count":  result.SolveCount,
				"team_name":    result.TeamName,
			})
			h.wsHub.BroadcastMessage("scoreboard_update", gin.H{
				"updated": true,
			})

			// Check and award achievements
			if h.achievementService != nil {
				teamID := primitive.NilObjectID
				go h.achievementService.CheckAndAwardAchievements(userID, teamID, challengeObjID)
			}
		}
	}

	c.JSON(http.StatusOK, response)
}
