package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ContestConfig represents the CTF competition time configuration
type ContestConfig struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	StartTime           time.Time          `bson:"start_time" json:"start_time"`
	EndTime             time.Time          `bson:"end_time" json:"end_time"`
	FreezeTime          *time.Time         `bson:"freeze_time,omitempty" json:"freeze_time,omitempty"` // Scoreboard freeze time
	Title               string             `bson:"title" json:"title"`
	IsActive            bool               `bson:"is_active" json:"is_active"`
	IsPaused            bool               `bson:"is_paused" json:"is_paused"`                           // Pause submissions
	ScoreboardVisibility string            `bson:"scoreboard_visibility" json:"scoreboard_visibility"` // "public", "private", "hidden"
	UpdatedAt           time.Time          `bson:"updated_at" json:"updated_at"`
}

// ContestStatus represents the current state of the contest
type ContestStatus string

const (
	ContestStatusNotStarted ContestStatus = "not_started"
	ContestStatusRunning    ContestStatus = "running"
	ContestStatusPaused     ContestStatus = "paused"
	ContestStatusEnded      ContestStatus = "ended"
)

// GetStatus returns the current status of the contest
func (c *ContestConfig) GetStatus() ContestStatus {
	now := time.Now()
	if !c.IsActive {
		return ContestStatusNotStarted
	}
	if now.Before(c.StartTime) {
		return ContestStatusNotStarted
	}
	if now.After(c.EndTime) {
		return ContestStatusEnded
	}
	if c.IsPaused {
		return ContestStatusPaused
	}
	return ContestStatusRunning
}

// IsScoreboardFrozen returns true if the scoreboard is currently frozen
func (c *ContestConfig) IsScoreboardFrozen() bool {
	if c.FreezeTime == nil {
		return false
	}
	return time.Now().After(*c.FreezeTime)
}
