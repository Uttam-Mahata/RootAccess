package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Contest represents a CTF competition event (separate from ContestConfig which is the active selector)
type Contest struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name                 string             `bson:"name" json:"name"`
	Description          string             `bson:"description" json:"description"`
	StartTime            time.Time          `bson:"start_time" json:"start_time"`
	EndTime              time.Time          `bson:"end_time" json:"end_time"`
	FreezeTime           *time.Time         `bson:"freeze_time,omitempty" json:"freeze_time,omitempty"`
	ScoreboardVisibility string             `bson:"scoreboard_visibility,omitempty" json:"scoreboard_visibility,omitempty"` // "public", "private", "hidden"
	IsActive             bool               `bson:"is_active" json:"is_active"`
	CreatedAt            time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time          `bson:"updated_at" json:"updated_at"`
}

// IsRunning returns true if the contest is currently running
func (c *Contest) IsRunning(now time.Time) bool {
	return c.IsActive && !now.Before(c.StartTime) && !now.After(c.EndTime)
}

// HasEnded returns true if the contest has ended
func (c *Contest) HasEnded(now time.Time) bool {
	return c.IsActive && now.After(c.EndTime)
}

// IsScoreboardFrozen returns true if the contest scoreboard is currently frozen
func (c *Contest) IsScoreboardFrozen(now time.Time) bool {
	if c.FreezeTime == nil {
		return false
	}
	return now.After(*c.FreezeTime)
}

// GetScoreboardVisibility returns the scoreboard visibility, defaulting to "public"
func (c *Contest) GetScoreboardVisibility() string {
	if c.ScoreboardVisibility == "" {
		return "public"
	}
	return c.ScoreboardVisibility
}

// Status returns "running" or "ended" based on current time
func (c *Contest) Status(now time.Time) string {
	if c.HasEnded(now) {
		return "ended"
	}
	return "running"
}

// ContestRound represents a timed round within a contest
type ContestRound struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ContestID   primitive.ObjectID `bson:"contest_id" json:"contest_id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Order       int                `bson:"order" json:"order"` // display/sequence order
	VisibleFrom time.Time          `bson:"visible_from" json:"visible_from"`
	StartTime   time.Time          `bson:"start_time" json:"start_time"`
	EndTime     time.Time          `bson:"end_time" json:"end_time"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// IsRoundVisibleAt returns true if the round is visible and active at the given time
func (r *ContestRound) IsRoundVisibleAt(now time.Time) bool {
	return !now.Before(r.VisibleFrom) && !now.Before(r.StartTime) && !now.After(r.EndTime)
}

// RoundChallenge represents a challenge assigned to a round (join collection)
type RoundChallenge struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoundID     primitive.ObjectID `bson:"round_id" json:"round_id"`
	ChallengeID primitive.ObjectID `bson:"challenge_id" json:"challenge_id"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}
