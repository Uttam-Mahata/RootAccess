package models

import "time"

type TeamContestRegistration struct {
	ID           string    `json:"id"`
	TeamID       string    `json:"team_id"`
	ContestID    string    `json:"contest_id"`
	RegisteredAt time.Time `json:"registered_at"`
}
