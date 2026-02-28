package models

import "time"

type ContestConfig struct {
	ID                   string    `json:"id"`
	ContestID            string    `json:"contest_id,omitempty"`
	StartTime            time.Time `json:"start_time"`
	EndTime              time.Time `json:"end_time"`
	FreezeTime           string    `json:"freeze_time,omitempty"`
	Title                string    `json:"title"`
	IsActive             bool      `json:"is_active"`
	IsPaused             bool      `json:"is_paused"`
	ScoreboardVisibility string    `json:"scoreboard_visibility"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ContestStatus string

const (
	ContestStatusNotStarted ContestStatus = "not_started"
	ContestStatusRunning    ContestStatus = "running"
	ContestStatusPaused     ContestStatus = "paused"
	ContestStatusEnded      ContestStatus = "ended"
)

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

func (c *ContestConfig) IsScoreboardFrozen() bool {
	if c.FreezeTime == "" {
		return false
	}
	t, _ := time.Parse(time.RFC3339, c.FreezeTime)
	return time.Now().After(t)
}
