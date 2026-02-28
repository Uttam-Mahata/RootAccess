package repositories

import (
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
)

type RoundChallengeRepository struct {
	db *sql.DB
}

func NewRoundChallengeRepository(db *sql.DB) *RoundChallengeRepository {
	return &RoundChallengeRepository{db: db}
}

func (r *RoundChallengeRepository) Attach(roundID, challengeID string) error {
	id := uuid.New().String()
	_, err := r.db.Exec("INSERT OR IGNORE INTO round_challenges (id, round_id, challenge_id, created_at) VALUES (?, ?, ?, ?)",
		id, roundID, challengeID, time.Now().Format(time.RFC3339))
	return err
}

func (r *RoundChallengeRepository) Detach(roundID, challengeID string) error {
	_, err := r.db.Exec("DELETE FROM round_challenges WHERE round_id=? AND challenge_id=?", roundID, challengeID)
	return err
}

func (r *RoundChallengeRepository) GetChallengeIDsForRounds(roundIDs []string) ([]string, error) {
	if len(roundIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(roundIDs))
	args := make([]interface{}, len(roundIDs))
	for i, id := range roundIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := "SELECT DISTINCT challenge_id FROM round_challenges WHERE round_id IN (" + strings.Join(placeholders, ",") + ")"
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *RoundChallengeRepository) GetRoundIDsForChallenge(challengeID string) ([]string, error) {
	rows, err := r.db.Query("SELECT round_id FROM round_challenges WHERE challenge_id=?", challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *RoundChallengeRepository) GetChallengesByRound(roundID string) ([]string, error) {
	rows, err := r.db.Query("SELECT challenge_id FROM round_challenges WHERE round_id=?", roundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
