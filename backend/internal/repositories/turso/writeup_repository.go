package turso

import (
	"database/sql"
	"strings"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const writeupColumns = `id, challenge_id, user_id, username, content, content_format, status, upvotes, created_at, updated_at`

// WriteupRepository implements interfaces.WriteupRepository using database/sql.
type WriteupRepository struct {
	db *sql.DB
}

// NewWriteupRepository creates a new Turso-backed WriteupRepository.
func NewWriteupRepository(db *sql.DB) *WriteupRepository {
	return &WriteupRepository{db: db}
}

func scanWriteup(s scanner) (*models.Writeup, error) {
	var w models.Writeup
	var (
		id, challengeID, userID string
		createdAt, updatedAt    string
	)

	err := s.Scan(&id, &challengeID, &userID, &w.Username, &w.Content,
		&w.ContentFormat, &w.Status, &w.Upvotes, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	w.ID = oidFromHex(id)
	w.ChallengeID = oidFromHex(challengeID)
	w.UserID = oidFromHex(userID)
	w.CreatedAt = strToTime(createdAt)
	w.UpdatedAt = strToTime(updatedAt)
	return &w, nil
}

func (r *WriteupRepository) loadUpvotedBy(writeupID string) ([]primitive.ObjectID, error) {
	rows, err := r.db.Query(`SELECT user_id FROM writeup_upvotes WHERE writeup_id = ?`, writeupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []primitive.ObjectID
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		ids = append(ids, oidFromHex(uid))
	}
	return ids, rows.Err()
}

func (r *WriteupRepository) scanAndLoad(s scanner) (*models.Writeup, error) {
	w, err := scanWriteup(s)
	if err != nil {
		return nil, err
	}
	w.UpvotedBy, err = r.loadUpvotedBy(w.ID.Hex())
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *WriteupRepository) queryWriteups(query string, args ...interface{}) ([]models.Writeup, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.Writeup
	for rows.Next() {
		w, err := scanWriteup(rows)
		if err != nil {
			return nil, err
		}
		w.UpvotedBy, err = r.loadUpvotedBy(w.ID.Hex())
		if err != nil {
			return nil, err
		}
		results = append(results, *w)
	}
	return results, rows.Err()
}

func (r *WriteupRepository) CreateWriteup(writeup *models.Writeup) error {
	id := newID()
	now := time.Now()
	writeup.CreatedAt = now
	writeup.UpdatedAt = now

	_, err := r.db.Exec(
		`INSERT INTO writeups (`+writeupColumns+`) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		id, writeup.ChallengeID.Hex(), writeup.UserID.Hex(), writeup.Username,
		writeup.Content, writeup.ContentFormat, writeup.Status, writeup.Upvotes,
		timeToStr(now), timeToStr(now),
	)
	if err != nil {
		return err
	}
	writeup.ID = oidFromHex(id)
	return nil
}

func (r *WriteupRepository) GetWriteupByID(id string) (*models.Writeup, error) {
	row := r.db.QueryRow(`SELECT `+writeupColumns+` FROM writeups WHERE id = ?`, id)
	return r.scanAndLoad(row)
}

func (r *WriteupRepository) GetWriteupsByChallenge(challengeID primitive.ObjectID, onlyApproved bool) ([]models.Writeup, error) {
	if onlyApproved {
		return r.queryWriteups(
			`SELECT `+writeupColumns+` FROM writeups WHERE challenge_id = ? AND status = 'approved' ORDER BY upvotes DESC`,
			challengeID.Hex(),
		)
	}
	return r.queryWriteups(
		`SELECT `+writeupColumns+` FROM writeups WHERE challenge_id = ? ORDER BY created_at DESC`,
		challengeID.Hex(),
	)
}

func (r *WriteupRepository) GetWriteupsByUser(userID primitive.ObjectID) ([]models.Writeup, error) {
	return r.queryWriteups(
		`SELECT `+writeupColumns+` FROM writeups WHERE user_id = ? ORDER BY created_at DESC`,
		userID.Hex(),
	)
}

func (r *WriteupRepository) GetAllWriteups() ([]models.Writeup, error) {
	return r.queryWriteups(`SELECT ` + writeupColumns + ` FROM writeups ORDER BY created_at DESC`)
}

func (r *WriteupRepository) UpdateWriteupStatus(id string, status string) error {
	_, err := r.db.Exec(
		`UPDATE writeups SET status = ?, updated_at = ? WHERE id = ?`,
		status, timeToStr(time.Now()), id,
	)
	return err
}

func (r *WriteupRepository) DeleteWriteup(id string) error {
	_, err := r.db.Exec(`DELETE FROM writeup_upvotes WHERE writeup_id = ?`, id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`DELETE FROM writeups WHERE id = ?`, id)
	return err
}

func (r *WriteupRepository) FindByUserAndChallenge(userID, challengeID primitive.ObjectID) (*models.Writeup, error) {
	row := r.db.QueryRow(
		`SELECT `+writeupColumns+` FROM writeups WHERE user_id = ? AND challenge_id = ? LIMIT 1`,
		userID.Hex(), challengeID.Hex(),
	)
	return r.scanAndLoad(row)
}

func (r *WriteupRepository) UpdateWriteupContent(id string, content string, contentFormat string) error {
	_, err := r.db.Exec(
		`UPDATE writeups SET content = ?, content_format = ?, updated_at = ? WHERE id = ?`,
		content, contentFormat, timeToStr(time.Now()), id,
	)
	return err
}

func (r *WriteupRepository) GetWriteupsByTeam(teamID string) ([]models.Writeup, error) {
	// Find all user IDs belonging to this team, then find their writeups.
	memberRows, err := r.db.Query(`SELECT member_id FROM team_member_ids WHERE team_id = ?`, teamID)
	if err != nil {
		return nil, err
	}
	defer memberRows.Close()

	var memberIDs []string
	for memberRows.Next() {
		var mid string
		if err := memberRows.Scan(&mid); err != nil {
			return nil, err
		}
		memberIDs = append(memberIDs, mid)
	}
	if err := memberRows.Err(); err != nil {
		return nil, err
	}
	if len(memberIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(memberIDs))
	args := make([]interface{}, len(memberIDs))
	for i, mid := range memberIDs {
		placeholders[i] = "?"
		args[i] = mid
	}

	return r.queryWriteups(
		`SELECT `+writeupColumns+` FROM writeups WHERE user_id IN (`+strings.Join(placeholders, ",")+`) ORDER BY created_at DESC`,
		args...,
	)
}

func (r *WriteupRepository) ToggleUpvote(id string, userID primitive.ObjectID) (bool, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var exists int
	err = tx.QueryRow(
		`SELECT COUNT(*) FROM writeup_upvotes WHERE writeup_id = ? AND user_id = ?`,
		id, userID.Hex(),
	).Scan(&exists)
	if err != nil {
		return false, err
	}

	now := timeToStr(time.Now())
	if exists > 0 {
		if _, err = tx.Exec(`DELETE FROM writeup_upvotes WHERE writeup_id = ? AND user_id = ?`, id, userID.Hex()); err != nil {
			return false, err
		}
		if _, err = tx.Exec(`UPDATE writeups SET upvotes = upvotes - 1, updated_at = ? WHERE id = ?`, now, id); err != nil {
			return false, err
		}
		return false, tx.Commit()
	}

	if _, err = tx.Exec(`INSERT INTO writeup_upvotes (writeup_id, user_id) VALUES (?, ?)`, id, userID.Hex()); err != nil {
		return false, err
	}
	if _, err = tx.Exec(`UPDATE writeups SET upvotes = upvotes + 1, updated_at = ? WHERE id = ?`, now, id); err != nil {
		return false, err
	}
	return true, tx.Commit()
}
