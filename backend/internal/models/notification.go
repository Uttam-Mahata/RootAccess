package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification represents an admin broadcast message
type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Content   string             `bson:"content" json:"content"`
	Type      string             `bson:"type" json:"type"` // "info", "warning", "success", "error"
	CreatedBy primitive.ObjectID `bson:"created_by" json:"created_by"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
}

// NotificationTypes defines valid notification types
var NotificationTypes = []string{"info", "warning", "success", "error"}

// IsValidType checks if the notification type is valid
func IsValidNotificationType(t string) bool {
	for _, validType := range NotificationTypes {
		if t == validType {
			return true
		}
	}
	return false
}
