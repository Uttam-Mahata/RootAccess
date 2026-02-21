package turso

import (
	"database/sql"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const achievementColumns = `id, user_id, team_id, type, name, description, icon, challenge_id, category, earned_at`

// AchievementRepository implements interfaces.AchievementRepository using database/sql.
type AchievementRepository struct {
	db *sql.DB
}

// NewAchievementRepository creates a new Turso-backed AchievementRepository.
func NewAchievementRepository(db *sql.DB) *AchievementRepository {
	return &AchievementRepository{db: db}
}

func scanAchievement(s scanner) (*models.Achievement, error) {
	var a models.Achievement
	var id, userID, teamID, challengeID, earnedAt string

	err := s.Scan(&id, &userID, &teamID, &a.Type, &a.Name, &a.Description,
		&a.Icon, &challengeID, &a.Category, &earnedAt)
	if err != nil {
		return nil, err
	}

	a.ID = oidFromHex(id)
	a.UserID = oidFromHex(userID)
	a.TeamID = oidFromHex(teamID)
	a.ChallengeID = oidFromHex(challengeID)
	a.EarnedAt = strToTime(earnedAt)
	return &a, nil
}

func (r *AchievementRepository) queryAchievements(query string, args ...interface{}) ([]models.Achievement, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.Achievement
	for rows.Next() {
		a, err := scanAchievement(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *a)
	}
	return results, rows.Err()
}

func (r *AchievementRepository) Create(achievement *models.Achievement) error {
	id := newID()
	achievement.EarnedAt = time.Now()

	_, err := r.db.Exec(
		`INSERT INTO achievements (`+achievementColumns+`) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		id, achievement.UserID.Hex(), achievement.TeamID.Hex(),
		achievement.Type, achievement.Name, achievement.Description, achievement.Icon,
		achievement.ChallengeID.Hex(), achievement.Category, timeToStr(achievement.EarnedAt),
	)
	if err != nil {
		return err
	}
	achievement.ID = oidFromHex(id)
	return nil
}

func (r *AchievementRepository) GetByUserID(userID primitive.ObjectID) ([]models.Achievement, error) {
	return r.queryAchievements(
		`SELECT `+achievementColumns+` FROM achievements WHERE user_id = ? ORDER BY earned_at DESC`,
		userID.Hex(),
	)
}

func (r *AchievementRepository) GetByTeamID(teamID primitive.ObjectID) ([]models.Achievement, error) {
	return r.queryAchievements(
		`SELECT `+achievementColumns+` FROM achievements WHERE team_id = ? ORDER BY earned_at DESC`,
		teamID.Hex(),
	)
}

func (r *AchievementRepository) GetByType(achievementType string) ([]models.Achievement, error) {
	return r.queryAchievements(
		`SELECT `+achievementColumns+` FROM achievements WHERE type = ? ORDER BY earned_at DESC`,
		achievementType,
	)
}

func (r *AchievementRepository) Exists(userID primitive.ObjectID, achievementType string) (bool, error) {
	var exists int
	err := r.db.QueryRow(
		`SELECT 1 FROM achievements WHERE user_id = ? AND type = ? LIMIT 1`,
		userID.Hex(), achievementType,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *AchievementRepository) ExistsForCategory(userID primitive.ObjectID, achievementType string, category string) (bool, error) {
	var exists int
	err := r.db.QueryRow(
		`SELECT 1 FROM achievements WHERE user_id = ? AND type = ? AND category = ? LIMIT 1`,
		userID.Hex(), achievementType, category,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
