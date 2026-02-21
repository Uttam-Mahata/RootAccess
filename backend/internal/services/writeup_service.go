package services

import (
	"errors"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WriteupService struct {
	writeupRepo    interfaces.WriteupRepository
	submissionRepo interfaces.SubmissionRepository
	teamRepo       interfaces.TeamRepository
}

func NewWriteupService(
	writeupRepo interfaces.WriteupRepository,
	submissionRepo interfaces.SubmissionRepository,
	teamRepo interfaces.TeamRepository,
) *WriteupService {
	return &WriteupService{
		writeupRepo:    writeupRepo,
		submissionRepo: submissionRepo,
		teamRepo:       teamRepo,
	}
}

// CreateWriteup creates a new writeup (user must have solved the challenge)
func (s *WriteupService) CreateWriteup(userID primitive.ObjectID, username string, challengeID string, content string, contentFormat string) (*models.Writeup, error) {
	cid, err := primitive.ObjectIDFromHex(challengeID)
	if err != nil {
		return nil, errors.New("invalid challenge ID")
	}

	// Check if user has solved the challenge
	submission, _ := s.submissionRepo.FindByChallengeAndUser(cid, userID)
	if submission == nil || !submission.IsCorrect {
		return nil, errors.New("you must solve the challenge before submitting a writeup")
	}

	// Check if user already submitted a writeup
	existing, _ := s.writeupRepo.FindByUserAndChallenge(userID, cid)
	if existing != nil {
		return nil, errors.New("you have already submitted a writeup for this challenge")
	}

	writeup := &models.Writeup{
		ChallengeID:   cid,
		UserID:        userID,
		Username:      username,
		Content:       content,
		ContentFormat: contentFormat,
	}

	// Check if user is in a team and if team has participated (solved any challenges)
	team, _ := s.teamRepo.FindTeamByMemberID(userID.Hex())
	if team != nil {
		// Check if team has any correct submissions (has participated)
		teamSubmissions, _ := s.submissionRepo.GetTeamSubmissions(team.ID)
		if len(teamSubmissions) > 0 {
			// Team has participated, auto-approve the writeup
			writeup.Status = models.WriteupStatusApproved
		}
	}

	if err := s.writeupRepo.CreateWriteup(writeup); err != nil {
		return nil, err
	}

	return writeup, nil
}

// GetWriteupsByChallenge returns approved writeups for a challenge
func (s *WriteupService) GetWriteupsByChallenge(challengeID string) ([]models.Writeup, error) {
	cid, err := primitive.ObjectIDFromHex(challengeID)
	if err != nil {
		return nil, errors.New("invalid challenge ID")
	}
	return s.writeupRepo.GetWriteupsByChallenge(cid, true)
}

// GetAllWriteups returns all writeups for admin
func (s *WriteupService) GetAllWriteups(teamID string) ([]models.Writeup, error) {
	if teamID != "" {
		return s.writeupRepo.GetWriteupsByTeam(teamID)
	}
	return s.writeupRepo.GetAllWriteups()
}

// UpdateWriteupStatus updates writeup status (admin only)
func (s *WriteupService) UpdateWriteupStatus(id string, status string) error {
	if status != models.WriteupStatusApproved && status != models.WriteupStatusRejected {
		return errors.New("invalid status, must be 'approved' or 'rejected'")
	}
	return s.writeupRepo.UpdateWriteupStatus(id, status)
}

// DeleteWriteup deletes a writeup
func (s *WriteupService) DeleteWriteup(id string) error {
	return s.writeupRepo.DeleteWriteup(id)
}

// GetMyWriteups returns writeups for a specific user
func (s *WriteupService) GetMyWriteups(userID primitive.ObjectID) ([]models.Writeup, error) {
	return s.writeupRepo.GetWriteupsByUser(userID)
}

// UpdateWriteupContent allows authors to edit their own writeup content
func (s *WriteupService) UpdateWriteupContent(writeupID string, userID primitive.ObjectID, content string, contentFormat string) error {
	writeup, err := s.writeupRepo.GetWriteupByID(writeupID)
	if err != nil {
		return errors.New("writeup not found")
	}
	if writeup.UserID != userID {
		return errors.New("you can only edit your own writeups")
	}
	return s.writeupRepo.UpdateWriteupContent(writeupID, content, contentFormat)
}

// ToggleUpvote toggles a user's upvote on a writeup
func (s *WriteupService) ToggleUpvote(writeupID string, userID primitive.ObjectID) (bool, error) {
	return s.writeupRepo.ToggleUpvote(writeupID, userID)
}
