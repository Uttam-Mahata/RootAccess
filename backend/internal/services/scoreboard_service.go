package services

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/go-ctf-platform/backend/internal/database"
	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/repositories"
)

type ScoreboardService struct {
	userRepo       *repositories.UserRepository
	submissionRepo *repositories.SubmissionRepository
	challengeRepo  *repositories.ChallengeRepository
	teamRepo       *repositories.TeamRepository
	contestRepo    *repositories.ContestRepository
}

type UserScore struct {
	Username string `json:"username"`
	Score    int    `json:"score"`
	TeamName string `json:"team_name,omitempty"`
}

type TeamScore struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Score       int       `json:"score"`
	MemberIDs   []string  `json:"member_ids"`
	LeaderID    string    `json:"leader_id,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

func NewScoreboardService(
	userRepo *repositories.UserRepository,
	submissionRepo *repositories.SubmissionRepository,
	challengeRepo *repositories.ChallengeRepository,
	teamRepo *repositories.TeamRepository,
	contestRepo *repositories.ContestRepository,
) *ScoreboardService {
	return &ScoreboardService{
		userRepo:       userRepo,
		submissionRepo: submissionRepo,
		challengeRepo:  challengeRepo,
		teamRepo:       teamRepo,
		contestRepo:    contestRepo,
	}
}

// getFreezeInfo checks contest config and returns freeze time and appropriate cache key suffix
func (s *ScoreboardService) getFreezeInfo() *time.Time {
	if s.contestRepo != nil {
		contest, err := s.contestRepo.GetActiveContest()
		if err == nil && contest != nil && contest.IsScoreboardFrozen() {
			return contest.FreezeTime
		}
	}
	return nil
}

// getCorrectSubmissions returns correct submissions, respecting freeze time if set
func (s *ScoreboardService) getCorrectSubmissions(freezeTime *time.Time) ([]models.Submission, error) {
	if freezeTime != nil {
		return s.submissionRepo.GetCorrectSubmissionsBefore(*freezeTime)
	}
	return s.submissionRepo.GetAllCorrectSubmissions()
}

func (s *ScoreboardService) GetScoreboard() ([]UserScore, error) {
	ctx := context.Background()
	cacheKey := "scoreboard"

	// Check if scoreboard is frozen
	freezeTime := s.getFreezeInfo()
	if freezeTime != nil {
		cacheKey = "scoreboard_frozen"
	}

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

	// Get submissions - use freeze time if scoreboard is frozen
	submissions, err := s.getCorrectSubmissions(freezeTime)
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

	// Sum points for every user's solve
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

	// Map user IDs to Team names
	userTeamMap := make(map[string]string)
	teams, err := s.teamRepo.GetAllTeamsWithScores()
	if err == nil {
		for _, team := range teams {
			for _, mid := range team.MemberIDs {
				userTeamMap[mid.Hex()] = team.Name
			}
		}
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
			TeamName: userTeamMap[uid],
		})
	}

	// Sort scores by score descending
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Score == scores[j].Score {
			return scores[i].Username < scores[j].Username
		}
		return scores[i].Score > scores[j].Score
	})

	// Store in Redis
	if database.RDB != nil {
		data, err := json.Marshal(scores)
		if err == nil {
			err = database.RDB.Set(ctx, cacheKey, data, 1*time.Minute).Err()
		}
	}

	return scores, nil
}

func (s *ScoreboardService) GetTeamScoreboard() ([]TeamScore, error) {
	ctx := context.Background()
	cacheKey := "team_scoreboard"

	// Check if scoreboard is frozen
	freezeTime := s.getFreezeInfo()
	if freezeTime != nil {
		cacheKey = "team_scoreboard_frozen"
	}

	// Try to get from Redis
	if database.RDB != nil {
		val, err := database.RDB.Get(ctx, cacheKey).Result()
		if err == nil {
			var scores []TeamScore
			if err := json.Unmarshal([]byte(val), &scores); err == nil {
				return scores, nil
			}
		}
	}

	// Calculate scores if not in cache
	teams, err := s.teamRepo.GetAllTeamsWithScores() // Just gets the teams
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

	// Get submissions - use freeze time if scoreboard is frozen
	submissions, err := s.getCorrectSubmissions(freezeTime)
	if err != nil {
		return nil, err
	}

	// Map TeamID -> Set of ChallengeIDs solved
	teamSolves := make(map[string]map[string]bool)
	for _, sub := range submissions {
		if sub.TeamID.IsZero() {
			continue
		}
		tid := sub.TeamID.Hex()
		cid := sub.ChallengeID.Hex()
		
		if teamSolves[tid] == nil {
			teamSolves[tid] = make(map[string]bool)
		}
		teamSolves[tid][cid] = true
	}

	var scores []TeamScore
	for _, team := range teams {
		tid := team.ID.Hex()
		totalScore := 0
		
		if solves, exists := teamSolves[tid]; exists {
			for cid := range solves {
				totalScore += challengePoints[cid]
			}
		}

		memberIDs := make([]string, len(team.MemberIDs))
		for i, mid := range team.MemberIDs {
			memberIDs[i] = mid.Hex()
		}

		scores = append(scores, TeamScore{
			ID:          tid,
			Name:        team.Name,
			Description: team.Description,
			Score:       totalScore,
			MemberIDs:   memberIDs,
			LeaderID:    team.LeaderID.Hex(),
			CreatedAt:   team.CreatedAt,
			UpdatedAt:   team.UpdatedAt,
		})
	}

	// Sort scores by score descending
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Score == scores[j].Score {
			return scores[i].Name < scores[j].Name
		}
		return scores[i].Score > scores[j].Score
	})

	// Store in Redis
	if database.RDB != nil {
		data, err := json.Marshal(scores)
		if err == nil {
			err = database.RDB.Set(ctx, cacheKey, data, 1*time.Minute).Err()
		}
	}

	return scores, nil
}