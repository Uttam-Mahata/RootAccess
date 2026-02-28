package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/google/uuid"
)

type WriteupRepository struct {
	db *sql.DB
}

func NewWriteupRepository(db *sql.DB) *WriteupRepository {
	return &WriteupRepository{db: db}
}

func (r *WriteupRepository) CreateWriteup(writeup *models.Writeup) error {
	if writeup.ID == "" {
		writeup.ID = uuid.New().String()
	}
	writeup.CreatedAt = time.Now()
	writeup.UpdatedAt = time.Now()
	if writeup.Status == "" {
		writeup.Status = models.WriteupStatusPending
	}

	query := `INSERT INTO writeups (id, challenge_id, user_id, username, content, content_format, status, upvotes, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, writeup.ID, writeup.ChallengeID, writeup.UserID, writeup.Username, writeup.Content, writeup.ContentFormat, writeup.Status, writeup.Upvotes, writeup.CreatedAt.Format(time.RFC3339), writeup.UpdatedAt.Format(time.RFC3339))
	return err
}

func (r *WriteupRepository) scanWriteups(rows *sql.Rows) ([]models.Writeup, error) {
	var writeups []models.Writeup
	for rows.Next() {
		var w models.Writeup
		var createdAt, updatedAt string
		if err := rows.Scan(&w.ID, &w.ChallengeID, &w.UserID, &w.Username, &w.Content, &w.ContentFormat, &w.Status, &w.Upvotes, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		w.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		w.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		writeups = append(writeups, w)
	}
	return writeups, nil
}

func (r *WriteupRepository) selectWriteupFields() string {
	return "id, challenge_id, user_id, username, content, content_format, status, upvotes, created_at, updated_at"
}

func (r *WriteupRepository) GetWriteupByID(id string) (*models.Writeup, error) {
	query := fmt.Sprintf("SELECT %s FROM writeups WHERE id=?", r.selectWriteupFields())
	rows, err := r.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	writeups, err := r.scanWriteups(rows)
	if err != nil || len(writeups) == 0 {
		return nil, sql.ErrNoRows
	}
	return &writeups[0], nil
}

func (r *WriteupRepository) GetWriteupsByChallenge(challengeID string, onlyApproved bool) ([]models.Writeup, error) {
	query := fmt.Sprintf("SELECT %s FROM writeups WHERE challenge_id=?", r.selectWriteupFields())
	if onlyApproved {
		query += " AND status='" + models.WriteupStatusApproved + "'"
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanWriteups(rows)
}

func (r *WriteupRepository) GetWriteupsByUser(userID string) ([]models.Writeup, error) {
	query := fmt.Sprintf("SELECT %s FROM writeups WHERE user_id=? ORDER BY created_at DESC", r.selectWriteupFields())
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanWriteups(rows)
}

func (r *WriteupRepository) GetAllWriteups() ([]models.Writeup, error) {
	query := fmt.Sprintf("SELECT %s FROM writeups ORDER BY created_at DESC", r.selectWriteupFields())
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanWriteups(rows)
}

func (r *WriteupRepository) UpdateWriteupStatus(id string, status string) error {
	_, err := r.db.Exec("UPDATE writeups SET status=?, updated_at=? WHERE id=?", status, time.Now().Format(time.RFC3339), id)
	return err
}

func (r *WriteupRepository) DeleteWriteup(id string) error {
	_, err := r.db.Exec("DELETE FROM writeups WHERE id=?", id)
	return err
}

func (r *WriteupRepository) FindByUserAndChallenge(userID, challengeID string) (*models.Writeup, error) {
	query := fmt.Sprintf("SELECT %s FROM writeups WHERE user_id=? AND challenge_id=?", r.selectWriteupFields())
	rows, err := r.db.Query(query, userID, challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	writeups, err := r.scanWriteups(rows)
	if err != nil || len(writeups) == 0 {
		return nil, sql.ErrNoRows
	}
	return &writeups[0], nil
}

func (r *WriteupRepository) UpdateWriteupContent(id string, content string, contentFormat string) error {
	_, err := r.db.Exec("UPDATE writeups SET content=?, content_format=?, updated_at=? WHERE id=?", content, contentFormat, time.Now().Format(time.RFC3339), id)
	return err
}

func (r *WriteupRepository) GetWriteupsByTeam(teamID string) ([]models.Writeup, error) {
	query := fmt.Sprintf(`
		SELECT w.id, w.challenge_id, w.user_id, w.username, w.content, w.content_format, w.status, w.upvotes, w.created_at, w.updated_at 
		FROM writeups w
		JOIN team_members tm ON w.user_id = tm.user_id
		WHERE tm.team_id = ? ORDER BY w.created_at DESC
	`)
	rows, err := r.db.Query(query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanWriteups(rows)
}

func (r *WriteupRepository) ToggleUpvote(writeupID string, userID string) (bool, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var exists int
	err = tx.QueryRow("SELECT 1 FROM writeup_upvotes WHERE writeup_id=? AND user_id=?", writeupID, userID).Scan(&exists)

	if err == sql.ErrNoRows {
		// Not upvoted yet, add upvote
		_, err = tx.Exec("INSERT INTO writeup_upvotes (writeup_id, user_id) VALUES (?, ?)", writeupID, userID)
		if err != nil {
			return false, err
		}
		_, err = tx.Exec("UPDATE writeups SET upvotes = upvotes + 1 WHERE id=?", writeupID)
		if err != nil {
			return false, err
		}
		return true, tx.Commit()
	} else if err != nil {
		return false, err
	}

	// Already upvoted, remove upvote
	_, err = tx.Exec("DELETE FROM writeup_upvotes WHERE writeup_id=? AND user_id=?", writeupID, userID)
	if err != nil {
		return false, err
	}
	_, err = tx.Exec("UPDATE writeups SET upvotes = upvotes - 1 WHERE id=?", writeupID)
	if err != nil {
		return false, err
	}
	return false, tx.Commit()
}
