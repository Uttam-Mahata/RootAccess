package services

import (
	"context"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/database"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/utils"
)

type ChallengeService struct {
	challengeRepo    *repositories.ChallengeRepository
	submissionRepo   *repositories.SubmissionRepository
	teamRepo         *repositories.TeamRepository
	contestSolveRepo *repositories.ContestSolveRepository
}

func NewChallengeService(
	challengeRepo *repositories.ChallengeRepository,
	submissionRepo *repositories.SubmissionRepository,
	teamRepo *repositories.TeamRepository,
	contestSolveRepo *repositories.ContestSolveRepository,
) *ChallengeService {
	return &ChallengeService{
		challengeRepo:    challengeRepo,
		submissionRepo:   submissionRepo,
		teamRepo:         teamRepo,
		contestSolveRepo: contestSolveRepo,
	}
}

func (s *ChallengeService) invalidateScoreboardCache() {
	if database.Registry != nil && database.Registry.Scoreboard != nil {
		ctx := context.Background()
		database.Registry.Scoreboard.Del(ctx, "scoreboard")
		database.Registry.Scoreboard.Del(ctx, "team_scoreboard")
	}
}

func (s *ChallengeService) CreateChallenge(challenge *models.Challenge) error {
	err := s.challengeRepo.CreateChallenge(challenge)
	if err == nil {
		s.invalidateScoreboardCache()
	}
	return err
}

func (s *ChallengeService) GetAllChallenges() ([]models.Challenge, error) {
	return s.challengeRepo.GetAllChallenges()
}

func (s *ChallengeService) GetAllChallengesForList() ([]models.Challenge, error) {
	return s.challengeRepo.GetAllChallengesForList()
}

func (s *ChallengeService) GetChallengeByID(id string) (*models.Challenge, error) {
	return s.challengeRepo.GetChallengeByID(id)
}

func (s *ChallengeService) UpdateChallenge(id string, challenge *models.Challenge) error {
	err := s.challengeRepo.UpdateChallenge(id, challenge)
	if err == nil {
		s.invalidateScoreboardCache()
	}
	return err
}

func (s *ChallengeService) DeleteChallenge(id string) error {
	err := s.challengeRepo.DeleteChallenge(id)
	if err == nil {
		s.invalidateScoreboardCache()
	}
	return err
}

// UpdateOfficialWriteup updates the official writeup content and format
func (s *ChallengeService) UpdateOfficialWriteup(id string, content, format string) error {
	return s.challengeRepo.UpdateOfficialWriteup(id, content, format)
}

// PublishOfficialWriteup sets OfficialWriteupPublished to true (caller must verify contest has ended)
func (s *ChallengeService) PublishOfficialWriteup(id string) error {
	return s.challengeRepo.PublishOfficialWriteup(id)
}

// SubmitFlagResult contains the result of a flag submission
type SubmitFlagResult struct {
	IsCorrect     bool   `json:"is_correct"`
	AlreadySolved bool   `json:"already_solved"`
	TeamID        string `json:"team_id,omitempty"`
	TeamName      string `json:"team_name,omitempty"`
	Points        int    `json:"points,omitempty"`
	SolveCount    int    `json:"solve_count,omitempty"`
	Message       string `json:"message,omitempty"`
}

func (s *ChallengeService) SubmitFlag(userID string, challengeID string, flag string, clientIP string, contestID *string) (*SubmitFlagResult, error) {
	cID := ""
	if contestID != nil {
		cID = *contestID
	}
	challenge, err := s.challengeRepo.GetChallengeByID(challengeID)
	if err != nil {
		return nil, err
	}

	result := &SubmitFlagResult{}

	// 1. Check if CURRENT user already solved it IN THIS CONTEST
	var existingUserSolve *models.Submission
	if cID != "" {
		existingUserSolve, _ = s.submissionRepo.FindByChallengeAndUserInContest(challengeID, userID, cID)
	} else {
		existingUserSolve, _ = s.submissionRepo.FindByChallengeAndUser(challengeID, userID)
	}
	if existingUserSolve != nil {
		result.IsCorrect = true
		result.AlreadySolved = true
		// Return contest-specific points if in a contest
		if cID != "" && s.contestSolveRepo != nil {
			contestSolves, _ := s.contestSolveRepo.GetContestSolveCount(cID, challengeID)
			result.Points = challenge.PointsForSolveCount(contestSolves)
			result.SolveCount = contestSolves
		} else {
			result.Points = challenge.CurrentPoints()
			result.SolveCount = challenge.SolveCount
		}

		team, _ := s.teamRepo.FindTeamByMemberID(userID)
		if team != nil {
			result.TeamID = team.ID
			result.TeamName = team.Name
		}

		return result, nil
	}

	// Verify flag using hash comparison
	isCorrect := utils.VerifyFlag(flag, challenge.FlagHash)
	result.IsCorrect = isCorrect

	// Check if user is in a team
	team, _ := s.teamRepo.FindTeamByMemberID(userID)

	if team != nil {
		result.TeamID = team.ID
		result.TeamName = team.Name

		// 2. Check if TEAM already solved it IN THIS CONTEST
		teamAlreadySolved := false
		if cID != "" {
			existingTeamSolve, _ := s.submissionRepo.FindByChallengeAndTeamInContest(challengeID, team.ID, cID)
			if existingTeamSolve != nil {
				teamAlreadySolved = true
			}
		} else {
			existingTeamSolve, _ := s.submissionRepo.FindByChallengeAndTeam(challengeID, team.ID)
			if existingTeamSolve != nil {
				teamAlreadySolved = true
			}
		}

		// Hash the submitted flag for storage
		flagHash := utils.HashFlag(flag)

		submission := &models.Submission{
			UserID:      userID,
			TeamID:      team.ID,
			ChallengeID: challengeID,
			ContestID:   cID,
			Flag:        flagHash,
			IsCorrect:   isCorrect,
			IPAddress:   clientIP,
		}

		err = s.submissionRepo.CreateSubmission(submission)
		if err != nil {
			return nil, err
		}

		if isCorrect {
			s.invalidateScoreboardCache()

			if !teamAlreadySolved {
				// Increment global solve count
				s.challengeRepo.IncrementSolveCount(challengeID)

				// Also increment contest-specific solve count
				if cID != "" && s.contestSolveRepo != nil {
					s.contestSolveRepo.IncrementContestSolveCount(cID, challengeID)
				}

				// Use contest-specific solve count for point calculation
				if cID != "" && s.contestSolveRepo != nil {
					contestSolves, _ := s.contestSolveRepo.GetContestSolveCount(cID, challengeID)
					points := challenge.PointsForSolveCount(contestSolves)
					result.Points = points
					result.SolveCount = contestSolves
				} else {
					challenge, _ = s.challengeRepo.GetChallengeByID(challengeID)
					result.Points = challenge.CurrentPoints()
					result.SolveCount = challenge.SolveCount
				}

				// Award points to team (global score still tracks cumulative)
				s.teamRepo.UpdateTeamScore(team.ID, result.Points)

				result.Message = "Flag correct! Points awarded to team " + team.Name
			} else {
				result.AlreadySolved = true
				if cID != "" && s.contestSolveRepo != nil {
					contestSolves, _ := s.contestSolveRepo.GetContestSolveCount(cID, challengeID)
					result.Points = challenge.PointsForSolveCount(contestSolves)
					result.SolveCount = contestSolves
				} else {
					result.Points = challenge.CurrentPoints()
					result.SolveCount = challenge.SolveCount
				}
				result.Message = "Flag correct! (Team already solved)"
			}
		}

		return result, nil
	}

	// Individual submission (no team)
	flagHash := utils.HashFlag(flag)

	submission := &models.Submission{
		UserID:      userID,
		ChallengeID: challengeID,
		ContestID:   cID,
		Flag:        flagHash,
		IsCorrect:   isCorrect,
		IPAddress:   clientIP,
	}

	err = s.submissionRepo.CreateSubmission(submission)
	if err != nil {
		return nil, err
	}

	if isCorrect {
		s.challengeRepo.IncrementSolveCount(challengeID)

		if cID != "" && s.contestSolveRepo != nil {
			s.contestSolveRepo.IncrementContestSolveCount(cID, challengeID)
			contestSolves, _ := s.contestSolveRepo.GetContestSolveCount(cID, challengeID)
			result.Points = challenge.PointsForSolveCount(contestSolves)
			result.SolveCount = contestSolves
		} else {
			challenge, _ = s.challengeRepo.GetChallengeByID(challengeID)
			result.Points = challenge.CurrentPoints()
			result.SolveCount = challenge.SolveCount
		}
		s.invalidateScoreboardCache()
	}

	return result, nil
}
