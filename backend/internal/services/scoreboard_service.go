package services

import (
	"github.com/go-ctf-platform/backend/internal/repositories"
)

type ScoreboardService struct {
	userRepo       *repositories.UserRepository
	submissionRepo *repositories.SubmissionRepository
	challengeRepo  *repositories.ChallengeRepository
}

type UserScore struct {
	Username string `json:"username"`
	Score    int    `json:"score"`
}

func NewScoreboardService(userRepo *repositories.UserRepository, submissionRepo *repositories.SubmissionRepository, challengeRepo *repositories.ChallengeRepository) *ScoreboardService {
	return &ScoreboardService{
		userRepo:       userRepo,
		submissionRepo: submissionRepo,
		challengeRepo:  challengeRepo,
	}
}

func (s *ScoreboardService) GetScoreboard() ([]UserScore, error) {
	submissions, err := s.submissionRepo.GetAllCorrectSubmissions()
	if err != nil {
		return nil, err
	}

	challenges, err := s.challengeRepo.GetAllChallenges()
	if err != nil {
		return nil, err
	}

	challengePoints := make(map[string]int)
	for _, c := range challenges {
		challengePoints[c.ID.Hex()] = c.Points
	}

	userScores := make(map[string]int)

	// This is not efficient for large datasets but works for MVP
	// Better approach: Aggregation pipeline in MongoDB
	for _, sub := range submissions {
		userID := sub.UserID.Hex()
		points := challengePoints[sub.ChallengeID.Hex()]
		userScores[userID] += points
	}

	// Fetch all users to map ID to Username
	users, err := s.userRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]string)
	for _, u := range users {
		userMap[u.ID.Hex()] = u.Username
	}

	var scores []UserScore
	for uid, score := range userScores {
		username, exists := userMap[uid]
		if !exists {
			username = "Unknown"
		}
		scores = append(scores, UserScore{
			Username: username,
			Score:    score,
		})
	}

	return scores, nil
}
