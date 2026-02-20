package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TeamContestRegistration represents a team's registration for a contest
type TeamContestRegistration struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TeamID    primitive.ObjectID `bson:"team_id" json:"team_id"`
	ContestID primitive.ObjectID `bson:"contest_id" json:"contest_id"`
	RegisteredAt time.Time       `bson:"registered_at" json:"registered_at"`
}
