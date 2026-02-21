package turso

import (
	"database/sql"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const hintRevealColumns = `id, hint_id, challenge_id, user_id, team_id, cost`

// HintRepository implements interfaces.HintRepository using database/sql.
type HintRepository struct {
	db *sql.DB
}

// NewHintRepository creates a new Turso-backed HintRepository.
func NewHintRepository(db *sql.DB) *HintRepository {
	return &HintRepository{db: db}
}

func scanHintReveal(s scanner) (*models.HintReveal, error) {
	var h models.HintReveal
	var id, hintID, challengeID, userID, teamID string

	err := s.Scan(&id, &hintID, &challengeID, &userID, &teamID, &h.Cost)
	if err != nil {
		return nil, err
	}

	h.ID = oidFromHex(id)
	h.HintID = oidFromHex(hintID)
	h.ChallengeID = oidFromHex(challengeID)
	h.UserID = oidFromHex(userID)
	h.TeamID = oidFromHex(teamID)
	return &h, nil
}

func (r *HintRepository) queryReveals(query string, args ...interface{}) ([]models.HintReveal, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.HintReveal
	for rows.Next() {
		h, err := scanHintReveal(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *h)
	}
	return results, rows.Err()
}

func (r *HintRepository) FindReveal(hintID, userID primitive.ObjectID) (*models.HintReveal, error) {
	row := r.db.QueryRow(
		`SELECT `+hintRevealColumns+` FROM hint_reveals WHERE hint_id = ? AND user_id = ?`,
		hintID.Hex(), userID.Hex(),
	)
	return scanHintReveal(row)
}

func (r *HintRepository) FindRevealByTeam(hintID, teamID primitive.ObjectID) (*models.HintReveal, error) {
	row := r.db.QueryRow(
		`SELECT `+hintRevealColumns+` FROM hint_reveals WHERE hint_id = ? AND team_id = ?`,
		hintID.Hex(), teamID.Hex(),
	)
	return scanHintReveal(row)
}

func (r *HintRepository) CreateReveal(reveal *models.HintReveal) error {
	id := newID()
	_, err := r.db.Exec(
		`INSERT INTO hint_reveals (`+hintRevealColumns+`) VALUES (?,?,?,?,?,?)`,
		id, reveal.HintID.Hex(), reveal.ChallengeID.Hex(),
		reveal.UserID.Hex(), reveal.TeamID.Hex(), reveal.Cost,
	)
	if err != nil {
		return err
	}
	reveal.ID = oidFromHex(id)
	return nil
}

func (r *HintRepository) GetRevealsByUserAndChallenge(userID, challengeID primitive.ObjectID) ([]models.HintReveal, error) {
	return r.queryReveals(
		`SELECT `+hintRevealColumns+` FROM hint_reveals WHERE user_id = ? AND challenge_id = ?`,
		userID.Hex(), challengeID.Hex(),
	)
}

func (r *HintRepository) GetRevealsByTeamAndChallenge(teamID, challengeID primitive.ObjectID) ([]models.HintReveal, error) {
	return r.queryReveals(
		`SELECT `+hintRevealColumns+` FROM hint_reveals WHERE team_id = ? AND challenge_id = ?`,
		teamID.Hex(), challengeID.Hex(),
	)
}
