package repositories

import (
	"database/sql"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/google/uuid"
)

type ContestRepository struct {
	db *sql.DB
}

func NewContestRepository(db *sql.DB) *ContestRepository {
	return &ContestRepository{db: db}
}

func (r *ContestRepository) GetActiveContest() (*models.ContestConfig, error) {
	var c models.ContestConfig
	var start, end, freeze, updated string
	var isActive, isPaused int

	err := r.db.QueryRow("SELECT id, contest_id, start_time, end_time, freeze_time, title, is_active, is_paused, scoreboard_visibility, updated_at FROM contest_config ORDER BY updated_at DESC LIMIT 1").
		Scan(&c.ID, &c.ContestID, &start, &end, &freeze, &c.Title, &isActive, &isPaused, &c.ScoreboardVisibility, &updated)
	if err != nil {
		return nil, err
	}

	c.StartTime, _ = time.Parse(time.RFC3339, start)
	c.EndTime, _ = time.Parse(time.RFC3339, end)
	c.FreezeTime = freeze
	c.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	c.IsActive = isActive == 1
	c.IsPaused = isPaused == 1

	return &c, nil
}

func (r *ContestRepository) UpsertContest(config *models.ContestConfig) error {
	config.UpdatedAt = time.Now()

	isActive := 0
	if config.IsActive {
		isActive = 1
	}
	isPaused := 0
	if config.IsPaused {
		isPaused = 1
	}

	if config.ID == "" {
		config.ID = uuid.New().String()
		_, err := r.db.Exec(`INSERT INTO contest_config (id, contest_id, start_time, end_time, freeze_time, title, is_active, is_paused, scoreboard_visibility, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			config.ID, config.ContestID, config.StartTime.Format(time.RFC3339), config.EndTime.Format(time.RFC3339), config.FreezeTime,
			config.Title, isActive, isPaused, config.ScoreboardVisibility, config.UpdatedAt.Format(time.RFC3339))
		return err
	}

	_, err := r.db.Exec(`UPDATE contest_config SET contest_id=?, start_time=?, end_time=?, freeze_time=?, title=?, is_active=?, is_paused=?, scoreboard_visibility=?, updated_at=? WHERE id=?`,
		config.ContestID, config.StartTime.Format(time.RFC3339), config.EndTime.Format(time.RFC3339), config.FreezeTime,
		config.Title, isActive, isPaused, config.ScoreboardVisibility, config.UpdatedAt.Format(time.RFC3339), config.ID)
	return err
}
