package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Submission struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID  `bson:"user_id" json:"user_id"`
	TeamID      primitive.ObjectID  `bson:"team_id,omitempty" json:"team_id,omitempty"`
	ChallengeID primitive.ObjectID  `bson:"challenge_id" json:"challenge_id"`
	ContestID   *primitive.ObjectID `bson:"contest_id,omitempty" json:"contest_id,omitempty"`
	Flag        string              `bson:"flag" json:"flag"`
	IsCorrect   bool                `bson:"is_correct" json:"is_correct"`
	IPAddress   string              `bson:"ip_address,omitempty" json:"ip_address,omitempty"`
	Timestamp   time.Time           `bson:"timestamp" json:"timestamp"`
}
