package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/services"
	"github.com/go-ctf-platform/backend/internal/utils"
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

// ImportChallenges imports challenges from JSON array
func (h *BulkChallengeHandler) ImportChallenges(c *gin.Context) {
	var challenges []BulkChallengeImport
	if err := c.ShouldBindJSON(&challenges); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format: " + err.Error()})
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
func (h *BulkChallengeHandler) ExportChallenges(c *gin.Context) {
	challenges, err := h.challengeService.GetAllChallenges()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
func (h *BulkChallengeHandler) DuplicateChallenge(c *gin.Context) {
	id := c.Param("id")
	original, err := h.challengeService.GetChallengeByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Challenge not found"})
		return
	}

	duplicate := &models.Challenge{
		Title:       original.Title + " (Copy)",
		Description: original.Description,
		Category:    original.Category,
		Difficulty:  original.Difficulty,
		MaxPoints:   original.MaxPoints,
		MinPoints:   original.MinPoints,
		Decay:       original.Decay,
		ScoringType: original.ScoringType,
		FlagHash:    original.FlagHash,
		Files:       original.Files,
		Tags:        original.Tags,
		IsPublished: false,
	}

	if err := h.challengeService.CreateChallenge(duplicate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Challenge duplicated successfully"})
}
