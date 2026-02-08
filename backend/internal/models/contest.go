package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ContestConfig represents the CTF competition time configuration
type ContestConfig struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	StartTime time.Time          `bson:"start_time" json:"start_time"`
	EndTime   time.Time          `bson:"end_time" json:"end_time"`
	Title     string             `bson:"title" json:"title"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// ContestStatus represents the current state of the contest
type ContestStatus string

const (
	ContestStatusNotStarted ContestStatus = "not_started"
	ContestStatusRunning    ContestStatus = "running"
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
	return ContestStatusRunning
}
