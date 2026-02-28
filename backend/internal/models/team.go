package models

import "time"

type Team struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Avatar      string    `json:"avatar,omitempty"`
	LeaderID    string    `json:"leader_id"`
	InviteCode  string    `json:"invite_code"`
	Score       int       `json:"score"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TeamInvitation struct {
	ID            string    `json:"id"`
	TeamID        string    `json:"team_id"`
	TeamName      string    `json:"team_name"`
	InviterID     string    `json:"inviter_id"`
	InviterName   string    `json:"inviter_name"`
	InviteeEmail  string    `json:"invitee_email,omitempty"`
	InviteeUserID string    `json:"invitee_user_id,omitempty"`
	Token         string    `json:"token"`
	Status        string    `json:"status"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
}

const (
	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusRejected = "rejected"
	InvitationStatusExpired  = "expired"
)

const (
	MinTeamSize = 2
	MaxTeamSize = 4
)
