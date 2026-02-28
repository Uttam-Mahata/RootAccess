package repositories

import (
	"database/sql"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/google/uuid"
)

type HintRepository struct {
	db *sql.DB
}

func NewHintRepository(db *sql.DB) *HintRepository {
	return &HintRepository{db: db}
}

func (r *HintRepository) FindReveal(hintID, userID string) (*models.HintReveal, error) {
	var h models.HintReveal
	err := r.db.QueryRow("SELECT id, hint_id, challenge_id, user_id, team_id, cost FROM hint_reveals WHERE hint_id=? AND user_id=?", hintID, userID).
		Scan(&h.ID, &h.HintID, &h.ChallengeID, &h.UserID, &h.TeamID, &h.Cost)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *HintRepository) FindRevealByTeam(hintID, teamID string) (*models.HintReveal, error) {
	var h models.HintReveal
	err := r.db.QueryRow("SELECT id, hint_id, challenge_id, user_id, team_id, cost FROM hint_reveals WHERE hint_id=? AND team_id=?", hintID, teamID).
		Scan(&h.ID, &h.HintID, &h.ChallengeID, &h.UserID, &h.TeamID, &h.Cost)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *HintRepository) CreateReveal(reveal *models.HintReveal) error {
	if reveal.ID == "" {
		reveal.ID = uuid.New().String()
	}
	_, err := r.db.Exec("INSERT INTO hint_reveals (id, hint_id, challenge_id, user_id, team_id, cost) VALUES (?, ?, ?, ?, ?, ?)",
		reveal.ID, reveal.HintID, reveal.ChallengeID, reveal.UserID, reveal.TeamID, reveal.Cost)
	return err
}

func (r *HintRepository) GetRevealsByUserAndChallenge(userID, challengeID string) ([]models.HintReveal, error) {
	rows, err := r.db.Query("SELECT id, hint_id, challenge_id, user_id, team_id, cost FROM hint_reveals WHERE user_id=? AND challenge_id=?", userID, challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reveals []models.HintReveal
	for rows.Next() {
		var h models.HintReveal
		if err := rows.Scan(&h.ID, &h.HintID, &h.ChallengeID, &h.UserID, &h.TeamID, &h.Cost); err != nil {
			return nil, err
		}
		reveals = append(reveals, h)
	}
	return reveals, nil
}

func (r *HintRepository) GetRevealsByTeamAndChallenge(teamID, challengeID string) ([]models.HintReveal, error) {
	rows, err := r.db.Query("SELECT id, hint_id, challenge_id, user_id, team_id, cost FROM hint_reveals WHERE team_id=? AND challenge_id=?", teamID, challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reveals []models.HintReveal
	for rows.Next() {
		var h models.HintReveal
		if err := rows.Scan(&h.ID, &h.HintID, &h.ChallengeID, &h.UserID, &h.TeamID, &h.Cost); err != nil {
			return nil, err
		}
		reveals = append(reveals, h)
	}
	return reveals, nil
}
