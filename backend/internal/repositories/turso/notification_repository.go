package turso

import (
	"database/sql"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
)

const notificationColumns = `id, title, content, type, created_by, created_at, is_active`

// NotificationRepository implements interfaces.NotificationRepository using database/sql.
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new Turso-backed NotificationRepository.
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func scanNotification(s scanner) (*models.Notification, error) {
	var n models.Notification
	var (
		id, createdBy string
		isActive      int
		createdAt     string
	)

	err := s.Scan(&id, &n.Title, &n.Content, &n.Type, &createdBy, &createdAt, &isActive)
	if err != nil {
		return nil, err
	}

	n.ID = oidFromHex(id)
	n.CreatedBy = oidFromHex(createdBy)
	n.CreatedAt = strToTime(createdAt)
	n.IsActive = intToBool(isActive)
	return &n, nil
}

func (r *NotificationRepository) queryNotifications(query string, args ...interface{}) ([]models.Notification, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *n)
	}
	return results, rows.Err()
}

func (r *NotificationRepository) CreateNotification(notification *models.Notification) error {
	id := newID()
	notification.CreatedAt = time.Now()

	_, err := r.db.Exec(
		`INSERT INTO notifications (`+notificationColumns+`) VALUES (?,?,?,?,?,?,?)`,
		id, notification.Title, notification.Content, notification.Type,
		notification.CreatedBy.Hex(), timeToStr(notification.CreatedAt),
		boolToInt(notification.IsActive),
	)
	if err != nil {
		return err
	}
	notification.ID = oidFromHex(id)
	return nil
}

func (r *NotificationRepository) GetAllNotifications() ([]models.Notification, error) {
	return r.queryNotifications(`SELECT ` + notificationColumns + ` FROM notifications ORDER BY created_at DESC`)
}

func (r *NotificationRepository) GetActiveNotifications() ([]models.Notification, error) {
	return r.queryNotifications(`SELECT `+notificationColumns+` FROM notifications WHERE is_active = 1 ORDER BY created_at DESC`)
}

func (r *NotificationRepository) GetNotificationByID(id string) (*models.Notification, error) {
	row := r.db.QueryRow(`SELECT `+notificationColumns+` FROM notifications WHERE id = ?`, id)
	return scanNotification(row)
}

func (r *NotificationRepository) UpdateNotification(id string, notification *models.Notification) error {
	_, err := r.db.Exec(
		`UPDATE notifications SET title=?, content=?, type=?, is_active=? WHERE id=?`,
		notification.Title, notification.Content, notification.Type,
		boolToInt(notification.IsActive), id,
	)
	return err
}

func (r *NotificationRepository) DeleteNotification(id string) error {
	_, err := r.db.Exec(`DELETE FROM notifications WHERE id = ?`, id)
	return err
}

func (r *NotificationRepository) ToggleNotificationActive(id string) error {
	_, err := r.db.Exec(`UPDATE notifications SET is_active = CASE WHEN is_active = 0 THEN 1 ELSE 0 END WHERE id = ?`, id)
	return err
}
