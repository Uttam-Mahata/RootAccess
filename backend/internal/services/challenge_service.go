package services

import (
	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/repositories"
	"github.com/go-ctf-platform/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChallengeService struct {
	challengeRepo  *repositories.ChallengeRepository
	submissionRepo *repositories.SubmissionRepository
	teamRepo       *repositories.TeamRepository
}

func NewChallengeService(
	challengeRepo *repositories.ChallengeRepository,
	submissionRepo *repositories.SubmissionRepository,
	teamRepo *repositories.TeamRepository,
) *ChallengeService {
	return &ChallengeService{
		challengeRepo:  challengeRepo,
		submissionRepo: submissionRepo,
		teamRepo:       teamRepo,
	}
}

func (s *ChallengeService) CreateChallenge(challenge *models.Challenge) error {
	return s.challengeRepo.CreateChallenge(challenge)
}

func (s *ChallengeService) GetAllChallenges() ([]models.Challenge, error) {
	return s.challengeRepo.GetAllChallenges()
}

func (s *ChallengeService) GetChallengeByID(id string) (*models.Challenge, error) {
	return s.challengeRepo.GetChallengeByID(id)
}

func (s *ChallengeService) UpdateChallenge(id string, challenge *models.Challenge) error {
	return s.challengeRepo.UpdateChallenge(id, challenge)
}

func (s *ChallengeService) DeleteChallenge(id string) error {
	return s.challengeRepo.DeleteChallenge(id)
}

// SubmitFlagResult contains the result of a flag submission
type SubmitFlagResult struct {
	IsCorrect     bool   `json:"is_correct"`
	AlreadySolved bool   `json:"already_solved"`
	TeamID        string `json:"team_id,omitempty"`
	TeamName      string `json:"team_name,omitempty"`
	Points        int    `json:"points,omitempty"`
	SolveCount    int    `json:"solve_count,omitempty"`
}

func (s *ChallengeService) SubmitFlag(userID primitive.ObjectID, challengeID string, flag string) (*SubmitFlagResult, error) {
	challenge, err := s.challengeRepo.GetChallengeByID(challengeID)
	if err != nil {
		return nil, err
	}

	cid, _ := primitive.ObjectIDFromHex(challengeID)

	result := &SubmitFlagResult{}

	// Verify flag using hash comparison
	isCorrect := utils.VerifyFlag(flag, challenge.FlagHash)
	result.IsCorrect = isCorrect

	// Check if user is in a team
	team, _ := s.teamRepo.FindTeamByMemberID(userID.Hex())

	if team != nil {
		// Team submission - check if team already solved
		result.TeamID = team.ID.Hex()
		result.TeamName = team.Name

		existing, _ := s.submissionRepo.FindByChallengeAndTeam(cid, team.ID)
		if existing != nil && existing.IsCorrect {
			result.IsCorrect = true
			result.AlreadySolved = true
			result.Points = challenge.CurrentPoints()
			result.SolveCount = challenge.SolveCount
			return result, nil // Team already solved
		}

		// Hash the submitted flag for storage (don't store plaintext)
		flagHash := utils.HashFlag(flag)

		submission := &models.Submission{
			UserID:      userID,
			TeamID:      team.ID,
			ChallengeID: cid,
			Flag:        flagHash, // Store hash of submitted flag
			IsCorrect:   isCorrect,
		}

		err = s.submissionRepo.CreateSubmission(submission)
		if err != nil {
			return nil, err
		}

		// Award points to team if correct
		if isCorrect {
			// Increment solve count first
			s.challengeRepo.IncrementSolveCount(challengeID)
			
			// Refresh challenge to get updated solve count
			challenge, _ = s.challengeRepo.GetChallengeByID(challengeID)
			
			// Calculate dynamic points
			points := challenge.CurrentPoints()
			result.Points = points
			result.SolveCount = challenge.SolveCount
			
			// Award points to team
			s.teamRepo.UpdateTeamScore(team.ID.Hex(), points)
		}

		return result, nil
	}

	// Individual submission (no team) - for backwards compatibility
	existing, _ := s.submissionRepo.FindByChallengeAndUser(cid, userID)
	if existing != nil && existing.IsCorrect {
		result.IsCorrect = true
		result.AlreadySolved = true
		result.Points = challenge.CurrentPoints()
		result.SolveCount = challenge.SolveCount
		return result, nil // Already solved
	}

	// Hash the submitted flag for storage
	flagHash := utils.HashFlag(flag)

	submission := &models.Submission{
		UserID:      userID,
		ChallengeID: cid,
		Flag:        flagHash, // Store hash of submitted flag
		IsCorrect:   isCorrect,
	}

	err = s.submissionRepo.CreateSubmission(submission)
	if err != nil {
		return nil, err
	}

	if isCorrect {
		// Increment solve count
		s.challengeRepo.IncrementSolveCount(challengeID)
		
		// Refresh challenge to get updated solve count
		challenge, _ = s.challengeRepo.GetChallengeByID(challengeID)
		
		result.Points = challenge.CurrentPoints()
		result.SolveCount = challenge.SolveCount
	}

	return result, nil
}
