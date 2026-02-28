package models

import "time"

const (
	AchievementFirstBlood     = "first_blood"
	AchievementCategoryMaster = "category_master"
	AchievementStreak         = "streak"
	AchievementSpeedDemon     = "speed_demon"
	AchievementNightOwl       = "night_owl"
)

type Achievement struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	TeamID      string    `json:"team_id,omitempty"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	ChallengeID string    `json:"challenge_id,omitempty"`
	Category    string    `json:"category,omitempty"`
	EarnedAt    time.Time `json:"earned_at"`
}
