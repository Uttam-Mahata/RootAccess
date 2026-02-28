package repositories

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/google/uuid"
)

type ChallengeRepository struct {
	db *sql.DB
}

func NewChallengeRepository(db *sql.DB) *ChallengeRepository {
	return &ChallengeRepository{db: db}
}

func (r *ChallengeRepository) CreateChallenge(challenge *models.Challenge) error {
	if challenge.ID == "" {
		challenge.ID = uuid.New().String()
	}
	challenge.SolveCount = 0

	filesJSON, _ := json.Marshal(challenge.Files)
	tagsJSON, _ := json.Marshal(challenge.Tags)

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO challenges (
			id, title, description, description_format, category, difficulty,
			max_points, min_points, decay, scoring_type, solve_count, flag_hash,
			files, tags, scheduled_at, is_published, contest_id, official_writeup,
			official_writeup_format, official_writeup_published
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	isPublished := 0
	if challenge.IsPublished {
		isPublished = 1
	}

	owPublished := 0
	if challenge.OfficialWriteupPublished {
		owPublished = 1
	}

	_, err = tx.Exec(query,
		challenge.ID, challenge.Title, challenge.Description, challenge.DescriptionFormat,
		challenge.Category, challenge.Difficulty, challenge.MaxPoints, challenge.MinPoints,
		challenge.Decay, challenge.ScoringType, challenge.SolveCount, challenge.FlagHash,
		string(filesJSON), string(tagsJSON), challenge.ScheduledAt, isPublished,
		challenge.ContestID, challenge.OfficialWriteup, challenge.OfficialWriteupFormat,
		owPublished,
	)
	if err != nil {
		return err
	}

	for _, hint := range challenge.Hints {
		hintID := hint.ID
		if hintID == "" {
			hintID = uuid.New().String()
		}
		_, err = tx.Exec("INSERT INTO hints (id, challenge_id, content, cost, display_order) VALUES (?, ?, ?, ?, ?)",
			hintID, challenge.ID, hint.Content, hint.Cost, hint.Order)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ChallengeRepository) scanChallenge(row *sql.Row) (*models.Challenge, error) {
	var c models.Challenge
	var filesJSON, tagsJSON string
	var isPub, owPub int

	err := row.Scan(
		&c.ID, &c.Title, &c.Description, &c.DescriptionFormat,
		&c.Category, &c.Difficulty, &c.MaxPoints, &c.MinPoints,
		&c.Decay, &c.ScoringType, &c.SolveCount, &c.FlagHash,
		&filesJSON, &tagsJSON, &c.ScheduledAt, &isPub,
		&c.ContestID, &c.OfficialWriteup, &c.OfficialWriteupFormat,
		&owPub,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("challenge not found")
		}
		return nil, err
	}

	c.IsPublished = isPub == 1
	c.OfficialWriteupPublished = owPub == 1
	if filesJSON != "" {
		json.Unmarshal([]byte(filesJSON), &c.Files)
	}
	if tagsJSON != "" {
		json.Unmarshal([]byte(tagsJSON), &c.Tags)
	}

	c.Hints, _ = r.getHints(c.ID)
	return &c, nil
}

func (r *ChallengeRepository) scanChallenges(rows *sql.Rows) ([]models.Challenge, error) {
	var challenges []models.Challenge
	for rows.Next() {
		var c models.Challenge
		var filesJSON, tagsJSON string
		var isPub, owPub int

		if err := rows.Scan(
			&c.ID, &c.Title, &c.Description, &c.DescriptionFormat,
			&c.Category, &c.Difficulty, &c.MaxPoints, &c.MinPoints,
			&c.Decay, &c.ScoringType, &c.SolveCount, &c.FlagHash,
			&filesJSON, &tagsJSON, &c.ScheduledAt, &isPub,
			&c.ContestID, &c.OfficialWriteup, &c.OfficialWriteupFormat,
			&owPub,
		); err != nil {
			return nil, err
		}

		c.IsPublished = isPub == 1
		c.OfficialWriteupPublished = owPub == 1
		if filesJSON != "" {
			json.Unmarshal([]byte(filesJSON), &c.Files)
		}
		if tagsJSON != "" {
			json.Unmarshal([]byte(tagsJSON), &c.Tags)
		}

		c.Hints, _ = r.getHints(c.ID)
		challenges = append(challenges, c)
	}
	return challenges, nil
}

func (r *ChallengeRepository) selectChallengeFields() string {
	return `id, title, description, description_format, category, difficulty,
			max_points, min_points, decay, scoring_type, solve_count, flag_hash,
			files, tags, scheduled_at, is_published, contest_id, official_writeup,
			official_writeup_format, official_writeup_published`
}

func (r *ChallengeRepository) getHints(challengeID string) ([]models.Hint, error) {
	rows, err := r.db.Query("SELECT id, content, cost, display_order FROM hints WHERE challenge_id=? ORDER BY display_order", challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hints []models.Hint
	for rows.Next() {
		var h models.Hint
		if err := rows.Scan(&h.ID, &h.Content, &h.Cost, &h.Order); err != nil {
			return nil, err
		}
		h.ChallengeID = challengeID
		hints = append(hints, h)
	}
	return hints, nil
}

func (r *ChallengeRepository) GetAllChallenges() ([]models.Challenge, error) {
	query := fmt.Sprintf("SELECT %s FROM challenges", r.selectChallengeFields())
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanChallenges(rows)
}

func (r *ChallengeRepository) GetAllChallengesForList() ([]models.Challenge, error) {
	query := `SELECT id, title, '', description_format, category, difficulty,
			max_points, min_points, decay, scoring_type, solve_count, flag_hash,
			files, tags, scheduled_at, is_published, contest_id, official_writeup,
			official_writeup_format, official_writeup_published FROM challenges`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanChallenges(rows)
}

func (r *ChallengeRepository) GetChallengeByID(id string) (*models.Challenge, error) {
	query := fmt.Sprintf("SELECT %s FROM challenges WHERE id=?", r.selectChallengeFields())
	return r.scanChallenge(r.db.QueryRow(query, id))
}

func (r *ChallengeRepository) UpdateChallenge(id string, challenge *models.Challenge) error {
	filesJSON, _ := json.Marshal(challenge.Files)
	tagsJSON, _ := json.Marshal(challenge.Tags)

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE challenges SET
			title=?, description=?, description_format=?, category=?, difficulty=?,
			max_points=?, min_points=?, decay=?, scoring_type=?, flag_hash=?,
			files=?, tags=?
		WHERE id=?
	`
	_, err = tx.Exec(query,
		challenge.Title, challenge.Description, challenge.DescriptionFormat, challenge.Category,
		challenge.Difficulty, challenge.MaxPoints, challenge.MinPoints, challenge.Decay,
		challenge.ScoringType, challenge.FlagHash, string(filesJSON), string(tagsJSON), id,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM hints WHERE challenge_id=?", id)
	if err != nil {
		return err
	}

	for _, hint := range challenge.Hints {
		hintID := hint.ID
		if hintID == "" {
			hintID = uuid.New().String()
		}
		_, err = tx.Exec("INSERT INTO hints (id, challenge_id, content, cost, display_order) VALUES (?, ?, ?, ?, ?)",
			hintID, id, hint.Content, hint.Cost, hint.Order)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ChallengeRepository) DeleteChallenge(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Turso/libSQL may not support ON DELETE CASCADE; delete dependent rows explicitly
	// Order matters: delete children before parents
	queries := []string{
		"DELETE FROM hint_reveals WHERE challenge_id=?",
		"DELETE FROM hints WHERE challenge_id=?",
		"DELETE FROM writeup_upvotes WHERE writeup_id IN (SELECT id FROM writeups WHERE challenge_id=?)",
		"DELETE FROM writeups WHERE challenge_id=?",
		"UPDATE achievements SET challenge_id=NULL WHERE challenge_id=?",
		"DELETE FROM round_challenges WHERE challenge_id=?",
		"DELETE FROM contest_challenge_solves WHERE challenge_id=?",
		"DELETE FROM submissions WHERE challenge_id=?",
		"DELETE FROM challenges WHERE id=?",
	}
	for _, q := range queries {
		if _, err := tx.Exec(q, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *ChallengeRepository) IncrementSolveCount(id string) error {
	_, err := r.db.Exec("UPDATE challenges SET solve_count = solve_count + 1 WHERE id=?", id)
	return err
}

func (r *ChallengeRepository) GetFlagHash(id string) (string, error) {
	var hash string
	err := r.db.QueryRow("SELECT flag_hash FROM challenges WHERE id=?", id).Scan(&hash)
	return hash, err
}

func (r *ChallengeRepository) CountChallenges() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM challenges").Scan(&count)
	return count, err
}

func (r *ChallengeRepository) GetChallengesByIDs(ids []string, publishedOnly bool) ([]models.Challenge, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("SELECT %s FROM challenges WHERE id IN (%s)", r.selectChallengeFields(), strings.Join(placeholders, ","))
	if publishedOnly {
		query += " AND is_published=1"
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanChallenges(rows)
}

func (r *ChallengeRepository) UpdateOfficialWriteup(id string, content, format string) error {
	_, err := r.db.Exec("UPDATE challenges SET official_writeup=?, official_writeup_format=? WHERE id=?", content, format, id)
	return err
}

func (r *ChallengeRepository) SetContestID(id string, contestID string) error {
	_, err := r.db.Exec("UPDATE challenges SET contest_id=? WHERE id=?", contestID, id)
	return err
}

func (r *ChallengeRepository) PublishOfficialWriteup(id string) error {
	_, err := r.db.Exec("UPDATE challenges SET official_writeup_published=1 WHERE id=?", id)
	return err
}
