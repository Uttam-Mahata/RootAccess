package services

import (
	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChallengeService struct {
	challengeRepo  *repositories.ChallengeRepository
	submissionRepo *repositories.SubmissionRepository
}

func NewChallengeService(challengeRepo *repositories.ChallengeRepository, submissionRepo *repositories.SubmissionRepository) *ChallengeService {
	return &ChallengeService{
		challengeRepo:  challengeRepo,
		submissionRepo: submissionRepo,
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

func (s *ChallengeService) SubmitFlag(userID primitive.ObjectID, challengeID string, flag string) (bool, error) {
	challenge, err := s.challengeRepo.GetChallengeByID(challengeID)
	if err != nil {
		return false, err
	}

	cid, _ := primitive.ObjectIDFromHex(challengeID)

	// Check if already solved
	existing, _ := s.submissionRepo.FindByChallengeAndUser(cid, userID)
	if existing != nil {
		return true, nil // Already solved
	}

	isCorrect := challenge.Flag == flag

	submission := &models.Submission{
		UserID:      userID,
		ChallengeID: cid,
		Flag:        flag,
		IsCorrect:   isCorrect,
	}

	err = s.submissionRepo.CreateSubmission(submission)
	if err != nil {
		return false, err
	}

	return isCorrect, nil
}
