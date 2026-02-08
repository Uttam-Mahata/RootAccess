package services

import (
	"errors"
	"time"

	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/repositories"
)

type ContestService struct {
	contestRepo *repositories.ContestRepository
}

func NewContestService(contestRepo *repositories.ContestRepository) *ContestService {
	return &ContestService{
		contestRepo: contestRepo,
	}
}

// GetContestConfig returns the current contest configuration
func (s *ContestService) GetContestConfig() (*models.ContestConfig, error) {
	return s.contestRepo.GetActiveContest()
}

// UpdateContestConfig updates the contest configuration (admin only)
func (s *ContestService) UpdateContestConfig(title string, startTime, endTime time.Time, isActive bool) (*models.ContestConfig, error) {
	if !endTime.After(startTime) {
		return nil, errors.New("end time must be after start time")
	}

	existing, err := s.contestRepo.GetActiveContest()
	if err != nil {
		// Create new config
		config := &models.ContestConfig{
			Title:     title,
			StartTime: startTime,
			EndTime:   endTime,
			IsActive:  isActive,
		}
		if err := s.contestRepo.UpsertContest(config); err != nil {
			return nil, err
		}
		return config, nil
	}

	// Update existing
	existing.Title = title
	existing.StartTime = startTime
	existing.EndTime = endTime
	existing.IsActive = isActive

	if err := s.contestRepo.UpsertContest(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// GetContestStatus returns the current status of the contest
func (s *ContestService) GetContestStatus() (models.ContestStatus, *models.ContestConfig, error) {
	config, err := s.contestRepo.GetActiveContest()
	if err != nil {
		return models.ContestStatusNotStarted, nil, nil
	}
	return config.GetStatus(), config, nil
}
