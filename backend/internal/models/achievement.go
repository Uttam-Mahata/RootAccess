package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Achievement types
const (
	AchievementFirstBlood     = "first_blood"
	AchievementCategoryMaster = "category_master"
	AchievementStreak         = "streak"
	AchievementSpeedDemon     = "speed_demon"
	AchievementNightOwl       = "night_owl"
)

// Achievement represents a badge/achievement earned by a user or team
type Achievement struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	TeamID      primitive.ObjectID `bson:"team_id,omitempty" json:"team_id,omitempty"`
	Type        string             `bson:"type" json:"type"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Icon        string             `bson:"icon" json:"icon"`
	ChallengeID primitive.ObjectID `bson:"challenge_id,omitempty" json:"challenge_id,omitempty"`
	Category    string             `bson:"category,omitempty" json:"category,omitempty"`
	EarnedAt    time.Time          `bson:"earned_at" json:"earned_at"`
}
