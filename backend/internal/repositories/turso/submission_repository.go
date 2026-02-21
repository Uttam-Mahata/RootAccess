package turso

import (
	"database/sql"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const submissionColumns = `id, user_id, team_id, challenge_id, flag, is_correct, ip_address, timestamp`

// SubmissionRepository implements interfaces.SubmissionRepository using database/sql.
type SubmissionRepository struct {
	db *sql.DB
}

// NewSubmissionRepository creates a new Turso-backed SubmissionRepository.
func NewSubmissionRepository(db *sql.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

// scanSubmission scans a single row into a Submission struct.
func scanSubmission(s scanner) (*models.Submission, error) {
	var sub models.Submission
	var (
		id, userID, teamID, challengeID string
		isCorrect                       int
		ipAddress                       sql.NullString
		ts                              string
	)

	err := s.Scan(&id, &userID, &teamID, &challengeID, &sub.Flag, &isCorrect, &ipAddress, &ts)
	if err != nil {
		return nil, err
	}

	sub.ID = oidFromHex(id)
	sub.UserID = oidFromHex(userID)
	sub.TeamID = oidFromHex(teamID)
	sub.ChallengeID = oidFromHex(challengeID)
	sub.IsCorrect = intToBool(isCorrect)
	sub.IPAddress = nullStr(ipAddress)
	sub.Timestamp = strToTime(ts)

	return &sub, nil
}

// querySubmissions executes a query and returns all matching submissions.
func (r *SubmissionRepository) querySubmissions(query string, args ...interface{}) ([]models.Submission, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.Submission
	for rows.Next() {
		sub, err := scanSubmission(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *sub)
	}
	return results, rows.Err()
}

func (r *SubmissionRepository) CreateSubmission(submission *models.Submission) error {
	submission.ID = oidFromHex(newID())
	submission.Timestamp = time.Now()

	_, err := r.db.Exec(
		`INSERT INTO submissions (`+submissionColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		submission.ID.Hex(),
		submission.UserID.Hex(),
		submission.TeamID.Hex(),
		submission.ChallengeID.Hex(),
		submission.Flag,
		boolToInt(submission.IsCorrect),
		submission.IPAddress,
		timeToStr(submission.Timestamp),
	)
	return err
}

func (r *SubmissionRepository) FindByChallengeAndUser(challengeID, userID primitive.ObjectID) (*models.Submission, error) {
	row := r.db.QueryRow(
		`SELECT `+submissionColumns+` FROM submissions WHERE challenge_id = ? AND user_id = ? AND is_correct = 1 LIMIT 1`,
		challengeID.Hex(), userID.Hex(),
	)
	return scanSubmission(row)
}

func (r *SubmissionRepository) FindByChallengeAndTeam(challengeID, teamID primitive.ObjectID) (*models.Submission, error) {
	row := r.db.QueryRow(
		`SELECT `+submissionColumns+` FROM submissions WHERE challenge_id = ? AND team_id = ? AND is_correct = 1 LIMIT 1`,
		challengeID.Hex(), teamID.Hex(),
	)
	return scanSubmission(row)
}

func (r *SubmissionRepository) GetTeamSubmissions(teamID primitive.ObjectID) ([]models.Submission, error) {
	return r.querySubmissions(
		`SELECT `+submissionColumns+` FROM submissions WHERE team_id = ?`,
		teamID.Hex(),
	)
}

func (r *SubmissionRepository) GetAllCorrectSubmissions() ([]models.Submission, error) {
	return r.querySubmissions(
		`SELECT ` + submissionColumns + ` FROM submissions WHERE is_correct = 1`,
	)
}

func (r *SubmissionRepository) GetUserCorrectSubmissions(userID primitive.ObjectID) ([]models.Submission, error) {
	return r.querySubmissions(
		`SELECT `+submissionColumns+` FROM submissions WHERE user_id = ? AND is_correct = 1`,
		userID.Hex(),
	)
}

func (r *SubmissionRepository) GetUserSubmissionCount(userID primitive.ObjectID) (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM submissions WHERE user_id = ?`, userID.Hex()).Scan(&count)
	return count, err
}

func (r *SubmissionRepository) GetUserCorrectSubmissionCount(userID primitive.ObjectID) (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM submissions WHERE user_id = ? AND is_correct = 1`, userID.Hex()).Scan(&count)
	return count, err
}

func (r *SubmissionRepository) CountSubmissions() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM submissions`).Scan(&count)
	return count, err
}

func (r *SubmissionRepository) CountCorrectSubmissions() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM submissions WHERE is_correct = 1`).Scan(&count)
	return count, err
}

func (r *SubmissionRepository) GetAllSubmissions() ([]models.Submission, error) {
	return r.querySubmissions(
		`SELECT ` + submissionColumns + ` FROM submissions`,
	)
}

func (r *SubmissionRepository) GetRecentSubmissions(limit int64) ([]models.Submission, error) {
	return r.querySubmissions(
		`SELECT `+submissionColumns+` FROM submissions ORDER BY timestamp DESC LIMIT ?`,
		limit,
	)
}

func (r *SubmissionRepository) GetCorrectSubmissionsSince(since time.Time) ([]models.Submission, error) {
	return r.querySubmissions(
		`SELECT `+submissionColumns+` FROM submissions WHERE is_correct = 1 AND timestamp >= ?`,
		timeToStr(since),
	)
}

func (r *SubmissionRepository) GetSubmissionsSince(since time.Time) ([]models.Submission, error) {
	return r.querySubmissions(
		`SELECT `+submissionColumns+` FROM submissions WHERE timestamp >= ?`,
		timeToStr(since),
	)
}

func (r *SubmissionRepository) GetCorrectSubmissionsByChallenge(challengeID primitive.ObjectID) ([]models.Submission, error) {
	return r.querySubmissions(
		`SELECT `+submissionColumns+` FROM submissions WHERE challenge_id = ? AND is_correct = 1 ORDER BY timestamp ASC`,
		challengeID.Hex(),
	)
}

func (r *SubmissionRepository) GetCorrectSubmissionsBefore(before time.Time) ([]models.Submission, error) {
	return r.querySubmissions(
		`SELECT `+submissionColumns+` FROM submissions WHERE is_correct = 1 AND timestamp < ?`,
		timeToStr(before),
	)
}

// Compile-time interface check.
var _ interface {
	CreateSubmission(submission *models.Submission) error
	FindByChallengeAndUser(challengeID, userID primitive.ObjectID) (*models.Submission, error)
	FindByChallengeAndTeam(challengeID, teamID primitive.ObjectID) (*models.Submission, error)
	GetTeamSubmissions(teamID primitive.ObjectID) ([]models.Submission, error)
	GetAllCorrectSubmissions() ([]models.Submission, error)
	GetUserCorrectSubmissions(userID primitive.ObjectID) ([]models.Submission, error)
	GetUserSubmissionCount(userID primitive.ObjectID) (int64, error)
	GetUserCorrectSubmissionCount(userID primitive.ObjectID) (int64, error)
	CountSubmissions() (int64, error)
	CountCorrectSubmissions() (int64, error)
	GetAllSubmissions() ([]models.Submission, error)
	GetRecentSubmissions(limit int64) ([]models.Submission, error)
	GetCorrectSubmissionsSince(since time.Time) ([]models.Submission, error)
	GetSubmissionsSince(since time.Time) ([]models.Submission, error)
	GetCorrectSubmissionsByChallenge(challengeID primitive.ObjectID) ([]models.Submission, error)
	GetCorrectSubmissionsBefore(before time.Time) ([]models.Submission, error)
} = (*SubmissionRepository)(nil)
