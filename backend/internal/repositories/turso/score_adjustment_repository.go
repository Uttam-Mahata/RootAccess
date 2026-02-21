package turso

import (
	"database/sql"
	"strings"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const scoreAdjustmentColumns = `id, target_type, target_id, delta, reason, created_by, created_at`

// ScoreAdjustmentRepository implements interfaces.ScoreAdjustmentRepository using database/sql.
type ScoreAdjustmentRepository struct {
	db *sql.DB
}

// NewScoreAdjustmentRepository creates a new Turso-backed ScoreAdjustmentRepository.
func NewScoreAdjustmentRepository(db *sql.DB) *ScoreAdjustmentRepository {
	return &ScoreAdjustmentRepository{db: db}
}

func (r *ScoreAdjustmentRepository) Create(adjustment *models.ScoreAdjustment) error {
	id := newID()
	adjustment.CreatedAt = time.Now()

	_, err := r.db.Exec(
		`INSERT INTO score_adjustments (`+scoreAdjustmentColumns+`) VALUES (?,?,?,?,?,?,?)`,
		id, adjustment.TargetType, adjustment.TargetID.Hex(), adjustment.Delta,
		adjustment.Reason, adjustment.CreatedBy.Hex(), timeToStr(adjustment.CreatedAt),
	)
	if err != nil {
		return err
	}
	adjustment.ID = oidFromHex(id)
	return nil
}

func (r *ScoreAdjustmentRepository) GetAdjustmentsForUsers(userIDs []primitive.ObjectID) (map[string]int, error) {
	return r.getAdjustments("user", userIDs)
}

func (r *ScoreAdjustmentRepository) GetAdjustmentsForTeams(teamIDs []primitive.ObjectID) (map[string]int, error) {
	return r.getAdjustments("team", teamIDs)
}

func (r *ScoreAdjustmentRepository) getAdjustments(targetType string, ids []primitive.ObjectID) (map[string]int, error) {
	result := make(map[string]int)
	if len(ids) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, targetType)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id.Hex())
	}

	rows, err := r.db.Query(
		`SELECT target_id, SUM(delta) FROM score_adjustments WHERE target_type = ? AND target_id IN (`+strings.Join(placeholders, ",")+`) GROUP BY target_id`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tid string
		var total int
		if err := rows.Scan(&tid, &total); err != nil {
			return nil, err
		}
		result[tid] = total
	}
	return result, rows.Err()
}
