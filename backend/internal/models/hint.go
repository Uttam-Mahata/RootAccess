package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Hint represents a hint for a challenge that costs points to reveal
type Hint struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ChallengeID primitive.ObjectID `bson:"challenge_id" json:"challenge_id"`
	Content     string             `bson:"content" json:"content"`
	Cost        int                `bson:"cost" json:"cost"` // Points deducted when revealed
	Order       int                `bson:"order" json:"order"` // Display order (1, 2, 3...)
}

// HintReveal tracks which hints have been revealed by which users/teams
type HintReveal struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	HintID      primitive.ObjectID `bson:"hint_id" json:"hint_id"`
	ChallengeID primitive.ObjectID `bson:"challenge_id" json:"challenge_id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	TeamID      primitive.ObjectID `bson:"team_id,omitempty" json:"team_id,omitempty"`
	Cost        int                `bson:"cost" json:"cost"`
}
