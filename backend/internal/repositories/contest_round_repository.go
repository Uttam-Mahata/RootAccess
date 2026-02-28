package repositories

import (
	"database/sql"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/google/uuid"
)

type ContestRoundRepository struct {
	db *sql.DB
}

func NewContestRoundRepository(db *sql.DB) *ContestRoundRepository {
	return &ContestRoundRepository{db: db}
}

func (r *ContestRoundRepository) Create(c *models.ContestRound) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	_, err := r.db.Exec(`INSERT INTO contest_rounds (id, contest_id, name, description, display_order, visible_from, start_time, end_time, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.ContestID, c.Name, c.Description, c.Order, c.VisibleFrom.Format(time.RFC3339), c.StartTime.Format(time.RFC3339), c.EndTime.Format(time.RFC3339), c.CreatedAt.Format(time.RFC3339), c.UpdatedAt.Format(time.RFC3339))
	return err
}

func (r *ContestRoundRepository) Update(c *models.ContestRound) error {
	c.UpdatedAt = time.Now()

	_, err := r.db.Exec(`UPDATE contest_rounds SET contest_id=?, name=?, description=?, display_order=?, visible_from=?, start_time=?, end_time=?, updated_at=? WHERE id=?`,
		c.ContestID, c.Name, c.Description, c.Order, c.VisibleFrom.Format(time.RFC3339), c.StartTime.Format(time.RFC3339), c.EndTime.Format(time.RFC3339), c.UpdatedAt.Format(time.RFC3339), c.ID)
	return err
}

func (r *ContestRoundRepository) scanRounds(rows *sql.Rows) ([]models.ContestRound, error) {
	var rs []models.ContestRound
	for rows.Next() {
		var c models.ContestRound
		var vf, start, end, created, updated string
		if err := rows.Scan(&c.ID, &c.ContestID, &c.Name, &c.Description, &c.Order, &vf, &start, &end, &created, &updated); err != nil {
			return nil, err
		}
		c.VisibleFrom, _ = time.Parse(time.RFC3339, vf)
		c.StartTime, _ = time.Parse(time.RFC3339, start)
		c.EndTime, _ = time.Parse(time.RFC3339, end)
		c.CreatedAt, _ = time.Parse(time.RFC3339, created)
		c.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		rs = append(rs, c)
	}
	return rs, nil
}

func (r *ContestRoundRepository) FindByID(id string) (*models.ContestRound, error) {
	rows, err := r.db.Query("SELECT id, contest_id, name, description, display_order, visible_from, start_time, end_time, created_at, updated_at FROM contest_rounds WHERE id=?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rs, err := r.scanRounds(rows)
	if err != nil || len(rs) == 0 {
		return nil, sql.ErrNoRows
	}
	return &rs[0], nil
}

func (r *ContestRoundRepository) ListByContestID(contestID string) ([]models.ContestRound, error) {
	rows, err := r.db.Query("SELECT id, contest_id, name, description, display_order, visible_from, start_time, end_time, created_at, updated_at FROM contest_rounds WHERE contest_id=? ORDER BY display_order ASC, start_time ASC", contestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRounds(rows)
}

func (r *ContestRoundRepository) GetActiveRounds(contestID string, now time.Time) ([]models.ContestRound, error) {
	rounds, err := r.ListByContestID(contestID)
	if err != nil {
		return nil, err
	}

	var active []models.ContestRound
	for _, round := range rounds {
		if round.IsRoundVisibleAt(now) {
			active = append(active, round)
		}
	}
	return active, nil
}

func (r *ContestRoundRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM contest_rounds WHERE id=?", id)
	return err
}

func (r *ContestRoundRepository) DeleteByContestID(contestID string) error {
	_, err := r.db.Exec("DELETE FROM contest_rounds WHERE contest_id=?", contestID)
	return err
}
