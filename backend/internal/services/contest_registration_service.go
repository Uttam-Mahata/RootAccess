package services

import (
	"errors"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ContestRegistrationService struct {
	contestEntityRepo      *repositories.ContestEntityRepository
	registrationRepo       *repositories.TeamContestRegistrationRepository
	teamRepo               *repositories.TeamRepository
}

func NewContestRegistrationService(
	contestEntityRepo *repositories.ContestEntityRepository,
	registrationRepo *repositories.TeamContestRegistrationRepository,
	teamRepo *repositories.TeamRepository,
) *ContestRegistrationService {
	return &ContestRegistrationService{
		contestEntityRepo: contestEntityRepo,
		registrationRepo:  registrationRepo,
		teamRepo:          teamRepo,
	}
}

// GetUpcomingContests returns contests that haven't started yet
func (s *ContestRegistrationService) GetUpcomingContests() ([]models.Contest, error) {
	allContests, err := s.contestEntityRepo.ListAll()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	upcoming := []models.Contest{} // Initialize as empty slice, not nil
	for _, contest := range allContests {
		if contest.StartTime.After(now) {
			upcoming = append(upcoming, contest)
		}
	}

	return upcoming, nil
}

// RegisterTeamForContest registers a team for a contest
func (s *ContestRegistrationService) RegisterTeamForContest(teamID, contestID string) error {
	teamOID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return errors.New("invalid team ID")
	}

	contestOID, err := primitive.ObjectIDFromHex(contestID)
	if err != nil {
		return errors.New("invalid contest ID")
	}

	// Verify contest exists
	contest, err := s.contestEntityRepo.FindByID(contestID)
	if err != nil {
		return errors.New("contest not found")
	}

	// Check if contest has already started
	now := time.Now()
	if !contest.StartTime.After(now) {
		return errors.New("cannot register for a contest that has already started")
	}

	// Verify team exists
	_, err = s.teamRepo.FindTeamByID(teamID)
	if err != nil {
		return errors.New("team not found")
	}

	return s.registrationRepo.RegisterTeam(teamOID, contestOID)
}

// UnregisterTeamFromContest unregisters a team from a contest
func (s *ContestRegistrationService) UnregisterTeamFromContest(teamID, contestID string) error {
	teamOID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return errors.New("invalid team ID")
	}

	contestOID, err := primitive.ObjectIDFromHex(contestID)
	if err != nil {
		return errors.New("invalid contest ID")
	}

	// Check if contest has already started
	contest, err := s.contestEntityRepo.FindByID(contestID)
	if err != nil {
		return errors.New("contest not found")
	}

	now := time.Now()
	if !contest.StartTime.After(now) {
		return errors.New("cannot unregister from a contest that has already started")
	}

	return s.registrationRepo.UnregisterTeam(teamOID, contestOID)
}

// IsTeamRegistered checks if a team is registered for a contest
func (s *ContestRegistrationService) IsTeamRegistered(teamID, contestID string) (bool, error) {
	teamOID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return false, errors.New("invalid team ID")
	}

	contestOID, err := primitive.ObjectIDFromHex(contestID)
	if err != nil {
		return false, errors.New("invalid contest ID")
	}

	return s.registrationRepo.IsTeamRegistered(teamOID, contestOID)
}

// GetRegisteredTeamsCount returns the number of teams registered for a contest
func (s *ContestRegistrationService) GetRegisteredTeamsCount(contestID string) (int64, error) {
	contestOID, err := primitive.ObjectIDFromHex(contestID)
	if err != nil {
		return 0, errors.New("invalid contest ID")
	}
	return s.registrationRepo.CountContestTeams(contestOID)
}

// GetTeamRegisteredContests returns all contests a team is registered for
func (s *ContestRegistrationService) GetTeamRegisteredContests(teamID string) ([]models.Contest, error) {
	teamOID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, errors.New("invalid team ID")
	}

	contestIDs, err := s.registrationRepo.GetTeamContests(teamOID)
	if err != nil {
		return nil, err
	}

	var contests []models.Contest
	for _, contestID := range contestIDs {
		contest, err := s.contestEntityRepo.FindByID(contestID.Hex())
		if err == nil {
			contests = append(contests, *contest)
		}
	}

	return contests, nil
}
