package services

import (
	"errors"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationService struct {
	notificationRepo *repositories.NotificationRepository
}

func NewNotificationService(notificationRepo *repositories.NotificationRepository) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
	}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(title, content, notifType string, createdBy primitive.ObjectID) (*models.Notification, error) {
	// Validate notification type
	if !models.IsValidNotificationType(notifType) {
		return nil, errors.New("invalid notification type")
	}

	notification := &models.Notification{
		Title:     title,
		Content:   content,
		Type:      notifType,
		CreatedBy: createdBy,
		IsActive:  true,
	}

	err := s.notificationRepo.CreateNotification(notification)
	if err != nil {
		return nil, err
	}

	return notification, nil
}

// GetAllNotifications returns all notifications (for admin)
func (s *NotificationService) GetAllNotifications() ([]models.Notification, error) {
	return s.notificationRepo.GetAllNotifications()
}

// GetActiveNotifications returns only active notifications (for users)
func (s *NotificationService) GetActiveNotifications() ([]models.Notification, error) {
	return s.notificationRepo.GetActiveNotifications()
}

// GetNotificationByID returns a notification by ID
func (s *NotificationService) GetNotificationByID(id string) (*models.Notification, error) {
	return s.notificationRepo.GetNotificationByID(id)
}

// UpdateNotification updates a notification
func (s *NotificationService) UpdateNotification(id string, title, content, notifType string, isActive bool) error {
	// Validate notification type if provided
	if notifType != "" && !models.IsValidNotificationType(notifType) {
		return errors.New("invalid notification type")
	}

	notification := &models.Notification{
		Title:    title,
		Content:  content,
		Type:     notifType,
		IsActive: isActive,
	}

	return s.notificationRepo.UpdateNotification(id, notification)
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(id string) error {
	return s.notificationRepo.DeleteNotification(id)
}

// ToggleNotificationActive toggles the is_active status
func (s *NotificationService) ToggleNotificationActive(id string) error {
	return s.notificationRepo.ToggleNotificationActive(id)
}
