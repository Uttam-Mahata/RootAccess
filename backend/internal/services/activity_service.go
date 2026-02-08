package services

import (
	"sort"
	"time"

	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type solveInfo struct {
	entry    models.SolveEntry
	solvedAt time.Time
}

type ActivityService struct {
	userRepo        *repositories.UserRepository
	submissionRepo  *repositories.SubmissionRepository
	challengeRepo   *repositories.ChallengeRepository
	achievementRepo *repositories.AchievementRepository
	teamRepo        *repositories.TeamRepository
}

func NewActivityService(
	userRepo *repositories.UserRepository,
	submissionRepo *repositories.SubmissionRepository,
	challengeRepo *repositories.ChallengeRepository,
	achievementRepo *repositories.AchievementRepository,
	teamRepo *repositories.TeamRepository,
) *ActivityService {
	return &ActivityService{
		userRepo:        userRepo,
		submissionRepo:  submissionRepo,
		challengeRepo:   challengeRepo,
		achievementRepo: achievementRepo,
		teamRepo:        teamRepo,
	}
}

func (s *ActivityService) GetUserActivity(userID primitive.ObjectID) (*models.UserActivity, error) {
	user, err := s.userRepo.FindByID(userID.Hex())
	if err != nil {
		return nil, err
	}

	correctSubs, err := s.submissionRepo.GetUserCorrectSubmissions(userID)
	if err != nil {
		return nil, err
	}

	challenges, err := s.challengeRepo.GetAllChallenges()
	if err != nil {
		return nil, err
	}

	challengeMap := make(map[string]*models.Challenge)
	for i := range challenges {
		challengeMap[challenges[i].ID.Hex()] = &challenges[i]
	}

	// Calculate total points and build recent solves
	totalPoints := 0
	solvedByCategory := make(map[string]int)
	pointsByCategory := make(map[string]int)

	solves := make([]solveInfo, 0, len(correctSubs))

	for _, sub := range correctSubs {
		c, ok := challengeMap[sub.ChallengeID.Hex()]
		if !ok {
			continue
		}
		points := c.CurrentPoints()
		totalPoints += points
		solvedByCategory[c.Category]++
		pointsByCategory[c.Category] += points

		solves = append(solves, solveInfo{
			entry: models.SolveEntry{
				ChallengeID:    c.ID,
				ChallengeTitle: c.Title,
				Category:       c.Category,
				Points:         points,
				SolvedAt:       sub.Timestamp,
			},
			solvedAt: sub.Timestamp,
		})
	}

	// Sort solves by time descending
	sort.Slice(solves, func(i, j int) bool {
		return solves[i].solvedAt.After(solves[j].solvedAt)
	})

	// Recent solves (last 20)
	recentLimit := 20
	if len(solves) < recentLimit {
		recentLimit = len(solves)
	}
	recentSolves := make([]models.SolveEntry, recentLimit)
	for i := 0; i < recentLimit; i++ {
		recentSolves[i] = solves[i].entry
	}

	// Category progress
	totalByCategory := make(map[string]int)
	for _, c := range challenges {
		totalByCategory[c.Category]++
	}

	categoryProgress := make(map[string]models.CategoryStat)
	for cat, total := range totalByCategory {
		categoryProgress[cat] = models.CategoryStat{
			Total:  total,
			Solved: solvedByCategory[cat],
			Points: pointsByCategory[cat],
		}
	}

	// Achievements
	achievements, err := s.achievementRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Solve streak (consecutive days with a solve, ending today or yesterday)
	solveStreak := calculateSolveStreak(solves)

	return &models.UserActivity{
		UserID:           userID,
		Username:         user.Username,
		TotalSolves:      len(correctSubs),
		TotalPoints:      totalPoints,
		CategoryProgress: categoryProgress,
		RecentSolves:     recentSolves,
		Achievements:     achievements,
		SolveStreak:      solveStreak,
	}, nil
}

func calculateSolveStreak(solves []solveInfo) int {
	if len(solves) == 0 {
		return 0
	}

	// Collect unique solve dates
	solveDates := make(map[string]bool)
	for _, s := range solves {
		day := s.solvedAt.Format("2006-01-02")
		solveDates[day] = true
	}

	// Start from today and count consecutive days
	today := time.Now().Truncate(24 * time.Hour)
	todayStr := today.Format("2006-01-02")
	yesterdayStr := today.AddDate(0, 0, -1).Format("2006-01-02")

	// Streak must include today or yesterday
	if !solveDates[todayStr] && !solveDates[yesterdayStr] {
		return 0
	}

	startDay := today
	if !solveDates[todayStr] {
		startDay = today.AddDate(0, 0, -1)
	}

	streak := 0
	for {
		dayStr := startDay.AddDate(0, 0, -streak).Format("2006-01-02")
		if !solveDates[dayStr] {
			break
		}
		streak++
	}

	return streak
}
