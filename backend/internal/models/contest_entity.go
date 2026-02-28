package models

import "time"

type Contest struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Description          string    `json:"description"`
	StartTime            time.Time `json:"start_time"`
	EndTime              time.Time `json:"end_time"`
	FreezeTime           string    `json:"freeze_time,omitempty"`
	ScoreboardVisibility string    `json:"scoreboard_visibility,omitempty"`
	IsActive             bool      `json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (c *Contest) IsRunning(now time.Time) bool {
	return c.IsActive && !now.Before(c.StartTime) && !now.After(c.EndTime)
}

func (c *Contest) HasEnded(now time.Time) bool {
	return c.IsActive && now.After(c.EndTime)
}

func (c *Contest) IsScoreboardFrozen(now time.Time) bool {
	if c.FreezeTime == "" {
		return false
	}
	t, _ := time.Parse(time.RFC3339, c.FreezeTime)
	return now.After(t)
}

func (c *Contest) GetScoreboardVisibility() string {
	if c.ScoreboardVisibility == "" {
		return "public"
	}
	return c.ScoreboardVisibility
}

func (c *Contest) Status(now time.Time) string {
	if c.HasEnded(now) {
		return "ended"
	}
	return "running"
}

type ContestRound struct {
	ID          string    `json:"id"`
	ContestID   string    `json:"contest_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Order       int       `json:"order"`
	VisibleFrom time.Time `json:"visible_from"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (r *ContestRound) IsRoundVisibleAt(now time.Time) bool {
	return !now.Before(r.VisibleFrom) && !now.Before(r.StartTime) && !now.After(r.EndTime)
}

type RoundChallenge struct {
	ID          string    `json:"id"`
	RoundID     string    `json:"round_id"`
	ChallengeID string    `json:"challenge_id"`
	CreatedAt   time.Time `json:"created_at"`
}
