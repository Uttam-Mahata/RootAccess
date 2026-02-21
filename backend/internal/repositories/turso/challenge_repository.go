package turso

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const challengeColumns = `id, title, description, description_format, category, difficulty,
	max_points, min_points, decay, scoring_type, solve_count, flag_hash,
	tags, files, hints, scheduled_at, is_published, contest_id,
	official_writeup, official_writeup_format, official_writeup_published`

// ChallengeRepository implements interfaces.ChallengeRepository using database/sql.
type ChallengeRepository struct {
	db *sql.DB
}

// NewChallengeRepository creates a new Turso-backed ChallengeRepository.
func NewChallengeRepository(db *sql.DB) *ChallengeRepository {
	return &ChallengeRepository{db: db}
}

// scanChallenge scans a single row into a Challenge struct.
func scanChallenge(s scanner) (*models.Challenge, error) {
	var c models.Challenge
	var (
		id                       string
		tagsJSON, filesJSON      string
		hintsJSON                string
		scheduledAt              sql.NullString
		isPublished              int
		contestID                sql.NullString
		officialWriteupPublished int
	)

	err := s.Scan(
		&id, &c.Title, &c.Description, &c.DescriptionFormat, &c.Category, &c.Difficulty,
		&c.MaxPoints, &c.MinPoints, &c.Decay, &c.ScoringType, &c.SolveCount, &c.FlagHash,
		&tagsJSON, &filesJSON, &hintsJSON, &scheduledAt, &isPublished, &contestID,
		&c.OfficialWriteup, &c.OfficialWriteupFormat, &officialWriteupPublished,
	)
	if err != nil {
		return nil, err
	}

	c.ID = oidFromHex(id)
	c.IsPublished = intToBool(isPublished)
	c.OfficialWriteupPublished = intToBool(officialWriteupPublished)
	c.ScheduledAt = strToTimePtr(nullStr(scheduledAt))

	if contestID.Valid && contestID.String != "" {
		oid := oidFromHex(contestID.String)
		c.ContestID = &oid
	}

	// Parse JSON arrays
	if tagsJSON != "" {
		_ = json.Unmarshal([]byte(tagsJSON), &c.Tags)
	}
	if c.Tags == nil {
		c.Tags = []string{}
	}

	if filesJSON != "" {
		_ = json.Unmarshal([]byte(filesJSON), &c.Files)
	}
	if c.Files == nil {
		c.Files = []string{}
	}

	if hintsJSON != "" {
		_ = json.Unmarshal([]byte(hintsJSON), &c.Hints)
	}
	if c.Hints == nil {
		c.Hints = []models.Hint{}
	}

	return &c, nil
}

// contestIDToStr converts a *primitive.ObjectID to a string for storage.
func contestIDToStr(oid *primitive.ObjectID) string {
	if oid == nil || oid.IsZero() {
		return ""
	}
	return oid.Hex()
}

func (r *ChallengeRepository) CreateChallenge(challenge *models.Challenge) error {
	id := newID()

	_, err := r.db.Exec(
		`INSERT INTO challenges (`+challengeColumns+`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		id, challenge.Title, challenge.Description, challenge.DescriptionFormat,
		challenge.Category, challenge.Difficulty,
		challenge.MaxPoints, challenge.MinPoints, challenge.Decay,
		challenge.ScoringType, challenge.SolveCount, challenge.FlagHash,
		jsonMarshal(challenge.Tags), jsonMarshal(challenge.Files), jsonMarshal(challenge.Hints),
		timePtrToStr(challenge.ScheduledAt), boolToInt(challenge.IsPublished),
		contestIDToStr(challenge.ContestID),
		challenge.OfficialWriteup, challenge.OfficialWriteupFormat,
		boolToInt(challenge.OfficialWriteupPublished),
	)
	if err != nil {
		return err
	}
	challenge.ID = oidFromHex(id)
	return nil
}

func (r *ChallengeRepository) GetAllChallenges() ([]models.Challenge, error) {
	rows, err := r.db.Query("SELECT " + challengeColumns + " FROM challenges")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var challenges []models.Challenge
	for rows.Next() {
		c, err := scanChallenge(rows)
		if err != nil {
			return nil, err
		}
		challenges = append(challenges, *c)
	}
	return challenges, rows.Err()
}

func (r *ChallengeRepository) GetAllChallengesForList() ([]models.Challenge, error) {
	challenges, err := r.GetAllChallenges()
	if err != nil {
		return nil, err
	}
	for i := range challenges {
		challenges[i].FlagHash = ""
	}
	return challenges, nil
}

func (r *ChallengeRepository) GetChallengeByID(id string) (*models.Challenge, error) {
	row := r.db.QueryRow("SELECT "+challengeColumns+" FROM challenges WHERE id = ?", id)
	return scanChallenge(row)
}

func (r *ChallengeRepository) UpdateChallenge(id string, challenge *models.Challenge) error {
	_, err := r.db.Exec(
		`UPDATE challenges SET title=?, description=?, description_format=?, category=?, difficulty=?,
		max_points=?, min_points=?, decay=?, scoring_type=?, solve_count=?, flag_hash=?,
		tags=?, files=?, hints=?, scheduled_at=?, is_published=?, contest_id=?,
		official_writeup=?, official_writeup_format=?, official_writeup_published=?
		WHERE id=?`,
		challenge.Title, challenge.Description, challenge.DescriptionFormat,
		challenge.Category, challenge.Difficulty,
		challenge.MaxPoints, challenge.MinPoints, challenge.Decay,
		challenge.ScoringType, challenge.SolveCount, challenge.FlagHash,
		jsonMarshal(challenge.Tags), jsonMarshal(challenge.Files), jsonMarshal(challenge.Hints),
		timePtrToStr(challenge.ScheduledAt), boolToInt(challenge.IsPublished),
		contestIDToStr(challenge.ContestID),
		challenge.OfficialWriteup, challenge.OfficialWriteupFormat,
		boolToInt(challenge.OfficialWriteupPublished),
		id,
	)
	return err
}

func (r *ChallengeRepository) DeleteChallenge(id string) error {
	_, err := r.db.Exec("DELETE FROM challenges WHERE id = ?", id)
	return err
}

func (r *ChallengeRepository) IncrementSolveCount(id string) error {
	_, err := r.db.Exec("UPDATE challenges SET solve_count = solve_count + 1 WHERE id = ?", id)
	return err
}

func (r *ChallengeRepository) GetFlagHash(id string) (string, error) {
	var hash string
	err := r.db.QueryRow("SELECT flag_hash FROM challenges WHERE id = ?", id).Scan(&hash)
	return hash, err
}

func (r *ChallengeRepository) CountChallenges() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM challenges").Scan(&count)
	return count, err
}

func (r *ChallengeRepository) GetChallengesByIDs(ids []primitive.ObjectID, publishedOnly bool) ([]models.Challenge, error) {
	if len(ids) == 0 {
		return []models.Challenge{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, oid := range ids {
		placeholders[i] = "?"
		args[i] = oid.Hex()
	}

	query := "SELECT " + challengeColumns + " FROM challenges WHERE id IN (" + strings.Join(placeholders, ",") + ")"
	if publishedOnly {
		query += " AND is_published = 1"
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var challenges []models.Challenge
	for rows.Next() {
		c, err := scanChallenge(rows)
		if err != nil {
			return nil, err
		}
		challenges = append(challenges, *c)
	}
	return challenges, rows.Err()
}

func (r *ChallengeRepository) UpdateOfficialWriteup(id string, content string, format string) error {
	_, err := r.db.Exec(
		"UPDATE challenges SET official_writeup = ?, official_writeup_format = ? WHERE id = ?",
		content, format, id,
	)
	return err
}

func (r *ChallengeRepository) SetContestID(id string, contestID primitive.ObjectID) error {
	val := ""
	if !contestID.IsZero() {
		val = contestID.Hex()
	}
	_, err := r.db.Exec("UPDATE challenges SET contest_id = ? WHERE id = ?", val, id)
	return err
}

func (r *ChallengeRepository) PublishOfficialWriteup(id string) error {
	_, err := r.db.Exec("UPDATE challenges SET official_writeup_published = 1 WHERE id = ?", id)
	return err
}

// Compile-time interface check.
var _ interface {
	CreateChallenge(challenge *models.Challenge) error
	GetAllChallenges() ([]models.Challenge, error)
	GetAllChallengesForList() ([]models.Challenge, error)
	GetChallengeByID(id string) (*models.Challenge, error)
	UpdateChallenge(id string, challenge *models.Challenge) error
	DeleteChallenge(id string) error
	IncrementSolveCount(id string) error
	GetFlagHash(id string) (string, error)
	CountChallenges() (int64, error)
	GetChallengesByIDs(ids []primitive.ObjectID, publishedOnly bool) ([]models.Challenge, error)
	UpdateOfficialWriteup(id string, content string, format string) error
	SetContestID(id string, contestID primitive.ObjectID) error
	PublishOfficialWriteup(id string) error
} = (*ChallengeRepository)(nil)
