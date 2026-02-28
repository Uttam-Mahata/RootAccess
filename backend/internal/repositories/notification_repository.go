package repositories

import (
	"database/sql"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/google/uuid"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) CreateNotification(n *models.Notification) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	n.CreatedAt = time.Now()
	n.IsActive = true

	query := `INSERT INTO notifications (id, title, content, type, created_by, created_at, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, n.ID, n.Title, n.Content, n.Type, n.CreatedBy, n.CreatedAt.Format(time.RFC3339), 1)
	return err
}

func (r *NotificationRepository) scanNotifications(rows *sql.Rows) ([]models.Notification, error) {
	var ns []models.Notification
	for rows.Next() {
		var n models.Notification
		var ts string
		var isActive int
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.Type, &n.CreatedBy, &ts, &isActive); err != nil {
			return nil, err
		}
		n.CreatedAt, _ = time.Parse(time.RFC3339, ts)
		n.IsActive = isActive == 1
		ns = append(ns, n)
	}
	return ns, nil
}

func (r *NotificationRepository) GetAllNotifications() ([]models.Notification, error) {
	rows, err := r.db.Query("SELECT id, title, content, type, created_by, created_at, is_active FROM notifications ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanNotifications(rows)
}

func (r *NotificationRepository) GetActiveNotifications() ([]models.Notification, error) {
	rows, err := r.db.Query("SELECT id, title, content, type, created_by, created_at, is_active FROM notifications WHERE is_active=1 ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanNotifications(rows)
}

func (r *NotificationRepository) GetNotificationByID(id string) (*models.Notification, error) {
	rows, err := r.db.Query("SELECT id, title, content, type, created_by, created_at, is_active FROM notifications WHERE id=?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ns, err := r.scanNotifications(rows)
	if err != nil || len(ns) == 0 {
		return nil, sql.ErrNoRows
	}
	return &ns[0], nil
}

func (r *NotificationRepository) UpdateNotification(id string, n *models.Notification) error {
	isActive := 0
	if n.IsActive {
		isActive = 1
	}
	_, err := r.db.Exec("UPDATE notifications SET title=?, content=?, type=?, is_active=? WHERE id=?", n.Title, n.Content, n.Type, isActive, id)
	return err
}

func (r *NotificationRepository) DeleteNotification(id string) error {
	_, err := r.db.Exec("DELETE FROM notifications WHERE id=?", id)
	return err
}

func (r *NotificationRepository) ToggleNotificationActive(id string) error {
	_, err := r.db.Exec("UPDATE notifications SET is_active = CASE WHEN is_active=1 THEN 0 ELSE 1 END WHERE id=?", id)
	return err
}
