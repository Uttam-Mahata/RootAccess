package services

import (
	"sort"
	"time"

	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/repositories"
)

type AnalyticsService struct {
	userRepo       *repositories.UserRepository
	submissionRepo *repositories.SubmissionRepository
	challengeRepo  *repositories.ChallengeRepository
	teamRepo       *repositories.TeamRepository
}

func NewAnalyticsService(
	userRepo *repositories.UserRepository,
	submissionRepo *repositories.SubmissionRepository,
	challengeRepo *repositories.ChallengeRepository,
	teamRepo *repositories.TeamRepository,
) *AnalyticsService {
	return &AnalyticsService{
		userRepo:       userRepo,
		submissionRepo: submissionRepo,
		challengeRepo:  challengeRepo,
		teamRepo:       teamRepo,
	}
}

func (s *AnalyticsService) GetPlatformAnalytics() (*models.AdminAnalytics, error) {
	totalUsers, err := s.userRepo.CountUsers()
	if err != nil {
		return nil, err
	}

	totalTeams, err := s.teamRepo.CountTeams()
	if err != nil {
		return nil, err
	}

	totalChallenges, err := s.challengeRepo.CountChallenges()
	if err != nil {
		return nil, err
	}

	totalSubmissions, err := s.submissionRepo.CountSubmissions()
	if err != nil {
		return nil, err
	}

	totalCorrect, err := s.submissionRepo.CountCorrectSubmissions()
	if err != nil {
		return nil, err
	}

	var successRate float64
	if totalSubmissions > 0 {
		successRate = float64(totalCorrect) / float64(totalSubmissions) * 100
	}

	// Challenge popularity
	challenges, err := s.challengeRepo.GetAllChallenges()
	if err != nil {
		return nil, err
	}

	allSubmissions, err := s.submissionRepo.GetAllSubmissions()
	if err != nil {
		return nil, err
	}

	// Count attempts per challenge
	attemptCounts := make(map[string]int)
	for _, sub := range allSubmissions {
		attemptCounts[sub.ChallengeID.Hex()]++
	}

	challengePopularity := make([]models.ChallengePopularity, 0, len(challenges))
	categoryBreakdown := make(map[string]int)
	difficultyBreakdown := make(map[string]int)

	for _, c := range challenges {
		attempts := attemptCounts[c.ID.Hex()]
		var rate float64
		if attempts > 0 {
			rate = float64(c.SolveCount) / float64(attempts) * 100
		}
		challengePopularity = append(challengePopularity, models.ChallengePopularity{
			ChallengeID:  c.ID,
			Title:        c.Title,
			Category:     c.Category,
			SolveCount:   c.SolveCount,
			AttemptCount: attempts,
			SuccessRate:  rate,
		})

		categoryBreakdown[c.Category]++
		difficultyBreakdown[c.Difficulty]++
	}

	// Sort popularity by solve count descending
	sort.Slice(challengePopularity, func(i, j int) bool {
		return challengePopularity[i].SolveCount > challengePopularity[j].SolveCount
	})

	// Recent activity (last 20 submissions)
	recentSubs, err := s.submissionRepo.GetRecentSubmissions(20)
	if err != nil {
		return nil, err
	}

	users, err := s.userRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}
	userMap := make(map[string]string)
	for _, u := range users {
		userMap[u.ID.Hex()] = u.Username
	}

	challengeMap := make(map[string]string)
	for _, c := range challenges {
		challengeMap[c.ID.Hex()] = c.Title
	}

	recentActivity := make([]models.RecentActivityEntry, 0, len(recentSubs))
	for _, sub := range recentSubs {
		action := "attempted"
		if sub.IsCorrect {
			action = "solved"
		}
		recentActivity = append(recentActivity, models.RecentActivityEntry{
			UserID:        sub.UserID,
			Username:      userMap[sub.UserID.Hex()],
			Action:        action,
			ChallengeID:   sub.ChallengeID,
			ChallengeName: challengeMap[sub.ChallengeID.Hex()],
			Timestamp:     sub.Timestamp,
		})
	}

	// Solves over time (last 30 days)
	since := time.Now().AddDate(0, 0, -30)
	correctSince, err := s.submissionRepo.GetCorrectSubmissionsSince(since)
	if err != nil {
		return nil, err
	}

	dayCounts := make(map[string]int)
	for _, sub := range correctSince {
		day := sub.Timestamp.Format("2006-01-02")
		dayCounts[day]++
	}

	solvesOverTime := make([]models.TimeSeriesEntry, 0, 30)
	for i := 29; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		solvesOverTime = append(solvesOverTime, models.TimeSeriesEntry{
			Date:  day,
			Count: dayCounts[day],
		})
	}

	return &models.AdminAnalytics{
		TotalUsers:          int(totalUsers),
		TotalTeams:          int(totalTeams),
		TotalChallenges:     int(totalChallenges),
		TotalSubmissions:    int(totalSubmissions),
		TotalCorrect:        int(totalCorrect),
		SuccessRate:         successRate,
		ChallengePopularity: challengePopularity,
		CategoryBreakdown:   categoryBreakdown,
		DifficultyBreakdown: difficultyBreakdown,
		RecentActivity:      recentActivity,
		SolvesOverTime:      solvesOverTime,
	}, nil
}
