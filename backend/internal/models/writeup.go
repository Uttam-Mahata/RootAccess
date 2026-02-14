package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Writeup represents a user-submitted writeup for a solved challenge
type Writeup struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	ChallengeID   primitive.ObjectID   `bson:"challenge_id" json:"challenge_id"`
	UserID        primitive.ObjectID   `bson:"user_id" json:"user_id"`
	Username      string               `bson:"username" json:"username"`
	Content       string               `bson:"content" json:"content"`               // Content (Markdown or HTML)
	ContentFormat string               `bson:"content_format" json:"content_format"` // "markdown" or "html"
	Status        string               `bson:"status" json:"status"`                 // pending, approved, rejected
	Upvotes       int                  `bson:"upvotes" json:"upvotes"`
	UpvotedBy     []primitive.ObjectID `bson:"upvoted_by,omitempty" json:"upvoted_by,omitempty"`
	CreatedAt     time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time            `bson:"updated_at" json:"updated_at"`
}

// Writeup statuses
const (
	WriteupStatusPending  = "pending"
	WriteupStatusApproved = "approved"
	WriteupStatusRejected = "rejected"
)
