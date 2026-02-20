package services

import (
	"errors"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ContestAdminService struct {
	contestRepo       *repositories.ContestRepository
	contestEntityRepo  *repositories.ContestEntityRepository
	contestRoundRepo   *repositories.ContestRoundRepository
	roundChallengeRepo *repositories.RoundChallengeRepository
	challengeRepo      *repositories.ChallengeRepository
	registrationRepo   *repositories.TeamContestRegistrationRepository
}

func NewContestAdminService(
	contestRepo *repositories.ContestRepository,
	contestEntityRepo *repositories.ContestEntityRepository,
	contestRoundRepo *repositories.ContestRoundRepository,
	roundChallengeRepo *repositories.RoundChallengeRepository,
	challengeRepo *repositories.ChallengeRepository,
	registrationRepo *repositories.TeamContestRegistrationRepository,
) *ContestAdminService {
	return &ContestAdminService{
		contestRepo:       contestRepo,
		contestEntityRepo: contestEntityRepo,
		contestRoundRepo:  contestRoundRepo,
		roundChallengeRepo: roundChallengeRepo,
		challengeRepo:     challengeRepo,
		registrationRepo:  registrationRepo,
	}
}

// ListContests returns all contest entities
func (s *ContestAdminService) ListContests() ([]models.Contest, error) {
	return s.contestEntityRepo.ListAll()
}

// GetContest returns a contest by ID
func (s *ContestAdminService) GetContest(id string) (*models.Contest, error) {
	return s.contestEntityRepo.FindByID(id)
}

// CreateContest creates a new contest
func (s *ContestAdminService) CreateContest(name, description string, startTime, endTime time.Time) (*models.Contest, error) {
	if !endTime.After(startTime) {
		return nil, errors.New("end time must be after start time")
	}

	contest := &models.Contest{
		Name:        name,
		Description: description,
		StartTime:   startTime,
		EndTime:     endTime,
		IsActive:    false,
	}
	if err := s.contestEntityRepo.Create(contest); err != nil {
		return nil, err
	}
	return contest, nil
}

// UpdateContest updates a contest
func (s *ContestAdminService) UpdateContest(id string, name, description string, startTime, endTime time.Time, isActive bool) (*models.Contest, error) {
	contest, err := s.contestEntityRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if !endTime.After(startTime) {
		return nil, errors.New("end time must be after start time")
	}

	contest.Name = name
	contest.Description = description
	contest.StartTime = startTime
	contest.EndTime = endTime
	contest.IsActive = isActive

	if err := s.contestEntityRepo.Update(contest); err != nil {
		return nil, err
	}
	return contest, nil
}

// DeleteContest deletes a contest and its rounds
func (s *ContestAdminService) DeleteContest(id string) error {
	if err := s.contestRoundRepo.DeleteByContestID(id); err != nil {
		return err
	}
	return s.contestEntityRepo.Delete(id)
}

// SetActiveContest sets which contest is the active one (updates ContestConfig)
func (s *ContestAdminService) SetActiveContest(contestID string) error {
	oid, err := primitive.ObjectIDFromHex(contestID)
	if err != nil {
		return err
	}

	config, err := s.contestRepo.GetActiveContest()
	if err != nil {
		config = &models.ContestConfig{
			Title:                "Active Contest",
			StartTime:            time.Now(),
			EndTime:              time.Now().Add(24 * time.Hour),
			IsActive:             true,
			ScoreboardVisibility: "public",
		}
	}

	contest, err := s.contestEntityRepo.FindByID(contestID)
	if err != nil {
		return err
	}

	config.ContestID = oid
	config.Title = contest.Name
	config.StartTime = contest.StartTime
	config.EndTime = contest.EndTime

	return s.contestRepo.UpsertContest(config)
}

// ListRounds returns all rounds for a contest
func (s *ContestAdminService) ListRounds(contestID string) ([]models.ContestRound, error) {
	return s.contestRoundRepo.ListByContestID(contestID)
}

// CreateRound creates a new round for a contest
func (s *ContestAdminService) CreateRound(contestID string, name, description string, order int, visibleFrom, startTime, endTime time.Time) (*models.ContestRound, error) {
	oid, err := primitive.ObjectIDFromHex(contestID)
	if err != nil {
		return nil, err
	}

	if !endTime.After(startTime) {
		return nil, errors.New("end time must be after start time")
	}

	round := &models.ContestRound{
		ContestID:   oid,
		Name:        name,
		Description: description,
		Order:       order,
		VisibleFrom: visibleFrom,
		StartTime:   startTime,
		EndTime:     endTime,
	}
	if err := s.contestRoundRepo.Create(round); err != nil {
		return nil, err
	}
	return round, nil
}

// UpdateRound updates a round
func (s *ContestAdminService) UpdateRound(id string, name, description string, order int, visibleFrom, startTime, endTime time.Time) (*models.ContestRound, error) {
	round, err := s.contestRoundRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if !endTime.After(startTime) {
		return nil, errors.New("end time must be after start time")
	}

	round.Name = name
	round.Description = description
	round.Order = order
	round.VisibleFrom = visibleFrom
	round.StartTime = startTime
	round.EndTime = endTime

	if err := s.contestRoundRepo.Update(round); err != nil {
		return nil, err
	}
	return round, nil
}

// DeleteRound deletes a round and its challenge attachments
func (s *ContestAdminService) DeleteRound(id string) error {
	return s.contestRoundRepo.Delete(id)
}

// AttachChallengesToRound attaches challenges to a round and sets their ContestID
func (s *ContestAdminService) AttachChallengesToRound(roundID string, challengeIDs []string) error {
	round, err := s.contestRoundRepo.FindByID(roundID)
	if err != nil {
		return err
	}

	for _, cid := range challengeIDs {
		challengeOID, err := primitive.ObjectIDFromHex(cid)
		if err != nil {
			continue
		}
		if err := s.roundChallengeRepo.Attach(round.ID, challengeOID); err != nil {
			return err
		}
		// Set contest_id on challenge so we know which contest it belongs to
		_ = s.challengeRepo.SetContestID(cid, round.ContestID)
	}
	return nil
}

// DetachChallengesFromRound removes challenges from a round
func (s *ContestAdminService) DetachChallengesFromRound(roundID string, challengeIDs []string) error {
	round, err := s.contestRoundRepo.FindByID(roundID)
	if err != nil {
		return err
	}

	for _, cid := range challengeIDs {
		challengeOID, err := primitive.ObjectIDFromHex(cid)
		if err != nil {
			continue
		}
		if err := s.roundChallengeRepo.Detach(round.ID, challengeOID); err != nil {
			return err
		}
	}
	return nil
}

// GetChallengesForRound returns challenge IDs attached to a round
func (s *ContestAdminService) GetChallengesForRound(roundID string) ([]primitive.ObjectID, error) {
	oid, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		return nil, err
	}
	return s.roundChallengeRepo.GetChallengesByRound(oid)
}

// GetVisibleChallengeIDs returns challenge IDs that are currently visible to players.
// Returns nil if no active contest or outside contest/round times.
// If teamID is provided, only returns challenges from contests the team is registered for
func (s *ContestAdminService) GetVisibleChallengeIDs(now time.Time, teamID *primitive.ObjectID) ([]primitive.ObjectID, error) {
	config, err := s.contestRepo.GetActiveContest()
	if err != nil || config == nil || config.ContestID.IsZero() {
		return nil, nil
	}

	contest, err := s.contestEntityRepo.FindByID(config.ContestID.Hex())
	if err != nil || contest == nil {
		return nil, nil
	}

	// Contest must be within its time window
	if now.Before(contest.StartTime) || now.After(contest.EndTime) {
		return nil, nil
	}

	// Always require team registration for active contests
	if teamID == nil || teamID.IsZero() {
		return nil, nil // No team = no access to contest challenges
	}
	if s.registrationRepo != nil {
		registered, err := s.registrationRepo.IsTeamRegistered(*teamID, contest.ID)
		if err != nil || !registered {
			return nil, nil // Team not registered, return empty
		}
	}

	// Get rounds that are currently visible/active
	rounds, err := s.contestRoundRepo.GetActiveRounds(config.ContestID.Hex(), now)
	if err != nil || len(rounds) == 0 {
		return nil, nil
	}

	roundIDs := make([]primitive.ObjectID, len(rounds))
	for i := range rounds {
		roundIDs[i] = rounds[i].ID
	}

	return s.roundChallengeRepo.GetChallengeIDsForRounds(roundIDs)
}

// GetVisibleChallenges returns challenges that are currently visible to players (published, in active contest/round)
// If teamID is provided, only returns challenges from contests the team is registered for
func (s *ContestAdminService) GetVisibleChallenges(now time.Time, teamID *primitive.ObjectID) ([]models.Challenge, error) {
	ids, err := s.GetVisibleChallengeIDs(now, teamID)
	if err != nil || len(ids) == 0 {
		return nil, err
	}
	return s.challengeRepo.GetChallengesByIDs(ids, true)
}

// HasContestEndedForChallenge returns true if the challenge's owning contest has ended (or challenge has no contest)
func (s *ContestAdminService) HasContestEndedForChallenge(challengeID string, now time.Time) (bool, error) {
	challenge, err := s.challengeRepo.GetChallengeByID(challengeID)
	if err != nil || challenge == nil {
		return false, err
	}
	if challenge.ContestID == nil || challenge.ContestID.IsZero() {
		return true, nil // No contest = treat as ended (legacy challenges)
	}
	contest, err := s.contestEntityRepo.FindByID(challenge.ContestID.Hex())
	if err != nil || contest == nil {
		return false, err
	}
	return now.After(contest.EndTime), nil
}

// IsChallengeVisible returns true if the challenge is currently visible to players
// If teamID is provided, also checks if team is registered for the contest
func (s *ContestAdminService) IsChallengeVisible(challengeID string, now time.Time, teamID *primitive.ObjectID) (bool, error) {
	ids, err := s.GetVisibleChallengeIDs(now, teamID)
	if err != nil || len(ids) == 0 {
		return false, err
	}

	cid, err := primitive.ObjectIDFromHex(challengeID)
	if err != nil {
		return false, err
	}

	for _, id := range ids {
		if id == cid {
			return true, nil
		}
	}
	return false, nil
}
