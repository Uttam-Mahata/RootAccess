package turso

import (
	"database/sql"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
)

const contestRoundColumns = `id, contest_id, name, description, "order", visible_from, start_time, end_time, created_at, updated_at`

// ContestRoundRepository implements interfaces.ContestRoundRepository using database/sql.
type ContestRoundRepository struct {
	db *sql.DB
}

// NewContestRoundRepository creates a new Turso-backed ContestRoundRepository.
func NewContestRoundRepository(db *sql.DB) *ContestRoundRepository {
	return &ContestRoundRepository{db: db}
}

func scanContestRound(s scanner) (*models.ContestRound, error) {
	var cr models.ContestRound
	var (
		id, contestID                                     string
		visibleFrom, startTime, endTime                   string
		createdAt, updatedAt                              string
	)

	err := s.Scan(&id, &contestID, &cr.Name, &cr.Description, &cr.Order,
		&visibleFrom, &startTime, &endTime, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	cr.ID = oidFromHex(id)
	cr.ContestID = oidFromHex(contestID)
	cr.VisibleFrom = strToTime(visibleFrom)
	cr.StartTime = strToTime(startTime)
	cr.EndTime = strToTime(endTime)
	cr.CreatedAt = strToTime(createdAt)
	cr.UpdatedAt = strToTime(updatedAt)
	return &cr, nil
}

func (r *ContestRoundRepository) queryRounds(query string, args ...interface{}) ([]models.ContestRound, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.ContestRound
	for rows.Next() {
		cr, err := scanContestRound(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *cr)
	}
	return results, rows.Err()
}

func (r *ContestRoundRepository) Create(round *models.ContestRound) error {
	id := newID()
	now := time.Now()
	round.CreatedAt = now
	round.UpdatedAt = now

	_, err := r.db.Exec(
		`INSERT INTO contest_rounds (`+contestRoundColumns+`) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		id, round.ContestID.Hex(), round.Name, round.Description, round.Order,
		timeToStr(round.VisibleFrom), timeToStr(round.StartTime), timeToStr(round.EndTime),
		timeToStr(now), timeToStr(now),
	)
	if err != nil {
		return err
	}
	round.ID = oidFromHex(id)
	return nil
}

func (r *ContestRoundRepository) Update(round *models.ContestRound) error {
	round.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		`UPDATE contest_rounds SET contest_id=?, name=?, description=?, "order"=?,
		visible_from=?, start_time=?, end_time=?, updated_at=? WHERE id=?`,
		round.ContestID.Hex(), round.Name, round.Description, round.Order,
		timeToStr(round.VisibleFrom), timeToStr(round.StartTime), timeToStr(round.EndTime),
		timeToStr(round.UpdatedAt), round.ID.Hex(),
	)
	return err
}

func (r *ContestRoundRepository) FindByID(id string) (*models.ContestRound, error) {
	row := r.db.QueryRow(`SELECT `+contestRoundColumns+` FROM contest_rounds WHERE id = ?`, id)
	return scanContestRound(row)
}

func (r *ContestRoundRepository) ListByContestID(contestID string) ([]models.ContestRound, error) {
	return r.queryRounds(
		`SELECT `+contestRoundColumns+` FROM contest_rounds WHERE contest_id = ? ORDER BY "order" ASC`,
		contestID,
	)
}

func (r *ContestRoundRepository) GetActiveRounds(contestID string, now time.Time) ([]models.ContestRound, error) {
	return r.queryRounds(
		`SELECT `+contestRoundColumns+` FROM contest_rounds WHERE contest_id = ? AND start_time <= ? AND end_time >= ? ORDER BY "order" ASC`,
		contestID, timeToStr(now), timeToStr(now),
	)
}

func (r *ContestRoundRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM contest_rounds WHERE id = ?`, id)
	return err
}

func (r *ContestRoundRepository) DeleteByContestID(contestID string) error {
	_, err := r.db.Exec(`DELETE FROM contest_rounds WHERE contest_id = ?`, contestID)
	return err
}
