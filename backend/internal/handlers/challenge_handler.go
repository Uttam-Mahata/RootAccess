package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChallengeHandler struct {
	challengeService    *services.ChallengeService
	achievementService  *services.AchievementService
	contestService      *services.ContestService
	contestAdminService *services.ContestAdminService
	submissionRepo      *repositories.SubmissionRepository
	userRepo            *repositories.UserRepository
	teamRepo            *repositories.TeamRepository
	wsHub               interface{ BroadcastMessage(string, interface{}) }
}

func NewChallengeHandler(challengeService *services.ChallengeService, achievementService *services.AchievementService, contestService *services.ContestService, wsHub interface{ BroadcastMessage(string, interface{}) }) *ChallengeHandler {
	return &ChallengeHandler{
		challengeService:   challengeService,
		achievementService: achievementService,
		contestService:     contestService,
		wsHub:              wsHub,
	}
}

func NewChallengeHandlerWithRepos(challengeService *services.ChallengeService, achievementService *services.AchievementService, contestService *services.ContestService, contestAdminService *services.ContestAdminService, wsHub interface{ BroadcastMessage(string, interface{}) }, submissionRepo *repositories.SubmissionRepository, userRepo *repositories.UserRepository, teamRepo *repositories.TeamRepository) *ChallengeHandler {
	return &ChallengeHandler{
		challengeService:    challengeService,
		achievementService:  achievementService,
		contestService:     contestService,
		contestAdminService: contestAdminService,
		submissionRepo:     submissionRepo,
		userRepo:           userRepo,
		teamRepo:           teamRepo,
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

// UpdateChallengeRequest is used for updating challenges – flag is optional
type UpdateChallengeRequest struct {
	Title             string        `json:"title" binding:"required"`
	Description       string        `json:"description" binding:"required"`
	DescriptionFormat string        `json:"description_format"`
	Category          string        `json:"category" binding:"required"`
	Difficulty        string        `json:"difficulty" binding:"required"`
	MaxPoints         int           `json:"max_points" binding:"required"`
	MinPoints         int           `json:"min_points" binding:"required"`
	Decay             int           `json:"decay" binding:"required"`
	ScoringType       string        `json:"scoring_type"`
	Flag              string        `json:"flag"` // optional on update
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
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
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
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Challenge created successfully"})
}

func (h *ChallengeHandler) UpdateChallenge(c *gin.Context) {
	id := c.Param("id")
	var req UpdateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
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

	// Only update flag hash if a new flag is provided
	var flagHash string
	if req.Flag != "" {
		flagHash = utils.HashFlag(req.Flag)
	} else {
		// Preserve existing flag hash
		existing, err := h.challengeService.GetChallengeByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Challenge not found"})
			return
		}
		flagHash = existing.FlagHash
	}

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
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Challenge updated successfully"})
}

func (h *ChallengeHandler) DeleteChallenge(c *gin.Context) {
	id := c.Param("id")

	if err := h.challengeService.DeleteChallenge(id); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
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
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
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
	ID                    string   `json:"id"`
	Title                 string   `json:"title"`
	Description           string   `json:"description"`
	DescriptionFormat     string   `json:"description_format"`
	Category              string   `json:"category"`
	Difficulty            string   `json:"difficulty"`
	MaxPoints             int      `json:"max_points"`
	CurrentPoints         int      `json:"current_points"`
	ScoringType           string   `json:"scoring_type"`
	SolveCount            int      `json:"solve_count"`
	Files                 []string `json:"files"`
	Tags                  []string `json:"tags"`
	HintCount             int      `json:"hint_count"`
	IsSolved              bool     `json:"is_solved"`
	OfficialWriteup       string   `json:"official_writeup,omitempty"`
	OfficialWriteupFormat string   `json:"official_writeup_format,omitempty"`
}

// GetAllChallenges returns all challenges for users (filtered by active contest/round visibility)
// @Summary Get all challenges
// @Description Retrieve a list of all published challenges with public details (no flags).
// @Tags Challenges
// @Produce json
// @Success 200 {array} ChallengePublicResponse
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /challenges [get]
func (h *ChallengeHandler) GetAllChallenges(c *gin.Context) {
	// Determine current user and their team
	var userID primitive.ObjectID
	var teamID *primitive.ObjectID
	if userIDStr, exists := c.Get("user_id"); exists {
		userID, _ = primitive.ObjectIDFromHex(userIDStr.(string))
		if h.teamRepo != nil && !userID.IsZero() {
			if team, err := h.teamRepo.FindTeamByMemberID(userID.Hex()); err == nil && team != nil {
				teamID = &team.ID
			}
		}
	}

	var challenges []models.Challenge
	var err error
	if h.contestAdminService != nil {
		challenges, err = h.contestAdminService.GetVisibleChallenges(time.Now(), teamID)
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
			return
		}
		// nil from GetVisibleChallenges means no active contest; challenges is already empty
	} else {
		challenges, err = h.challengeService.GetAllChallenges()
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
			return
		}
	}

	var result []ChallengePublicResponse
	for _, ch := range challenges {
		isSolved := false
		if h.submissionRepo != nil && !userID.IsZero() {
			if sub, _ := h.submissionRepo.FindByChallengeAndUser(ch.ID, userID); sub != nil {
				isSolved = true
			}
			if !isSolved && teamID != nil && !teamID.IsZero() {
				if teamSub, _ := h.submissionRepo.FindByChallengeAndTeam(ch.ID, *teamID); teamSub != nil {
					isSolved = true
				}
			}
		}
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
			IsSolved:          isSolved,
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

	// Enforce visibility for non-admin users: challenge must be in active contest/round and team must be registered
	if h.contestAdminService != nil {
		role, _ := c.Get("role")
		if role != "admin" {
			var teamID *primitive.ObjectID
			if userIDStr, exists := c.Get("user_id"); exists && h.teamRepo != nil {
				userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))
				if !userID.IsZero() {
					if team, err := h.teamRepo.FindTeamByMemberID(userIDStr.(string)); err == nil && team != nil {
						teamID = &team.ID
					}
				}
			}
			visible, err := h.contestAdminService.IsChallengeVisible(id, time.Now(), teamID)
			if err != nil || !visible {
				c.JSON(http.StatusNotFound, gin.H{"error": "Challenge not found"})
				return
			}
		}
	}

	// Determine if the current user (or their team) has already solved this challenge
	isSolved := false
	if userIDStr, exists := c.Get("user_id"); exists && h.submissionRepo != nil {
		userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))
		if !userID.IsZero() {
			if sub, _ := h.submissionRepo.FindByChallengeAndUser(challenge.ID, userID); sub != nil {
				isSolved = true
			}
			if !isSolved && h.teamRepo != nil {
				if team, err := h.teamRepo.FindTeamByMemberID(userID.Hex()); err == nil && team != nil {
					if teamSub, _ := h.submissionRepo.FindByChallengeAndTeam(challenge.ID, team.ID); teamSub != nil {
						isSolved = true
					}
				}
			}
		}
	}

	// Build public response (no flag hash)
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
		IsSolved:          isSolved,
	}
	// Include official writeup only when contest has ended and it is published
	if challenge.OfficialWriteupPublished && challenge.OfficialWriteup != "" {
		ended := true
		if h.contestAdminService != nil {
			ended, _ = h.contestAdminService.HasContestEndedForChallenge(id, time.Now())
		}
		if ended {
			response.OfficialWriteup = challenge.OfficialWriteup
			response.OfficialWriteupFormat = challenge.OfficialWriteupFormat
		}
	}

	c.JSON(http.StatusOK, response)
}

type UpdateOfficialWriteupRequest struct {
	Content string `json:"content" binding:"required"`
	Format  string `json:"format"` // "markdown" or "html"
}

// UpdateOfficialWriteup updates the official writeup for a challenge (admin only)
func (h *ChallengeHandler) UpdateOfficialWriteup(c *gin.Context) {
	id := c.Param("id")
	var req UpdateOfficialWriteupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	format := req.Format
	if format == "" {
		format = "markdown"
	}
	if format != "markdown" && format != "html" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format must be 'markdown' or 'html'"})
		return
	}
	if _, err := h.challengeService.GetChallengeByID(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Challenge not found"})
		return
	}
	if err := h.challengeService.UpdateOfficialWriteup(id, req.Content, format); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Official writeup updated"})
}

// PublishOfficialWriteup publishes the official writeup (admin only, contest must have ended)
func (h *ChallengeHandler) PublishOfficialWriteup(c *gin.Context) {
	id := c.Param("id")
	if _, err := h.challengeService.GetChallengeByID(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Challenge not found"})
		return
	}
	if h.contestAdminService != nil {
		ended, err := h.contestAdminService.HasContestEndedForChallenge(id, time.Now())
		if err != nil || !ended {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot publish official writeup until the contest has ended"})
			return
		}
	}
	if err := h.challengeService.PublishOfficialWriteup(id); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Official writeup published"})
}

type SubmitFlagRequest struct {
	Flag string `json:"flag" binding:"required"`
}

// SubmitFlag submits a flag for a challenge
// @Summary Submit flag
// @Description Submit a flag for a specific challenge. If correct, points are awarded.
// @Tags Challenges
// @Accept json
// @Produce json
// @Param id path string true "Challenge ID"
// @Param request body SubmitFlagRequest true "Flag submission"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 429 {object} map[string]string
// @Security ApiKeyAuth
// @Router /challenges/{id}/submit [post]
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

	// Reject submissions to non-visible challenges (not in active contest/round or team not registered)
	if h.contestAdminService != nil {
		var teamID *primitive.ObjectID
		if h.teamRepo != nil {
			if team, err := h.teamRepo.FindTeamByMemberID(userIDStr.(string)); err == nil && team != nil {
				teamID = &team.ID
			}
		}
		visible, err := h.contestAdminService.IsChallengeVisible(challengeID, time.Now(), teamID)
		if err != nil || !visible {
			c.JSON(http.StatusForbidden, gin.H{"error": "This challenge is not currently available for submissions."})
			return
		}
	}

	// Convert interface{} to string then ObjectID
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	var req SubmitFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	clientIP := c.ClientIP()
	result, err := h.challengeService.SubmitFlag(userID, challengeID, req.Flag, clientIP)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
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

// SolveEntryResponse represents a single solve for the challenge solves endpoint
type SolveEntryResponse struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	TeamID    string    `json:"team_id,omitempty"`
	TeamName  string    `json:"team_name,omitempty"`
	SolvedAt  time.Time `json:"solved_at"`
}

// GetChallengeSolves returns the list of users/teams that solved a challenge
func (h *ChallengeHandler) GetChallengeSolves(c *gin.Context) {
	challengeID := c.Param("id")
	cid, err := primitive.ObjectIDFromHex(challengeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid challenge ID"})
		return
	}

	if h.submissionRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Service not available"})
		return
	}

	submissions, err := h.submissionRepo.GetCorrectSubmissionsByChallenge(cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load solves"})
		return
	}

	// Deduplicate: keep first solve per team (or per user if no team)
	seen := make(map[string]bool)
	var solves []SolveEntryResponse
	for _, sub := range submissions {
		key := sub.UserID.Hex()
		if !sub.TeamID.IsZero() {
			key = "team:" + sub.TeamID.Hex()
		}
		if seen[key] {
			continue
		}
		seen[key] = true

		entry := SolveEntryResponse{
			UserID:   sub.UserID.Hex(),
			SolvedAt: sub.Timestamp,
		}

		// Look up username
		if h.userRepo != nil {
			user, err := h.userRepo.FindByID(sub.UserID.Hex())
			if err == nil && user != nil {
				entry.Username = user.Username
			}
		}

		// Look up team name
		if !sub.TeamID.IsZero() && h.teamRepo != nil {
			team, err := h.teamRepo.FindTeamByID(sub.TeamID.Hex())
			if err == nil && team != nil {
				entry.TeamID = team.ID.Hex()
				entry.TeamName = team.Name
			}
		}

		solves = append(solves, entry)
	}

	c.JSON(http.StatusOK, solves)
}
