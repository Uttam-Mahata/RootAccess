package turso

import (
	"database/sql"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
)

const contestConfigColumns = `id, contest_id, start_time, end_time, freeze_time, title, is_active, is_paused, scoreboard_visibility, updated_at`

// ContestRepository implements interfaces.ContestRepository using database/sql.
type ContestRepository struct {
	db *sql.DB
}

// NewContestRepository creates a new Turso-backed ContestRepository.
func NewContestRepository(db *sql.DB) *ContestRepository {
	return &ContestRepository{db: db}
}

func scanContestConfig(s scanner) (*models.ContestConfig, error) {
	var c models.ContestConfig
	var (
		id, contestID                    string
		startTime, endTime, freezeTime   string
		isActive, isPaused               int
		scoreboardVis, updatedAt         string
	)

	err := s.Scan(&id, &contestID, &startTime, &endTime, &freezeTime,
		&c.Title, &isActive, &isPaused, &scoreboardVis, &updatedAt)
	if err != nil {
		return nil, err
	}

	c.ID = oidFromHex(id)
	c.ContestID = oidFromHex(contestID)
	c.StartTime = strToTime(startTime)
	c.EndTime = strToTime(endTime)
	c.FreezeTime = strToTimePtr(freezeTime)
	c.IsActive = intToBool(isActive)
	c.IsPaused = intToBool(isPaused)
	c.ScoreboardVisibility = scoreboardVis
	c.UpdatedAt = strToTime(updatedAt)
	return &c, nil
}

func (r *ContestRepository) GetActiveContest() (*models.ContestConfig, error) {
	row := r.db.QueryRow(
		`SELECT ` + contestConfigColumns + ` FROM contest_config ORDER BY updated_at DESC LIMIT 1`,
	)
	return scanContestConfig(row)
}

func (r *ContestRepository) UpsertContest(config *models.ContestConfig) error {
	config.UpdatedAt = time.Now()

	if config.ID.IsZero() {
		id := newID()
		_, err := r.db.Exec(
			`INSERT INTO contest_config (`+contestConfigColumns+`) VALUES (?,?,?,?,?,?,?,?,?,?)`,
			id, config.ContestID.Hex(),
			timeToStr(config.StartTime), timeToStr(config.EndTime), timePtrToStr(config.FreezeTime),
			config.Title, boolToInt(config.IsActive), boolToInt(config.IsPaused),
			config.ScoreboardVisibility, timeToStr(config.UpdatedAt),
		)
		if err != nil {
			return err
		}
		config.ID = oidFromHex(id)
		return nil
	}

	_, err := r.db.Exec(
		`UPDATE contest_config SET contest_id=?, start_time=?, end_time=?, freeze_time=?,
		title=?, is_active=?, is_paused=?, scoreboard_visibility=?, updated_at=? WHERE id=?`,
		config.ContestID.Hex(),
		timeToStr(config.StartTime), timeToStr(config.EndTime), timePtrToStr(config.FreezeTime),
		config.Title, boolToInt(config.IsActive), boolToInt(config.IsPaused),
		config.ScoreboardVisibility, timeToStr(config.UpdatedAt), config.ID.Hex(),
	)
	return err
}
