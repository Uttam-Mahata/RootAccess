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
)

type ScoreboardService struct {
	userRepo           *repositories.UserRepository
	submissionRepo     *repositories.SubmissionRepository
	challengeRepo      *repositories.ChallengeRepository
	teamRepo           *repositories.TeamRepository
	contestRepo        *repositories.ContestRepository
	adjustmentRepo     *repositories.ScoreAdjustmentRepository
	contestEntityRepo  *repositories.ContestEntityRepository
	contestRoundRepo   *repositories.ContestRoundRepository
	roundChallengeRepo *repositories.RoundChallengeRepository
	registrationRepo   *repositories.TeamContestRegistrationRepository
	contestSolveRepo   *repositories.ContestSolveRepository
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
	contestEntityRepo *repositories.ContestEntityRepository,
	contestRoundRepo *repositories.ContestRoundRepository,
	roundChallengeRepo *repositories.RoundChallengeRepository,
	registrationRepo *repositories.TeamContestRegistrationRepository,
	contestSolveRepo *repositories.ContestSolveRepository,
) *ScoreboardService {
	return &ScoreboardService{
		userRepo:           userRepo,
		submissionRepo:     submissionRepo,
		challengeRepo:      challengeRepo,
		teamRepo:           teamRepo,
		contestRepo:        contestRepo,
		adjustmentRepo:     adjustmentRepo,
		contestEntityRepo:  contestEntityRepo,
		contestRoundRepo:   contestRoundRepo,
		roundChallengeRepo: roundChallengeRepo,
		registrationRepo:   registrationRepo,
		contestSolveRepo:   contestSolveRepo,
	}
}

// getContestChallengeIDs returns the set of challenge IDs belonging to a contest (via its rounds)
func (s *ScoreboardService) getContestChallengeIDs(contestID string) (map[string]bool, error) {
	rounds, err := s.contestRoundRepo.ListByContestID(contestID)
	if err != nil {
		return nil, err
	}
	if len(rounds) == 0 {
		return map[string]bool{}, nil
	}

	roundIDs := make([]string, len(rounds))
	for i := range rounds {
		roundIDs[i] = rounds[i].ID
	}

	challengeOIDs, err := s.roundChallengeRepo.GetChallengeIDsForRounds(roundIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool, len(challengeOIDs))
	for _, oid := range challengeOIDs {
		result[oid] = true
	}
	return result, nil
}

// getContestTeamIDs returns the set of team IDs registered for a contest
func (s *ScoreboardService) getContestTeamIDs(contestID string) (map[string]bool, error) {
	contestOID := contestID
	if contestOID == "" {
		return nil, nil
	}

	teamOIDs, err := s.registrationRepo.GetContestTeams(contestOID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool, len(teamOIDs))
	for _, oid := range teamOIDs {
		result[oid] = true
	}
	return result, nil
}

// getFreezeInfoForContest checks per-contest freeze time
func (s *ScoreboardService) getFreezeInfoForContest(contestID string) *time.Time {
	if s.contestEntityRepo == nil {
		return nil
	}
	contest, err := s.contestEntityRepo.FindByID(contestID)
	if err != nil || contest == nil {
		return nil
	}
	now := time.Now()
	if contest.IsScoreboardFrozen(now) && contest.FreezeTime != "" {
		t, _ := time.Parse(time.RFC3339, contest.FreezeTime)
		return &t
	}
	return nil
}

// getCorrectSubmissions returns correct submissions, respecting freeze time if set.
// When contestID is provided, only returns submissions for that contest.
func (s *ScoreboardService) getCorrectSubmissions(freezeTime *time.Time) ([]models.Submission, error) {
	if freezeTime != nil {
		return s.submissionRepo.GetCorrectSubmissionsBefore(*freezeTime)
	}
	return s.submissionRepo.GetAllCorrectSubmissions()
}

// getCorrectSubmissionsForContest returns correct submissions scoped to a specific contest,
// respecting freeze time if set.
func (s *ScoreboardService) getCorrectSubmissionsForContest(contestID string, freezeTime *time.Time) ([]models.Submission, error) {
	if freezeTime != nil {
		return s.submissionRepo.GetCorrectSubmissionsByContestBefore(contestID, *freezeTime)
	}
	return s.submissionRepo.GetCorrectSubmissionsByContest(contestID)
}

// GetScoreboard returns the individual scoreboard for a specific contest.
// If contestID is empty, returns an empty slice.
func (s *ScoreboardService) GetScoreboard(contestID string) ([]UserScore, error) {
	if contestID == "" {
		return []UserScore{}, nil
	}

	ctx := context.Background()
	freezeTime := s.getFreezeInfoForContest(contestID)
	cacheKey := fmt.Sprintf("scoreboard:%s", contestID)
	if freezeTime != nil {
		cacheKey = fmt.Sprintf("scoreboard:%s:frozen", contestID)
	}

	// Try Redis cache
	if database.Registry != nil && database.Registry.Scoreboard != nil {
		val, err := database.Registry.Scoreboard.Get(ctx, cacheKey).Result()
		if err == nil {
			var scores []UserScore
			if err := json.Unmarshal([]byte(val), &scores); err == nil {
				return scores, nil
			}
		}
	}

	// Get contest challenge IDs
	contestChallenges, err := s.getContestChallengeIDs(contestID)
	if err != nil {
		return nil, err
	}

	// Get registered team IDs
	contestTeams, err := s.getContestTeamIDs(contestID)
	if err != nil {
		return nil, err
	}

	// Build user-to-team membership map (to filter users to registered teams)
	allTeams, err := s.teamRepo.GetAllTeamsWithScores()
	if err != nil {
		return nil, err
	}

	registeredUserIDs := make(map[string]bool)
	userTeamMap := make(map[string]string)
	for _, team := range allTeams {
		tid := team.ID
		if !contestTeams[tid] {
			continue
		}
		members, _ := s.teamRepo.GetTeamMembers(team.ID)
		for _, mid := range members {
			uid := mid
			registeredUserIDs[uid] = true
			userTeamMap[uid] = team.Name
		}
	}

	// Get all challenges for point calculation
	challenges, err := s.challengeRepo.GetAllChallenges()
	if err != nil {
		return nil, err
	}

	// Use contest-specific solve counts for dynamic scoring
	var contestSolveCounts map[string]int
	if s.contestSolveRepo != nil {
		contestSolveCounts, _ = s.contestSolveRepo.GetContestSolveCounts(contestID)
	}

	challengePoints := make(map[string]int)
	for _, c := range challenges {
		if contestChallenges[c.ID] {
			if contestSolveCounts != nil {
				challengePoints[c.ID] = c.PointsForSolveCount(contestSolveCounts[c.ID])
			} else {
				challengePoints[c.ID] = c.CurrentPoints()
			}
		}
	}

	// Get submissions scoped to this contest
	submissions, err := s.getCorrectSubmissionsForContest(contestID, freezeTime)
	if err != nil {
		return nil, err
	}

	userScores := make(map[string]int)
	for _, sub := range submissions {
		userID := sub.UserID
		challengeID := sub.ChallengeID

		// Only count if user belongs to a registered team AND challenge is in contest
		if !registeredUserIDs[userID] {
			continue
		}
		points, inContest := challengePoints[challengeID]
		if !inContest {
			continue
		}
		userScores[userID] += points
	}

	// Apply manual user score adjustments
	if s.adjustmentRepo != nil && len(userScores) > 0 {
		userIDs := make([]string, 0, len(userScores))
		for uid := range userScores {
			if oid := uid; err == nil {
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

	usernameMap := make(map[string]string)
	for _, u := range users {
		usernameMap[u.ID] = u.Username
	}

	var scores []UserScore
	for uid, score := range userScores {
		username, exists := usernameMap[uid]
		if !exists {
			username = "Unknown"
		}
		scores = append(scores, UserScore{
			Username: username,
			Score:    score,
			TeamName: userTeamMap[uid],
		})
	}

	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Score == scores[j].Score {
			return scores[i].Username < scores[j].Username
		}
		return scores[i].Score > scores[j].Score
	})

	// Cache in Redis
	if database.Registry != nil && database.Registry.Scoreboard != nil {
		data, err := json.Marshal(scores)
		if err == nil {
			_ = database.Registry.Scoreboard.Set(ctx, cacheKey, data, 5*time.Minute).Err()
		}
	}

	return scores, nil
}

// GetTeamScoreboard returns the team scoreboard for a specific contest.
// If contestID is empty, returns an empty slice.
func (s *ScoreboardService) GetTeamScoreboard(contestID string) ([]TeamScore, error) {
	if contestID == "" {
		return []TeamScore{}, nil
	}

	ctx := context.Background()
	freezeTime := s.getFreezeInfoForContest(contestID)
	cacheKey := fmt.Sprintf("team_scoreboard:%s", contestID)
	if freezeTime != nil {
		cacheKey = fmt.Sprintf("team_scoreboard:%s:frozen", contestID)
	}

	// Try Redis cache
	if database.Registry != nil && database.Registry.Scoreboard != nil {
		val, err := database.Registry.Scoreboard.Get(ctx, cacheKey).Result()
		if err == nil {
			var scores []TeamScore
			if err := json.Unmarshal([]byte(val), &scores); err == nil {
				return scores, nil
			}
		}
	}

	// Get contest challenge IDs
	contestChallenges, err := s.getContestChallengeIDs(contestID)
	if err != nil {
		return nil, err
	}

	// Get registered team IDs
	contestTeams, err := s.getContestTeamIDs(contestID)
	if err != nil {
		return nil, err
	}

	// Get all teams
	allTeams, err := s.teamRepo.GetAllTeamsWithScores()
	if err != nil {
		return nil, err
	}

	// Get all challenges for point values
	challenges, err := s.challengeRepo.GetAllChallenges()
	if err != nil {
		return nil, err
	}

	// Use contest-specific solve counts for dynamic scoring
	var contestSolveCountsTeam map[string]int
	if s.contestSolveRepo != nil {
		contestSolveCountsTeam, _ = s.contestSolveRepo.GetContestSolveCounts(contestID)
	}

	challengePoints := make(map[string]int)
	for _, c := range challenges {
		if contestChallenges[c.ID] {
			if contestSolveCountsTeam != nil {
				challengePoints[c.ID] = c.PointsForSolveCount(contestSolveCountsTeam[c.ID])
			} else {
				challengePoints[c.ID] = c.CurrentPoints()
			}
		}
	}

	// Get submissions scoped to this contest
	submissions, err := s.getCorrectSubmissionsForContest(contestID, freezeTime)
	if err != nil {
		return nil, err
	}

	// Map TeamID -> Set of ChallengeIDs solved (only contest challenges)
	teamSolves := make(map[string]map[string]bool)
	for _, sub := range submissions {
		if sub.TeamID == "" {
			continue
		}
		tid := sub.TeamID
		cid := sub.ChallengeID

		if !contestTeams[tid] || !contestChallenges[cid] {
			continue
		}

		if teamSolves[tid] == nil {
			teamSolves[tid] = make(map[string]bool)
		}
		teamSolves[tid][cid] = true
	}

	var scores []TeamScore
	for _, team := range allTeams {
		tid := team.ID
		if !contestTeams[tid] {
			continue
		}

		totalScore := 0
		if solves, exists := teamSolves[tid]; exists {
			for cid := range solves {
				totalScore += challengePoints[cid]
			}
		}

		memberIDs, _ := s.teamRepo.GetTeamMembers(team.ID)
		for i, mid := range memberIDs {
			memberIDs[i] = mid
		}

		scores = append(scores, TeamScore{
			ID:          tid,
			Name:        team.Name,
			Description: team.Description,
			Score:       totalScore,
			MemberIDs:   memberIDs,
			LeaderID:    team.LeaderID,
			CreatedAt:   team.CreatedAt,
			UpdatedAt:   team.UpdatedAt,
		})
	}

	// Apply manual team score adjustments
	if s.adjustmentRepo != nil && len(scores) > 0 {
		teamIDs := make([]string, 0, len(scores))
		for _, ts := range scores {
			oid := ts.ID
			if true {
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

	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Score == scores[j].Score {
			return scores[i].Name < scores[j].Name
		}
		return scores[i].Score > scores[j].Score
	})

	// Cache in Redis
	if database.Registry != nil && database.Registry.Scoreboard != nil {
		data, err := json.Marshal(scores)
		if err == nil {
			_ = database.Registry.Scoreboard.Set(ctx, cacheKey, data, 5*time.Minute).Err()
		}
	}

	return scores, nil
}

// TeamScoreProgression represents a team's score at a point in time
type TeamScoreProgression struct {
	TeamID string            `json:"team_id"`
	Name   string            `json:"name"`
	Data   []TimeSeriesScore `json:"data"`
}

// TimeSeriesScore represents score at a specific time
type TimeSeriesScore struct {
	Date  string `json:"date"`
	Score int    `json:"score"`
}

// GetTeamScoreProgression returns score progression over time for teams in a specific contest.
// If contestID is empty, returns an empty slice.
func (s *ScoreboardService) GetTeamScoreProgression(days int, contestID string) ([]TeamScoreProgression, error) {
	if contestID == "" {
		return []TeamScoreProgression{}, nil
	}

	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("team_score_progression:%s:%d", contestID, days)
	if database.Registry != nil && database.Registry.Scoreboard != nil {
		if val, err := database.Registry.Scoreboard.Get(ctx, cacheKey).Result(); err == nil {
			var cached []TeamScoreProgression
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				return cached, nil
			}
		}
	}

	// Get contest challenge IDs
	contestChallenges, err := s.getContestChallengeIDs(contestID)
	if err != nil {
		return nil, err
	}

	// Get registered team IDs
	contestTeams, err := s.getContestTeamIDs(contestID)
	if err != nil {
		return nil, err
	}

	// Get all teams (filter to registered)
	allTeams, err := s.teamRepo.GetAllTeamsWithScores()
	if err != nil {
		return nil, err
	}

	// Get all challenges
	challenges, err := s.challengeRepo.GetAllChallenges()
	if err != nil {
		return nil, err
	}

	// Use contest-specific solve counts for dynamic scoring
	var contestSolveCountsProg map[string]int
	if s.contestSolveRepo != nil {
		contestSolveCountsProg, _ = s.contestSolveRepo.GetContestSolveCounts(contestID)
	}

	challengePoints := make(map[string]int)
	for _, c := range challenges {
		if contestChallenges[c.ID] {
			if contestSolveCountsProg != nil {
				challengePoints[c.ID] = c.PointsForSolveCount(contestSolveCountsProg[c.ID])
			} else {
				challengePoints[c.ID] = c.CurrentPoints()
			}
		}
	}

	// Get correct submissions scoped to this contest
	submissions, err := s.submissionRepo.GetCorrectSubmissionsByContest(contestID)
	if err != nil {
		return nil, err
	}

	sort.Slice(submissions, func(i, j int) bool {
		return submissions[i].Timestamp.Before(submissions[j].Timestamp)
	})

	// Group submissions by team and date (only contest teams + contest challenges)
	teamSolvesByDate := make(map[string]map[string]map[string]bool)
	since := time.Now().AddDate(0, 0, -days)
	for _, sub := range submissions {
		if sub.TeamID == "" || sub.Timestamp.Before(since) {
			continue
		}

		teamID := sub.TeamID
		challengeID := sub.ChallengeID

		if !contestTeams[teamID] || !contestChallenges[challengeID] {
			continue
		}

		date := sub.Timestamp.Format("2006-01-02")
		if teamSolvesByDate[teamID] == nil {
			teamSolvesByDate[teamID] = make(map[string]map[string]bool)
		}
		if teamSolvesByDate[teamID][date] == nil {
			teamSolvesByDate[teamID][date] = make(map[string]bool)
		}
		teamSolvesByDate[teamID][date][challengeID] = true
	}

	// Build progression for registered teams
	var progressions []TeamScoreProgression
	for _, team := range allTeams {
		teamID := team.ID
		if !contestTeams[teamID] {
			continue
		}

		progression := TeamScoreProgression{
			TeamID: teamID,
			Name:   team.Name,
			Data:   make([]TimeSeriesScore, 0, days),
		}

		cumulativeSolves := make(map[string]bool)
		for i := days - 1; i >= 0; i-- {
			day := time.Now().AddDate(0, 0, -i).Format("2006-01-02")

			if daySolves, exists := teamSolvesByDate[teamID][day]; exists {
				for cid := range daySolves {
					cumulativeSolves[cid] = true
				}
			}

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

	// Apply manual team score adjustments as a flat offset
	if s.adjustmentRepo != nil && len(progressions) > 0 {
		teamIDs := make([]string, 0, len(progressions))
		for _, p := range progressions {
			oid := p.TeamID
			if true {
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

	if database.Registry != nil && database.Registry.Scoreboard != nil {
		if data, err := json.Marshal(progressions); err == nil {
			_ = database.Registry.Scoreboard.Set(ctx, cacheKey, data, 5*time.Minute).Err()
		}
	}

	return progressions, nil
}
