package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/google/uuid"
)

type ScoreAdjustmentRepository struct {
	db *sql.DB
}

func NewScoreAdjustmentRepository(db *sql.DB) *ScoreAdjustmentRepository {
	return &ScoreAdjustmentRepository{db: db}
}

func (r *ScoreAdjustmentRepository) Create(adj *models.ScoreAdjustment) error {
	if adj.ID == "" {
		adj.ID = uuid.New().String()
	}
	adj.CreatedAt = time.Now()

	query := `INSERT INTO score_adjustments (id, target_type, target_id, delta, reason, created_by, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, adj.ID, adj.TargetType, adj.TargetID, adj.Delta, adj.Reason, adj.CreatedBy, adj.CreatedAt.Format(time.RFC3339))
	return err
}

func (r *ScoreAdjustmentRepository) GetAdjustmentsForUsers(userIDs []string) (map[string]int, error) {
	return r.getAdjustmentsByTargets(models.ScoreAdjustmentTargetUser, userIDs)
}

func (r *ScoreAdjustmentRepository) GetAdjustmentsForTeams(teamIDs []string) (map[string]int, error) {
	return r.getAdjustmentsByTargets(models.ScoreAdjustmentTargetTeam, teamIDs)
}

func (r *ScoreAdjustmentRepository) getAdjustmentsByTargets(targetType string, ids []string) (map[string]int, error) {
	result := make(map[string]int)
	if len(ids) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+1)
	args[0] = targetType
	for i, id := range ids {
		placeholders[i] = "?"
		args[i+1] = id
	}

	query := fmt.Sprintf(`SELECT target_id, SUM(delta) FROM score_adjustments 
						  WHERE target_type=? AND target_id IN (%s) GROUP BY target_id`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var targetID string
		var sum int
		if err := rows.Scan(&targetID, &sum); err != nil {
			return nil, err
		}
		result[targetID] = sum
	}

	return result, nil
}
