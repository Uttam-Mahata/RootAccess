package services

import (
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
)

type AchievementService struct {
	achievementRepo *repositories.AchievementRepository
	submissionRepo  *repositories.SubmissionRepository
	challengeRepo   *repositories.ChallengeRepository
}

func NewAchievementService(
	achievementRepo *repositories.AchievementRepository,
	submissionRepo *repositories.SubmissionRepository,
	challengeRepo *repositories.ChallengeRepository,
) *AchievementService {
	return &AchievementService{
		achievementRepo: achievementRepo,
		submissionRepo:  submissionRepo,
		challengeRepo:   challengeRepo,
	}
}

// CheckAndAwardAchievements checks after a solve if any new achievements should be awarded
func (s *AchievementService) CheckAndAwardAchievements(userID, teamID string, challengeID string) {
	challenge, err := s.challengeRepo.GetChallengeByID(challengeID)
	if err != nil {
		return
	}

	now := time.Now()

	s.checkFirstBlood(userID, teamID, challenge)
	s.checkCategoryMaster(userID, teamID, challenge.Category)
	s.checkNightOwl(userID, teamID, challengeID, now)
	s.checkSpeedDemon(userID, teamID, now)
}

// GetUserAchievements returns all achievements for a user
func (s *AchievementService) GetUserAchievements(userID string) ([]models.Achievement, error) {
	return s.achievementRepo.GetByUserID(userID)
}

// GetTeamAchievements returns all achievements for a team
func (s *AchievementService) GetTeamAchievements(teamID string) ([]models.Achievement, error) {
	return s.achievementRepo.GetByTeamID(teamID)
}

// checkFirstBlood awards the First Blood achievement if this solve was the first for the challenge
func (s *AchievementService) checkFirstBlood(userID, teamID string, challenge *models.Challenge) {
	if challenge.SolveCount != 1 {
		return
	}

	exists, err := s.achievementRepo.Exists(userID, models.AchievementFirstBlood)
	if err != nil || exists {
		return
	}

	achievement := &models.Achievement{
		UserID:      userID,
		TeamID:      teamID,
		Type:        models.AchievementFirstBlood,
		Name:        "First Blood",
		Description: "First solver of challenge: " + challenge.Title,
		Icon:        "🩸",
		ChallengeID: challenge.ID,
	}
	s.achievementRepo.Create(achievement)
}

// checkCategoryMaster awards the Category Master achievement if user has solved all challenges in a category
func (s *AchievementService) checkCategoryMaster(userID, teamID string, category string) {
	exists, err := s.achievementRepo.ExistsForCategory(userID, models.AchievementCategoryMaster, category)
	if err != nil || exists {
		return
	}

	allChallenges, err := s.challengeRepo.GetAllChallenges()
	if err != nil {
		return
	}

	var categoryChallenges []models.Challenge
	for _, c := range allChallenges {
		if c.Category == category {
			categoryChallenges = append(categoryChallenges, c)
		}
	}

	if len(categoryChallenges) == 0 {
		return
	}

	submissions, err := s.submissionRepo.GetUserCorrectSubmissions(userID)
	if err != nil {
		return
	}

	solvedSet := make(map[string]bool)
	for _, sub := range submissions {
		solvedSet[sub.ChallengeID] = true
	}

	for _, c := range categoryChallenges {
		if !solvedSet[c.ID] {
			return
		}
	}

	achievement := &models.Achievement{
		UserID:      userID,
		TeamID:      teamID,
		Type:        models.AchievementCategoryMaster,
		Name:        "Category Master",
		Description: "Solved all challenges in category: " + category,
		Icon:        "👑",
		Category:    category,
	}
	s.achievementRepo.Create(achievement)
}

// checkNightOwl awards the Night Owl achievement if the solve happened between midnight and 5 AM
func (s *AchievementService) checkNightOwl(userID, teamID string, challengeID string, solveTime time.Time) {
	hour := solveTime.Hour()
	if hour >= 5 {
		return
	}

	exists, err := s.achievementRepo.Exists(userID, models.AchievementNightOwl)
	if err != nil || exists {
		return
	}

	achievement := &models.Achievement{
		UserID:      userID,
		TeamID:      teamID,
		Type:        models.AchievementNightOwl,
		Name:        "Night Owl",
		Description: "Solved a challenge between midnight and 5 AM",
		Icon:        "🦉",
		ChallengeID: challengeID,
	}
	s.achievementRepo.Create(achievement)
}

// checkSpeedDemon awards the Speed Demon achievement if the user solved 5+ challenges in a single day
func (s *AchievementService) checkSpeedDemon(userID, teamID string, solveTime time.Time) {
	exists, err := s.achievementRepo.Exists(userID, models.AchievementSpeedDemon)
	if err != nil || exists {
		return
	}

	submissions, err := s.submissionRepo.GetUserCorrectSubmissions(userID)
	if err != nil {
		return
	}

	year, month, day := solveTime.Date()
	loc := solveTime.Location()
	dayStart := time.Date(year, month, day, 0, 0, 0, 0, loc)
	dayEnd := dayStart.Add(24 * time.Hour)

	count := 0
	for _, sub := range submissions {
		if !sub.Timestamp.Before(dayStart) && sub.Timestamp.Before(dayEnd) {
			count++
		}
	}

	if count < 5 {
		return
	}

	achievement := &models.Achievement{
		UserID:      userID,
		TeamID:      teamID,
		Type:        models.AchievementSpeedDemon,
		Name:        "Speed Demon",
		Description: "Solved 5 or more challenges in a single day",
		Icon:        "⚡",
	}
	s.achievementRepo.Create(achievement)
}
