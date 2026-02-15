package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProfileHandler struct {
	userRepo       *repositories.UserRepository
	submissionRepo *repositories.SubmissionRepository
	challengeRepo  *repositories.ChallengeRepository
	teamRepo       *repositories.TeamRepository
}

func NewProfileHandler(
	userRepo *repositories.UserRepository,
	submissionRepo *repositories.SubmissionRepository,
	challengeRepo *repositories.ChallengeRepository,
	teamRepo *repositories.TeamRepository,
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
	Bio               string            `json:"bio,omitempty"`
	Website           string            `json:"website,omitempty"`
	SocialLinks       *models.SocialLinks `json:"social_links,omitempty"`
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
		Bio:              user.Bio,
		Website:          user.Website,
		SocialLinks:      user.SocialLinks,
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

// UpdateProfileRequest represents the request body for profile update
type UpdateProfileRequest struct {
	Bio         string             `json:"bio"`
	Website     string             `json:"website"`
	SocialLinks *models.SocialLinks `json:"social_links"`
}

// UpdateMyProfile updates the authenticated user's profile
func (h *ProfileHandler) UpdateMyProfile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate bio length
	if len(req.Bio) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bio must be 500 characters or less"})
		return
	}

	// Validate website length
	if len(req.Website) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Website must be 200 characters or less"})
		return
	}

	err = h.userRepo.UpdateProfile(userID, req.Bio, req.Website, req.SocialLinks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}
