package models

import "time"

type Submission struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	TeamID      string    `json:"team_id,omitempty"`
	ChallengeID string    `json:"challenge_id"`
	ContestID   string    `json:"contest_id,omitempty"`
	Flag        string    `json:"flag"`
	IsCorrect   bool      `json:"is_correct"`
	IPAddress   string    `json:"ip_address,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}
