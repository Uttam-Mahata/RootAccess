package turso

import (
	"database/sql"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RoundChallengeRepository implements interfaces.RoundChallengeRepository using database/sql.
type RoundChallengeRepository struct {
	db *sql.DB
}

// NewRoundChallengeRepository creates a new Turso-backed RoundChallengeRepository.
func NewRoundChallengeRepository(db *sql.DB) *RoundChallengeRepository {
	return &RoundChallengeRepository{db: db}
}

func (r *RoundChallengeRepository) Attach(roundID, challengeID primitive.ObjectID) error {
	id := newID()
	_, err := r.db.Exec(
		`INSERT INTO round_challenges (id, round_id, challenge_id, created_at) VALUES (?,?,?,?)`,
		id, roundID.Hex(), challengeID.Hex(), timeToStr(time.Now()),
	)
	return err
}

func (r *RoundChallengeRepository) Detach(roundID, challengeID primitive.ObjectID) error {
	_, err := r.db.Exec(
		`DELETE FROM round_challenges WHERE round_id = ? AND challenge_id = ?`,
		roundID.Hex(), challengeID.Hex(),
	)
	return err
}

func (r *RoundChallengeRepository) GetChallengeIDsForRounds(roundIDs []primitive.ObjectID) ([]primitive.ObjectID, error) {
	if len(roundIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(roundIDs))
	args := make([]interface{}, len(roundIDs))
	for i, rid := range roundIDs {
		placeholders[i] = "?"
		args[i] = rid.Hex()
	}

	rows, err := r.db.Query(
		`SELECT DISTINCT challenge_id FROM round_challenges WHERE round_id IN (`+strings.Join(placeholders, ",")+`)`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []primitive.ObjectID
	for rows.Next() {
		var cid string
		if err := rows.Scan(&cid); err != nil {
			return nil, err
		}
		ids = append(ids, oidFromHex(cid))
	}
	return ids, rows.Err()
}

func (r *RoundChallengeRepository) GetRoundIDsForChallenge(challengeID primitive.ObjectID) ([]primitive.ObjectID, error) {
	rows, err := r.db.Query(
		`SELECT round_id FROM round_challenges WHERE challenge_id = ?`,
		challengeID.Hex(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []primitive.ObjectID
	for rows.Next() {
		var rid string
		if err := rows.Scan(&rid); err != nil {
			return nil, err
		}
		ids = append(ids, oidFromHex(rid))
	}
	return ids, rows.Err()
}

func (r *RoundChallengeRepository) GetChallengesByRound(roundID primitive.ObjectID) ([]primitive.ObjectID, error) {
	rows, err := r.db.Query(
		`SELECT challenge_id FROM round_challenges WHERE round_id = ?`,
		roundID.Hex(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []primitive.ObjectID
	for rows.Next() {
		var cid string
		if err := rows.Scan(&cid); err != nil {
			return nil, err
		}
		ids = append(ids, oidFromHex(cid))
	}
	return ids, rows.Err()
}
