package repositories

import (
	"database/sql"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/google/uuid"
)

type SubmissionRepository struct {
	db *sql.DB
}

func NewSubmissionRepository(db *sql.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (r *SubmissionRepository) CreateSubmission(sub *models.Submission) error {
	if sub.ID == "" {
		sub.ID = uuid.New().String()
	}
	sub.Timestamp = time.Now()

	isCorrect := 0
	if sub.IsCorrect {
		isCorrect = 1
	}

	query := `INSERT INTO submissions (id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, sub.ID, sub.UserID, sub.TeamID, sub.ChallengeID, sub.ContestID, sub.Flag, isCorrect, sub.IPAddress, sub.Timestamp.Format(time.RFC3339))
	return err
}

func (r *SubmissionRepository) scanSubmissions(rows *sql.Rows) ([]models.Submission, error) {
	var subs []models.Submission
	for rows.Next() {
		var s models.Submission
		var isCorrect int
		var ts string
		if err := rows.Scan(&s.ID, &s.UserID, &s.TeamID, &s.ChallengeID, &s.ContestID, &s.Flag, &isCorrect, &s.IPAddress, &ts); err != nil {
			return nil, err
		}
		s.IsCorrect = isCorrect == 1
		s.Timestamp, _ = time.Parse(time.RFC3339, ts)
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *SubmissionRepository) FindByChallengeAndUser(challengeID, userID string) (*models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions WHERE challenge_id=? AND user_id=? AND is_correct=1 LIMIT 1"
	rows, err := r.db.Query(query, challengeID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	subs, err := r.scanSubmissions(rows)
	if err != nil || len(subs) == 0 {
		return nil, sql.ErrNoRows
	}
	return &subs[0], nil
}

func (r *SubmissionRepository) FindByChallengeAndTeam(challengeID, teamID string) (*models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions WHERE challenge_id=? AND team_id=? AND is_correct=1 LIMIT 1"
	rows, err := r.db.Query(query, challengeID, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	subs, err := r.scanSubmissions(rows)
	if err != nil || len(subs) == 0 {
		return nil, sql.ErrNoRows
	}
	return &subs[0], nil
}

func (r *SubmissionRepository) FindByChallengeAndUserInContest(challengeID, userID, contestID string) (*models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions WHERE challenge_id=? AND user_id=? AND contest_id=? AND is_correct=1 LIMIT 1"
	rows, err := r.db.Query(query, challengeID, userID, contestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	subs, err := r.scanSubmissions(rows)
	if err != nil || len(subs) == 0 {
		return nil, sql.ErrNoRows
	}
	return &subs[0], nil
}

func (r *SubmissionRepository) FindByChallengeAndTeamInContest(challengeID, teamID, contestID string) (*models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions WHERE challenge_id=? AND team_id=? AND contest_id=? AND is_correct=1 LIMIT 1"
	rows, err := r.db.Query(query, challengeID, teamID, contestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	subs, err := r.scanSubmissions(rows)
	if err != nil || len(subs) == 0 {
		return nil, sql.ErrNoRows
	}
	return &subs[0], nil
}

func (r *SubmissionRepository) GetCorrectSubmissionsByContestAndChallenge(contestID, challengeID string) ([]models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions WHERE contest_id=? AND challenge_id=? AND is_correct=1 ORDER BY timestamp ASC"
	rows, err := r.db.Query(query, contestID, challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanSubmissions(rows)
}

func (r *SubmissionRepository) GetTeamSubmissions(teamID string) ([]models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions WHERE team_id=? AND is_correct=1"
	rows, err := r.db.Query(query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanSubmissions(rows)
}

func (r *SubmissionRepository) GetCorrectSubmissionsByContest(contestID string) ([]models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions WHERE contest_id=? AND is_correct=1"
	rows, err := r.db.Query(query, contestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanSubmissions(rows)
}

func (r *SubmissionRepository) GetCorrectSubmissionsByContestBefore(contestID string, before time.Time) ([]models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions WHERE contest_id=? AND is_correct=1 AND timestamp <= ?"
	rows, err := r.db.Query(query, contestID, before.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanSubmissions(rows)
}

func (r *SubmissionRepository) GetAllCorrectSubmissions() ([]models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions WHERE is_correct=1"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanSubmissions(rows)
}

func (r *SubmissionRepository) GetUserCorrectSubmissions(userID string) ([]models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions WHERE user_id=? AND is_correct=1"
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanSubmissions(rows)
}

func (r *SubmissionRepository) GetUserSubmissionCount(userID string) (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM submissions WHERE user_id=?", userID).Scan(&count)
	return count, err
}

func (r *SubmissionRepository) GetUserCorrectSubmissionCount(userID string) (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM submissions WHERE user_id=? AND is_correct=1", userID).Scan(&count)
	return count, err
}

func (r *SubmissionRepository) CountSubmissions() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM submissions").Scan(&count)
	return count, err
}

func (r *SubmissionRepository) CountCorrectSubmissions() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM submissions WHERE is_correct=1").Scan(&count)
	return count, err
}

func (r *SubmissionRepository) GetAllSubmissions() ([]models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanSubmissions(rows)
}

func (r *SubmissionRepository) GetRecentSubmissions(limit int64) ([]models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions ORDER BY timestamp DESC LIMIT ?"
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanSubmissions(rows)
}

func (r *SubmissionRepository) GetCorrectSubmissionsSince(since time.Time) ([]models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions WHERE is_correct=1 AND timestamp >= ?"
	rows, err := r.db.Query(query, since.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanSubmissions(rows)
}

func (r *SubmissionRepository) GetSubmissionsSince(since time.Time) ([]models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions WHERE timestamp >= ?"
	rows, err := r.db.Query(query, since.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanSubmissions(rows)
}

func (r *SubmissionRepository) GetCorrectSubmissionsByChallenge(challengeID string) ([]models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions WHERE challenge_id=? AND is_correct=1 ORDER BY timestamp ASC"
	rows, err := r.db.Query(query, challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanSubmissions(rows)
}

func (r *SubmissionRepository) GetCorrectSubmissionsBefore(before time.Time) ([]models.Submission, error) {
	query := "SELECT id, user_id, team_id, challenge_id, contest_id, flag, is_correct, ip_address, timestamp FROM submissions WHERE is_correct=1 AND timestamp <= ?"
	rows, err := r.db.Query(query, before.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanSubmissions(rows)
}
