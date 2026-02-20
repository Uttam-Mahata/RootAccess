package services

import (
	"errors"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HintService struct {
	hintRepo      *repositories.HintRepository
	challengeRepo *repositories.ChallengeRepository
	teamRepo      *repositories.TeamRepository
}

func NewHintService(
	hintRepo *repositories.HintRepository,
	challengeRepo *repositories.ChallengeRepository,
	teamRepo *repositories.TeamRepository,
) *HintService {
	return &HintService{
		hintRepo:      hintRepo,
		challengeRepo: challengeRepo,
		teamRepo:      teamRepo,
	}
}

// HintResponse represents a hint with its reveal status
type HintResponse struct {
	ID       string `json:"id"`
	Cost     int    `json:"cost"`
	Order    int    `json:"order"`
	Content  string `json:"content,omitempty"` // Only set if revealed
	Revealed bool   `json:"revealed"`
}

// GetHintsForChallenge returns hints for a challenge with reveal status
func (s *HintService) GetHintsForChallenge(challengeID string, userID primitive.ObjectID) ([]HintResponse, error) {
	challenge, err := s.challengeRepo.GetChallengeByID(challengeID)
	if err != nil {
		return nil, err
	}

	cid, _ := primitive.ObjectIDFromHex(challengeID)

	// Check if user is in a team
	team, _ := s.teamRepo.FindTeamByMemberID(userID.Hex())

	var reveals []models.HintReveal
	if team != nil {
		reveals, _ = s.hintRepo.GetRevealsByTeamAndChallenge(team.ID, cid)
	} else {
		reveals, _ = s.hintRepo.GetRevealsByUserAndChallenge(userID, cid)
	}

	revealedHintIDs := make(map[primitive.ObjectID]bool)
	for _, r := range reveals {
		revealedHintIDs[r.HintID] = true
	}

	var response []HintResponse
	for _, hint := range challenge.Hints {
		hr := HintResponse{
			ID:       hint.ID.Hex(),
			Cost:     hint.Cost,
			Order:    hint.Order,
			Revealed: revealedHintIDs[hint.ID],
		}
		if hr.Revealed {
			hr.Content = hint.Content
		}
		response = append(response, hr)
	}

	return response, nil
}

// RevealHint reveals a hint for a user, deducting points from their team
func (s *HintService) RevealHint(challengeID string, hintID string, userID primitive.ObjectID) (*HintResponse, error) {
	challenge, err := s.challengeRepo.GetChallengeByID(challengeID)
	if err != nil {
		return nil, err
	}

	hintOID, err := primitive.ObjectIDFromHex(hintID)
	if err != nil {
		return nil, errors.New("invalid hint ID")
	}

	cid, _ := primitive.ObjectIDFromHex(challengeID)

	// Find the hint in the challenge
	var targetHint *models.Hint
	for i := range challenge.Hints {
		if challenge.Hints[i].ID == hintOID {
			targetHint = &challenge.Hints[i]
			break
		}
	}
	if targetHint == nil {
		return nil, errors.New("hint not found")
	}

	// Check if user is in a team
	team, _ := s.teamRepo.FindTeamByMemberID(userID.Hex())

	// Check if already revealed (by team or user)
	if team != nil {
		existing, _ := s.hintRepo.FindRevealByTeam(hintOID, team.ID)
		if existing != nil {
			return &HintResponse{
				ID:       targetHint.ID.Hex(),
				Cost:     targetHint.Cost,
				Order:    targetHint.Order,
				Content:  targetHint.Content,
				Revealed: true,
			}, nil
		}
	} else {
		existing, _ := s.hintRepo.FindReveal(hintOID, userID)
		if existing != nil {
			return &HintResponse{
				ID:       targetHint.ID.Hex(),
				Cost:     targetHint.Cost,
				Order:    targetHint.Order,
				Content:  targetHint.Content,
				Revealed: true,
			}, nil
		}
	}

	// Record the reveal
	reveal := &models.HintReveal{
		HintID:      hintOID,
		ChallengeID: cid,
		UserID:      userID,
		Cost:        targetHint.Cost,
	}
	if team != nil {
		reveal.TeamID = team.ID
		// Deduct points from team
		s.teamRepo.UpdateTeamScore(team.ID.Hex(), -targetHint.Cost)
	}

	if err := s.hintRepo.CreateReveal(reveal); err != nil {
		return nil, err
	}

	return &HintResponse{
		ID:       targetHint.ID.Hex(),
		Cost:     targetHint.Cost,
		Order:    targetHint.Order,
		Content:  targetHint.Content,
		Revealed: true,
	}, nil
}
