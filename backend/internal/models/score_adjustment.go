package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ScoreAdjustment represents a manual score change applied by an admin.
// It is used to adjust computed scores for users or teams without
// faking submissions.
type ScoreAdjustment struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TargetType string             `bson:"target_type" json:"target_type"` // "user" or "team"
	TargetID   primitive.ObjectID `bson:"target_id" json:"target_id"`
	Delta      int                `bson:"delta" json:"delta"`
	Reason     string             `bson:"reason,omitempty" json:"reason,omitempty"`
	CreatedBy  primitive.ObjectID `bson:"created_by" json:"created_by"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

const (
	ScoreAdjustmentTargetUser = "user"
	ScoreAdjustmentTargetTeam = "team"
)

