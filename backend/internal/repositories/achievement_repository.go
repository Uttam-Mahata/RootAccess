package repositories

import (
	"database/sql"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/google/uuid"
)

type AchievementRepository struct {
	db *sql.DB
}

func NewAchievementRepository(db *sql.DB) *AchievementRepository {
	return &AchievementRepository{db: db}
}

func (r *AchievementRepository) Create(a *models.Achievement) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	a.EarnedAt = time.Now()

	_, err := r.db.Exec("INSERT INTO achievements (id, user_id, team_id, type, name, description, icon, challenge_id, category, earned_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		a.ID, a.UserID, a.TeamID, a.Type, a.Name, a.Description, a.Icon, a.ChallengeID, a.Category, a.EarnedAt.Format(time.RFC3339))
	return err
}

func (r *AchievementRepository) scanAchievements(rows *sql.Rows) ([]models.Achievement, error) {
	var acts []models.Achievement
	for rows.Next() {
		var a models.Achievement
		var earned string
		if err := rows.Scan(&a.ID, &a.UserID, &a.TeamID, &a.Type, &a.Name, &a.Description, &a.Icon, &a.ChallengeID, &a.Category, &earned); err != nil {
			return nil, err
		}
		a.EarnedAt, _ = time.Parse(time.RFC3339, earned)
		acts = append(acts, a)
	}
	return acts, nil
}

func (r *AchievementRepository) GetByUserID(userID string) ([]models.Achievement, error) {
	rows, err := r.db.Query("SELECT id, user_id, team_id, type, name, description, icon, challenge_id, category, earned_at FROM achievements WHERE user_id=? ORDER BY earned_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanAchievements(rows)
}

func (r *AchievementRepository) GetByTeamID(teamID string) ([]models.Achievement, error) {
	rows, err := r.db.Query("SELECT id, user_id, team_id, type, name, description, icon, challenge_id, category, earned_at FROM achievements WHERE team_id=? ORDER BY earned_at DESC", teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanAchievements(rows)
}

func (r *AchievementRepository) GetByType(achievementType string) ([]models.Achievement, error) {
	rows, err := r.db.Query("SELECT id, user_id, team_id, type, name, description, icon, challenge_id, category, earned_at FROM achievements WHERE type=? ORDER BY earned_at DESC", achievementType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanAchievements(rows)
}

func (r *AchievementRepository) Exists(userID string, achievementType string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM achievements WHERE user_id=? AND type=?", userID, achievementType).Scan(&count)
	return count > 0, err
}

func (r *AchievementRepository) ExistsForCategory(userID string, achievementType string, category string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM achievements WHERE user_id=? AND type=? AND category=?", userID, achievementType, category).Scan(&count)
	return count > 0, err
}
