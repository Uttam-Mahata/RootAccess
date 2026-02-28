package handlers

import (
	"fmt"
	"net/http"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type BulkChallengeHandler struct {
	challengeService *services.ChallengeService
}

func NewBulkChallengeHandler(challengeService *services.ChallengeService) *BulkChallengeHandler {
	return &BulkChallengeHandler{challengeService: challengeService}
}

type BulkChallengeImport struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Difficulty  string   `json:"difficulty"`
	MaxPoints   int      `json:"max_points"`
	MinPoints   int      `json:"min_points"`
	Decay       int      `json:"decay"`
	ScoringType string   `json:"scoring_type"`
	Flag        string   `json:"flag"`
	Files       []string `json:"files"`
	Tags        []string `json:"tags"`
}

type ExportChallenge struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Difficulty  string   `json:"difficulty"`
	MaxPoints   int      `json:"max_points"`
	MinPoints   int      `json:"min_points"`
	Decay       int      `json:"decay"`
	ScoringType string   `json:"scoring_type"`
	Files       []string `json:"files"`
	Tags        []string `json:"tags"`
}

// ImportChallenges imports challenges from JSON array
// @Summary Import challenges in bulk
// @Description Create multiple challenges at once from a JSON array.
// @Tags Admin Challenges
// @Accept json
// @Produce json
// @Param request body []BulkChallengeImport true "List of challenges to import"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/challenges/import [post]
func (h *BulkChallengeHandler) ImportChallenges(c *gin.Context) {
	var challenges []BulkChallengeImport
	if err := c.ShouldBindJSON(&challenges); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}

	imported := 0
	var errors []string
	for i, ch := range challenges {
		if ch.Title == "" || ch.Flag == "" {
			errors = append(errors, fmt.Sprintf("Challenge %d: missing title or flag", i))
			continue
		}
		if ch.MaxPoints <= 0 {
			errors = append(errors, fmt.Sprintf("Challenge %d (%s): max_points must be positive", i, ch.Title))
			continue
		}
		if ch.MinPoints < 0 {
			errors = append(errors, fmt.Sprintf("Challenge %d (%s): min_points must be non-negative", i, ch.Title))
			continue
		}
		if ch.MinPoints > ch.MaxPoints {
			errors = append(errors, fmt.Sprintf("Challenge %d (%s): min_points cannot exceed max_points", i, ch.Title))
			continue
		}
		scoringType := ch.ScoringType
		if scoringType == "" {
			scoringType = models.ScoringDynamic
		}
		challenge := &models.Challenge{
			Title:       ch.Title,
			Description: ch.Description,
			Category:    ch.Category,
			Difficulty:  ch.Difficulty,
			MaxPoints:   ch.MaxPoints,
			MinPoints:   ch.MinPoints,
			Decay:       ch.Decay,
			ScoringType: scoringType,
			FlagHash:    utils.HashFlag(ch.Flag),
			Files:       ch.Files,
			Tags:        ch.Tags,
			IsPublished: true,
		}
		if err := h.challengeService.CreateChallenge(challenge); err != nil {
			errors = append(errors, fmt.Sprintf("Challenge %d (%s): %s", i, ch.Title, err.Error()))
			continue
		}
		imported++
	}

	c.JSON(http.StatusOK, gin.H{
		"imported": imported,
		"errors":   errors,
		"total":    len(challenges),
	})
}

// ExportChallenges exports all challenges as JSON
// @Summary Export all challenges
// @Description Retrieve all challenges in a format suitable for bulk import.
// @Tags Admin Challenges
// @Produce json
// @Success 200 {array} ExportChallenge
// @Security ApiKeyAuth
// @Router /admin/challenges/export [get]
func (h *BulkChallengeHandler) ExportChallenges(c *gin.Context) {
	challenges, err := h.challengeService.GetAllChallenges()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	var result []ExportChallenge
	for _, ch := range challenges {
		result = append(result, ExportChallenge{
			Title:       ch.Title,
			Description: ch.Description,
			Category:    ch.Category,
			Difficulty:  ch.Difficulty,
			MaxPoints:   ch.MaxPoints,
			MinPoints:   ch.MinPoints,
			Decay:       ch.Decay,
			ScoringType: ch.ScoringType,
			Files:       ch.Files,
			Tags:        ch.Tags,
		})
	}

	c.Header("Content-Disposition", "attachment; filename=challenges.json")
	c.JSON(http.StatusOK, result)
}

// DuplicateChallenge clones a challenge
// @Summary Duplicate a challenge
// @Description Create a copy of an existing challenge.
// @Tags Admin Challenges
// @Produce json
// @Param id path string true "Challenge ID"
// @Success 201 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/challenges/{id}/duplicate [post]
func (h *BulkChallengeHandler) DuplicateChallenge(c *gin.Context) {
	id := c.Param("id")
	original, err := h.challengeService.GetChallengeByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Challenge not found"})
		return
	}

	duplicate := &models.Challenge{
		Title:             original.Title + " (Copy)",
		Description:       original.Description,
		DescriptionFormat: original.DescriptionFormat,
		Category:          original.Category,
		Difficulty:        original.Difficulty,
		MaxPoints:         original.MaxPoints,
		MinPoints:         original.MinPoints,
		Decay:             original.Decay,
		ScoringType:       original.ScoringType,
		FlagHash:          original.FlagHash,
		Files:             original.Files,
		Tags:              original.Tags,
		Hints:             original.Hints,
		IsPublished:       original.IsPublished,
	}

	if err := h.challengeService.CreateChallenge(duplicate); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Challenge duplicated successfully"})
}
