package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/database"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ScoreboardService struct {
	userRepo        *repositories.UserRepository
	submissionRepo  *repositories.SubmissionRepository
	challengeRepo   *repositories.ChallengeRepository
	teamRepo        *repositories.TeamRepository
	contestRepo     *repositories.ContestRepository
	adjustmentRepo  *repositories.ScoreAdjustmentRepository
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
	adjustmentRepo *repositories.ScoreAdjustmentRepository,
) *ScoreboardService {
	return &ScoreboardService{
		userRepo:       userRepo,
		submissionRepo: submissionRepo,
		challengeRepo:  challengeRepo,
		teamRepo:       teamRepo,
		contestRepo:    contestRepo,
		adjustmentRepo: adjustmentRepo,
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

	// Apply manual user score adjustments (if any)
	if s.adjustmentRepo != nil && len(userScores) > 0 {
		userIDs := make([]primitive.ObjectID, 0, len(userScores))
		for uid := range userScores {
			if oid, err := primitive.ObjectIDFromHex(uid); err == nil {
				userIDs = append(userIDs, oid)
			}
		}
		if len(userIDs) > 0 {
			if deltas, err := s.adjustmentRepo.GetAdjustmentsForUsers(userIDs); err == nil {
				for uid, delta := range deltas {
					userScores[uid] += delta
				}
			}
		}
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

	// Apply manual team score adjustments (if any)
	if s.adjustmentRepo != nil && len(scores) > 0 {
		teamIDs := make([]primitive.ObjectID, 0, len(scores))
		for _, ts := range scores {
			if oid, err := primitive.ObjectIDFromHex(ts.ID); err == nil {
				teamIDs = append(teamIDs, oid)
			}
		}
		if len(teamIDs) > 0 {
			if deltas, err := s.adjustmentRepo.GetAdjustmentsForTeams(teamIDs); err == nil {
				for i := range scores {
					if delta, ok := deltas[scores[i].ID]; ok {
						scores[i].Score += delta
					}
				}
			}
		}
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

// TeamScoreProgression represents a team's score at a point in time
type TeamScoreProgression struct {
	TeamID string             `json:"team_id"`
	Name   string             `json:"name"`
	Data   []TimeSeriesScore  `json:"data"`
}

// TimeSeriesScore represents score at a specific time
type TimeSeriesScore struct {
	Date  string `json:"date"`
	Score int    `json:"score"`
}

// GetTeamScoreProgression returns score progression over time for all teams
func (s *ScoreboardService) GetTeamScoreProgression(days int) ([]TeamScoreProgression, error) {
	if days <= 0 {
		days = 30 // Default to 30 days
	}
	if days > 90 {
		days = 90 // Max 90 days
	}

	// Try Redis cache first – this endpoint is used for charts and
	// can be expensive, so we cache it briefly.
	ctx := context.Background()
	cacheKey := fmt.Sprintf("team_score_progression:%d", days)
	if database.RDB != nil {
		if val, err := database.RDB.Get(ctx, cacheKey).Result(); err == nil {
			var cached []TeamScoreProgression
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				return cached, nil
			}
		}
	}

	// Get all teams
	teams, err := s.teamRepo.GetAllTeamsWithScores()
	if err != nil {
		return nil, err
	}

	// Get all challenges
	challenges, err := s.challengeRepo.GetAllChallenges()
	if err != nil {
		return nil, err
	}

	challengePoints := make(map[string]int)
	for _, c := range challenges {
		challengePoints[c.ID.Hex()] = c.CurrentPoints()
	}

	// Get all correct submissions
	submissions, err := s.submissionRepo.GetAllCorrectSubmissions()
	if err != nil {
		return nil, err
	}

	// Group submissions by team and date
	teamSolvesByDate := make(map[string]map[string]map[string]bool)
	
	// Sort submissions by timestamp
	sort.Slice(submissions, func(i, j int) bool {
		return submissions[i].Timestamp.Before(submissions[j].Timestamp)
	})

	since := time.Now().AddDate(0, 0, -days)
	for _, sub := range submissions {
		if sub.TeamID.IsZero() || sub.Timestamp.Before(since) {
			continue
		}
		
		teamID := sub.TeamID.Hex()
		date := sub.Timestamp.Format("2006-01-02")
		challengeID := sub.ChallengeID.Hex()

		if teamSolvesByDate[teamID] == nil {
			teamSolvesByDate[teamID] = make(map[string]map[string]bool)
		}
		if teamSolvesByDate[teamID][date] == nil {
			teamSolvesByDate[teamID][date] = make(map[string]bool)
		}
		teamSolvesByDate[teamID][date][challengeID] = true
	}

	// Build progression for each team
	var progressions []TeamScoreProgression
	for _, team := range teams {
		teamID := team.ID.Hex()
		progression := TeamScoreProgression{
			TeamID: teamID,
			Name:   team.Name,
			Data:   make([]TimeSeriesScore, 0, days),
		}

		// Calculate cumulative score for each day
		cumulativeSolves := make(map[string]bool)
		for i := days - 1; i >= 0; i-- {
			day := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
			
			// Add solves from this day to cumulative
			if daySolves, exists := teamSolvesByDate[teamID][day]; exists {
				for cid := range daySolves {
					cumulativeSolves[cid] = true
				}
			}

			// Calculate score based on cumulative solves (use current challenge points)
			score := 0
			for cid := range cumulativeSolves {
				if points, exists := challengePoints[cid]; exists {
					score += points
				}
			}

			progression.Data = append(progression.Data, TimeSeriesScore{
				Date:  day,
				Score: score,
			})
		}

		progressions = append(progressions, progression)
	}

	// Apply manual team score adjustments as a flat offset across the timeline.
	if s.adjustmentRepo != nil && len(progressions) > 0 {
		teamIDs := make([]primitive.ObjectID, 0, len(progressions))
		for _, p := range progressions {
			if oid, err := primitive.ObjectIDFromHex(p.TeamID); err == nil {
				teamIDs = append(teamIDs, oid)
			}
		}
		if len(teamIDs) > 0 {
			if deltas, err := s.adjustmentRepo.GetAdjustmentsForTeams(teamIDs); err == nil {
				for i := range progressions {
					if delta, ok := deltas[progressions[i].TeamID]; ok && delta != 0 {
						for j := range progressions[i].Data {
							progressions[i].Data[j].Score += delta
						}
					}
				}
			}
		}
	}

	// Sort by current score (highest first)
	sort.Slice(progressions, func(i, j int) bool {
		iScore := 0
		jScore := 0
		if len(progressions[i].Data) > 0 {
			iScore = progressions[i].Data[len(progressions[i].Data)-1].Score
		}
		if len(progressions[j].Data) > 0 {
			jScore = progressions[j].Data[len(progressions[j].Data)-1].Score
		}
		if iScore == jScore {
			return progressions[i].Name < progressions[j].Name
		}
		return iScore > jScore
	})

	// Store in Redis with a short TTL – charts don't need
	// millisecond-level freshness and this keeps tab loads fast.
	if database.RDB != nil {
		if data, err := json.Marshal(progressions); err == nil {
			_ = database.RDB.Set(ctx, cacheKey, data, 1*time.Minute).Err()
		}
	}

	return progressions, nil
}