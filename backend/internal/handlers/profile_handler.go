package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories/interfaces"
)

type ProfileHandler struct {
	userRepo       interfaces.UserRepository
	submissionRepo interfaces.SubmissionRepository
	challengeRepo  interfaces.ChallengeRepository
	teamRepo       interfaces.TeamRepository
}

func NewProfileHandler(
	userRepo interfaces.UserRepository,
	submissionRepo interfaces.SubmissionRepository,
	challengeRepo interfaces.ChallengeRepository,
	teamRepo interfaces.TeamRepository,
) *ProfileHandler {
	return &ProfileHandler{
		userRepo:       userRepo,
		submissionRepo: submissionRepo,
		challengeRepo:  challengeRepo,
		teamRepo:       teamRepo,
	}
}

// SolvedChallenge represents a challenge solved by the user
type SolvedChallenge struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Category      string `json:"category"`
	Difficulty    string `json:"difficulty"`
	Points        int    `json:"points"`
	SolvedAt      string `json:"solved_at"`
}

// CategoryStats represents solve statistics by category
type CategoryStats struct {
	Category    string `json:"category"`
	SolveCount  int    `json:"solve_count"`
	TotalPoints int    `json:"total_points"`
}

// UserProfileResponse represents the user profile data
type UserProfileResponse struct {
	Username          string            `json:"username"`
	JoinedAt          string            `json:"joined_at"`
	TeamID            string            `json:"team_id,omitempty"`
	TeamName          string            `json:"team_name,omitempty"`
	TotalPoints       int               `json:"total_points"`
	SolveCount        int               `json:"solve_count"`
	TotalSubmissions  int64             `json:"total_submissions"`
	SolvedChallenges  []SolvedChallenge `json:"solved_challenges"`
	CategoryStats     []CategoryStats   `json:"category_stats"`
}

// GetUserProfile returns the public profile of a user by username
// @Summary Get user profile
// @Description Retrieve the public profile, solve history, and statistics of a user by their username.
// @Tags Profiles
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} UserProfileResponse
// @Failure 404 {object} map[string]string
// @Router /users/{username}/profile [get]
func (h *ProfileHandler) GetUserProfile(c *gin.Context) {
	username := c.Param("username")

	// Find user by username
	user, err := h.userRepo.FindByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get all challenges for point lookup
	challenges, err := h.challengeRepo.GetAllChallenges()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch challenges"})
		return
	}

	// Create a map of challenge ID to challenge
	challengeMap := make(map[string]models.Challenge)
	for _, ch := range challenges {
		challengeMap[ch.ID.Hex()] = ch
	}

	// Get user's correct submissions
	submissions, err := h.submissionRepo.GetUserCorrectSubmissions(user.ID)
	if err != nil {
		submissions = []models.Submission{} // Empty if error
	}

	// Get total submission count
	totalSubmissions, _ := h.submissionRepo.GetUserSubmissionCount(user.ID)

	// Build solved challenges list and calculate stats
	var solvedChallenges []SolvedChallenge
	categoryStatsMap := make(map[string]*CategoryStats)
	totalPoints := 0
	seenChallenges := make(map[string]bool) // To avoid duplicates

	for _, sub := range submissions {
		challengeID := sub.ChallengeID.Hex()
		if seenChallenges[challengeID] {
			continue
		}
		seenChallenges[challengeID] = true

		challenge, exists := challengeMap[challengeID]
		if !exists {
			continue
		}

		points := challenge.CurrentPoints()
		totalPoints += points

		solvedChallenges = append(solvedChallenges, SolvedChallenge{
			ID:         challengeID,
			Title:      challenge.Title,
			Category:   challenge.Category,
			Difficulty: challenge.Difficulty,
			Points:     points,
			SolvedAt:   sub.Timestamp.Format("2006-01-02T15:04:05Z"),
		})

		// Update category stats
		if _, exists := categoryStatsMap[challenge.Category]; !exists {
			categoryStatsMap[challenge.Category] = &CategoryStats{
				Category: challenge.Category,
			}
		}
		categoryStatsMap[challenge.Category].SolveCount++
		categoryStatsMap[challenge.Category].TotalPoints += points
	}

	// Convert category stats map to slice
	var categoryStats []CategoryStats
	for _, stats := range categoryStatsMap {
		categoryStats = append(categoryStats, *stats)
	}

	// Check if user is in a team
	teamID := ""
	teamName := ""
	team, _ := h.teamRepo.FindTeamByMemberID(user.ID.Hex())
	if team != nil {
		teamID = team.ID.Hex()
		teamName = team.Name
	}

	profile := UserProfileResponse{
		Username:         user.Username,
		JoinedAt:         user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		TeamID:           teamID,
		TeamName:         teamName,
		TotalPoints:      totalPoints,
		SolveCount:       len(solvedChallenges),
		TotalSubmissions: totalSubmissions,
		SolvedChallenges: solvedChallenges,
		CategoryStats:    categoryStats,
	}

	c.JSON(http.StatusOK, profile)
}
