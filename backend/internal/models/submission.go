package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Submission struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	ChallengeID primitive.ObjectID `bson:"challenge_id" json:"challenge_id"`
	Flag        string             `bson:"flag" json:"flag"`
	IsCorrect   bool               `bson:"is_correct" json:"is_correct"`
	Timestamp   time.Time          `bson:"timestamp" json:"timestamp"`
}
