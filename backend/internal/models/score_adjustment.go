package models

import "time"

type ScoreAdjustment struct {
	ID         string    `json:"id"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	Delta      int       `json:"delta"`
	Reason     string    `json:"reason,omitempty"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

const (
	ScoreAdjustmentTargetUser = "user"
	ScoreAdjustmentTargetTeam = "team"
)
