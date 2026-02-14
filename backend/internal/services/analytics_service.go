package services

import (
	"sort"
	"time"

	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	// Enhanced user statistics
	activeUsers, _ := s.userRepo.CountUsersByStatus("active")
	bannedUsers, _ := s.userRepo.CountUsersByStatus("banned")
	verifiedUsers, _ := s.userRepo.CountVerifiedUsers()
	adminCount, _ := s.userRepo.CountAdmins()

	// New users today and this week
	today := time.Now().Truncate(24 * time.Hour)
	weekAgo := today.AddDate(0, 0, -7)

	newUsersToday := 0
	newUsersThisWeek := 0
	recentUsers, _ := s.userRepo.GetRecentUsers(weekAgo)
	for _, u := range recentUsers {
		if u.CreatedAt.After(today) || u.CreatedAt.Equal(today) {
			newUsersToday++
		}
		newUsersThisWeek++
	}

	// New teams today and this week
	newTeamsToday := 0
	newTeamsThisWeek := 0
	recentTeams, _ := s.teamRepo.GetRecentTeams(weekAgo)
	for _, t := range recentTeams {
		if t.CreatedAt.After(today) || t.CreatedAt.Equal(today) {
			newTeamsToday++
		}
		newTeamsThisWeek++
	}

	// Submissions and solves today
	submissionsToday := 0
	solvesToday := 0
	recentSubs, _ := s.submissionRepo.GetSubmissionsSince(today)
	for _, sub := range recentSubs {
		submissionsToday++
		if sub.IsCorrect {
			solvesToday++
		}
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
	recentActivitySubs, err := s.submissionRepo.GetRecentSubmissions(20)
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

	recentActivity := make([]models.RecentActivityEntry, 0, len(recentActivitySubs))
	for _, sub := range recentActivitySubs {
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

	// User growth over time (last 30 days)
	userDayCounts := make(map[string]int)
	for _, u := range users {
		if u.CreatedAt.After(since) {
			day := u.CreatedAt.Format("2006-01-02")
			userDayCounts[day]++
		}
	}

	userGrowth := make([]models.TimeSeriesEntry, 0, 30)
	for i := 29; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		userGrowth = append(userGrowth, models.TimeSeriesEntry{
			Date:  day,
			Count: userDayCounts[day],
		})
	}

	// Team growth over time (last 30 days)
	allTeams, _ := s.teamRepo.GetAllTeams()
	teamDayCounts := make(map[string]int)
	totalMembers := 0
	for _, t := range allTeams {
		totalMembers += len(t.MemberIDs)
		if t.CreatedAt.After(since) {
			day := t.CreatedAt.Format("2006-01-02")
			teamDayCounts[day]++
		}
	}

	teamGrowth := make([]models.TimeSeriesEntry, 0, 30)
	for i := 29; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		teamGrowth = append(teamGrowth, models.TimeSeriesEntry{
			Date:  day,
			Count: teamDayCounts[day],
		})
	}

	// Calculate average team size
	var avgTeamSize float64
	if len(allTeams) > 0 {
		avgTeamSize = float64(totalMembers) / float64(len(allTeams))
	}

	// Top teams (sorted by score)
	topTeams := make([]models.TeamStats, 0, 10)
	sortedTeams := make([]models.Team, len(allTeams))
	copy(sortedTeams, allTeams)
	sort.Slice(sortedTeams, func(i, j int) bool {
		return sortedTeams[i].Score > sortedTeams[j].Score
	})
	for i, t := range sortedTeams {
		if i >= 10 {
			break
		}
		topTeams = append(topTeams, models.TeamStats{
			TeamID:      t.ID,
			Name:        t.Name,
			Score:       t.Score,
			MemberCount: len(t.MemberIDs),
		})
	}

	// Top users - calculate scores from correct submissions
	userScores := make(map[string]int)
	userSolveCounts := make(map[string]int)
	challengeScores := make(map[string]int)
	for _, c := range challenges {
		challengeScores[c.ID.Hex()] = c.CurrentPoints()
	}

	for _, sub := range allSubmissions {
		if sub.IsCorrect {
			userScores[sub.UserID.Hex()] += challengeScores[sub.ChallengeID.Hex()]
			userSolveCounts[sub.UserID.Hex()]++
		}
	}

	type userScoreEntry struct {
		ID         string
		Username   string
		Score      int
		SolveCount int
	}
	userEntries := make([]userScoreEntry, 0, len(users))
	for _, u := range users {
		userEntries = append(userEntries, userScoreEntry{
			ID:         u.ID.Hex(),
			Username:   u.Username,
			Score:      userScores[u.ID.Hex()],
			SolveCount: userSolveCounts[u.ID.Hex()],
		})
	}
	sort.Slice(userEntries, func(i, j int) bool {
		return userEntries[i].Score > userEntries[j].Score
	})

	topUsers := make([]models.UserStats, 0, 10)
	for i, e := range userEntries {
		if i >= 10 {
			break
		}
		userID, _ := primitive.ObjectIDFromHex(e.ID)
		topUsers = append(topUsers, models.UserStats{
			UserID:     userID,
			Username:   e.Username,
			Score:      e.Score,
			SolveCount: e.SolveCount,
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
		// Enhanced statistics
		ActiveUsers:      int(activeUsers),
		BannedUsers:      int(bannedUsers),
		VerifiedUsers:    int(verifiedUsers),
		AdminCount:       int(adminCount),
		NewUsersToday:    newUsersToday,
		NewUsersThisWeek: newUsersThisWeek,
		NewTeamsToday:    newTeamsToday,
		NewTeamsThisWeek: newTeamsThisWeek,
		SubmissionsToday: submissionsToday,
		SolvesToday:      solvesToday,
		AverageTeamSize:  avgTeamSize,
		UserGrowth:       userGrowth,
		TeamGrowth:       teamGrowth,
		TopTeams:         topTeams,
		TopUsers:         topUsers,
	}, nil
}
