package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-ctf-platform/backend/internal/database"
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
	ctx := context.Background()
	cacheKey := "scoreboard"

	// Try to get from Redis
	if database.RDB != nil {
		val, err := database.RDB.Get(ctx, cacheKey).Result()
		if err == nil {
			var scores []UserScore
			if err := json.Unmarshal([]byte(val), &scores); err == nil {
				return scores, nil
			}
		}
	}

	// Calculate scores if not in cache
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
		challengePoints[c.ID.Hex()] = c.CurrentPoints()
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

	// Store in Redis
	if database.RDB != nil {
		data, err := json.Marshal(scores)
		if err == nil {
			err = database.RDB.Set(ctx, cacheKey, data, 1*time.Minute).Err() // Cache for 1 minute
			if err != nil {
				log.Printf("Failed to cache scoreboard: %v", err)
			}
		}
	}

	return scores, nil
}