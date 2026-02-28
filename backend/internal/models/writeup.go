package models

import "time"

type Writeup struct {
	ID            string    `json:"id"`
	ChallengeID   string    `json:"challenge_id"`
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	Content       string    `json:"content"`
	ContentFormat string    `json:"content_format"`
	Status        string    `json:"status"`
	Upvotes       int       `json:"upvotes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

const (
	WriteupStatusPending  = "pending"
	WriteupStatusApproved = "approved"
	WriteupStatusRejected = "rejected"
)
