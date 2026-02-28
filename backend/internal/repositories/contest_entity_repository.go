package repositories

import (
	"database/sql"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/google/uuid"
)

type ContestEntityRepository struct {
	db *sql.DB
}

func NewContestEntityRepository(db *sql.DB) *ContestEntityRepository {
	return &ContestEntityRepository{db: db}
}

func (r *ContestEntityRepository) Create(c *models.Contest) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	isActive := 0
	if c.IsActive {
		isActive = 1
	}

	_, err := r.db.Exec(`INSERT INTO contests (id, name, description, start_time, end_time, freeze_time, scoreboard_visibility, is_active, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.Name, c.Description, c.StartTime.Format(time.RFC3339), c.EndTime.Format(time.RFC3339), c.FreezeTime, c.ScoreboardVisibility, isActive, c.CreatedAt.Format(time.RFC3339), c.UpdatedAt.Format(time.RFC3339))
	return err
}

func (r *ContestEntityRepository) Update(c *models.Contest) error {
	c.UpdatedAt = time.Now()
	isActive := 0
	if c.IsActive {
		isActive = 1
	}

	_, err := r.db.Exec(`UPDATE contests SET name=?, description=?, start_time=?, end_time=?, freeze_time=?, scoreboard_visibility=?, is_active=?, updated_at=? WHERE id=?`,
		c.Name, c.Description, c.StartTime.Format(time.RFC3339), c.EndTime.Format(time.RFC3339), c.FreezeTime, c.ScoreboardVisibility, isActive, c.UpdatedAt.Format(time.RFC3339), c.ID)
	return err
}

func (r *ContestEntityRepository) scanContests(rows *sql.Rows) ([]models.Contest, error) {
	var cs []models.Contest
	for rows.Next() {
		var c models.Contest
		var start, end, created, updated string
		var isActive int
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &start, &end, &c.FreezeTime, &c.ScoreboardVisibility, &isActive, &created, &updated); err != nil {
			return nil, err
		}
		c.StartTime, _ = time.Parse(time.RFC3339, start)
		c.EndTime, _ = time.Parse(time.RFC3339, end)
		c.CreatedAt, _ = time.Parse(time.RFC3339, created)
		c.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		c.IsActive = isActive == 1
		cs = append(cs, c)
	}
	return cs, nil
}

func (r *ContestEntityRepository) FindByID(id string) (*models.Contest, error) {
	rows, err := r.db.Query("SELECT id, name, description, start_time, end_time, freeze_time, scoreboard_visibility, is_active, created_at, updated_at FROM contests WHERE id=?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cs, err := r.scanContests(rows)
	if err != nil || len(cs) == 0 {
		return nil, sql.ErrNoRows
	}
	return &cs[0], nil
}

func (r *ContestEntityRepository) ListAll() ([]models.Contest, error) {
	rows, err := r.db.Query("SELECT id, name, description, start_time, end_time, freeze_time, scoreboard_visibility, is_active, created_at, updated_at FROM contests ORDER BY start_time DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanContests(rows)
}

func (r *ContestEntityRepository) GetScoreboardContests() ([]models.Contest, error) {
	rows, err := r.db.Query("SELECT id, name, description, start_time, end_time, freeze_time, scoreboard_visibility, is_active, created_at, updated_at FROM contests WHERE is_active=1 AND start_time <= ? ORDER BY end_time DESC", time.Now().Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	all, err := r.scanContests(rows)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var running, ended []models.Contest
	for _, c := range all {
		if c.IsRunning(now) {
			running = append(running, c)
		} else {
			ended = append(ended, c)
		}
	}

	result := make([]models.Contest, 0, len(running)+len(ended))
	result = append(result, running...)
	result = append(result, ended...)
	return result, nil
}

func (r *ContestEntityRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM contests WHERE id=?", id)
	return err
}
