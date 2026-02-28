package models

import "time"

type Notification struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active"`
}

var NotificationTypes = []string{"info", "warning", "success", "error"}

func IsValidNotificationType(t string) bool {
	for _, validType := range NotificationTypes {
		if t == validType {
			return true
		}
	}
	return false
}
