package turso

import (
	"database/sql"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
)

const contestEntityColumns = `id, name, description, start_time, end_time, freeze_time, scoreboard_visibility, is_active, created_at, updated_at`

// ContestEntityRepository implements interfaces.ContestEntityRepository using database/sql.
type ContestEntityRepository struct {
	db *sql.DB
}

// NewContestEntityRepository creates a new Turso-backed ContestEntityRepository.
func NewContestEntityRepository(db *sql.DB) *ContestEntityRepository {
	return &ContestEntityRepository{db: db}
}

func scanContest(s scanner) (*models.Contest, error) {
	var c models.Contest
	var (
		id                                   string
		startTime, endTime, freezeTime       string
		isActive                             int
		createdAt, updatedAt                 string
	)

	err := s.Scan(&id, &c.Name, &c.Description, &startTime, &endTime, &freezeTime,
		&c.ScoreboardVisibility, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	c.ID = oidFromHex(id)
	c.StartTime = strToTime(startTime)
	c.EndTime = strToTime(endTime)
	c.FreezeTime = strToTimePtr(freezeTime)
	c.IsActive = intToBool(isActive)
	c.CreatedAt = strToTime(createdAt)
	c.UpdatedAt = strToTime(updatedAt)
	return &c, nil
}

func (r *ContestEntityRepository) queryContests(query string, args ...interface{}) ([]models.Contest, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.Contest
	for rows.Next() {
		c, err := scanContest(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *c)
	}
	return results, rows.Err()
}

func (r *ContestEntityRepository) Create(contest *models.Contest) error {
	id := newID()
	now := time.Now()
	contest.CreatedAt = now
	contest.UpdatedAt = now

	_, err := r.db.Exec(
		`INSERT INTO contests (`+contestEntityColumns+`) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		id, contest.Name, contest.Description,
		timeToStr(contest.StartTime), timeToStr(contest.EndTime), timePtrToStr(contest.FreezeTime),
		contest.ScoreboardVisibility, boolToInt(contest.IsActive),
		timeToStr(now), timeToStr(now),
	)
	if err != nil {
		return err
	}
	contest.ID = oidFromHex(id)
	return nil
}

func (r *ContestEntityRepository) Update(contest *models.Contest) error {
	contest.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		`UPDATE contests SET name=?, description=?, start_time=?, end_time=?, freeze_time=?,
		scoreboard_visibility=?, is_active=?, updated_at=? WHERE id=?`,
		contest.Name, contest.Description,
		timeToStr(contest.StartTime), timeToStr(contest.EndTime), timePtrToStr(contest.FreezeTime),
		contest.ScoreboardVisibility, boolToInt(contest.IsActive),
		timeToStr(contest.UpdatedAt), contest.ID.Hex(),
	)
	return err
}

func (r *ContestEntityRepository) FindByID(id string) (*models.Contest, error) {
	row := r.db.QueryRow(`SELECT `+contestEntityColumns+` FROM contests WHERE id = ?`, id)
	return scanContest(row)
}

func (r *ContestEntityRepository) ListAll() ([]models.Contest, error) {
	return r.queryContests(`SELECT ` + contestEntityColumns + ` FROM contests ORDER BY created_at DESC`)
}

func (r *ContestEntityRepository) GetScoreboardContests() ([]models.Contest, error) {
	return r.queryContests(`SELECT ` + contestEntityColumns + ` FROM contests WHERE is_active = 1 ORDER BY start_time DESC`)
}

func (r *ContestEntityRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM contests WHERE id = ?`, id)
	return err
}
