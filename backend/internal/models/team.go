package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Team represents a team in the CTF competition
type Team struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name        string               `bson:"name" json:"name"`
	Description string               `bson:"description" json:"description"`
	Avatar      string               `bson:"avatar,omitempty" json:"avatar,omitempty"`
	LeaderID    primitive.ObjectID   `bson:"leader_id" json:"leader_id"`
	MemberIDs   []primitive.ObjectID `bson:"member_ids" json:"member_ids"`
	InviteCode  string               `bson:"invite_code" json:"invite_code"`
	Score       int                  `bson:"score" json:"score"`
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at" json:"updated_at"`
}

// TeamInvitation represents an invitation to join a team
type TeamInvitation struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TeamID        primitive.ObjectID `bson:"team_id" json:"team_id"`
	TeamName      string             `bson:"team_name" json:"team_name"`
	InviterID     primitive.ObjectID `bson:"inviter_id" json:"inviter_id"`
	InviterName   string             `bson:"inviter_name" json:"inviter_name"`
	InviteeEmail  string             `bson:"invitee_email,omitempty" json:"invitee_email,omitempty"`
	InviteeUserID primitive.ObjectID `bson:"invitee_user_id,omitempty" json:"invitee_user_id,omitempty"`
	Token         string             `bson:"token" json:"token"`
	Status        string             `bson:"status" json:"status"` // pending, accepted, rejected, expired
	ExpiresAt     time.Time          `bson:"expires_at" json:"expires_at"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}

// Invitation statuses
const (
	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusRejected = "rejected"
	InvitationStatusExpired  = "expired"
)

// Team size constants
const (
	MinTeamSize = 2
	MaxTeamSize = 4
)
