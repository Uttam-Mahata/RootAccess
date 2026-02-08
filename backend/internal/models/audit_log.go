package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuditLog tracks admin actions and important events
type AuditLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Username  string             `bson:"username" json:"username"`
	Action    string             `bson:"action" json:"action"`     // e.g. "create_challenge", "delete_user"
	Resource  string             `bson:"resource" json:"resource"` // e.g. "challenge", "notification"
	Details   string             `bson:"details" json:"details"`   // Additional info
	IPAddress string             `bson:"ip_address" json:"ip_address"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
