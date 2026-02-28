package repositories

import (
	"database/sql"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/google/uuid"
)

type AuditLogRepository struct {
	db *sql.DB
}

func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) CreateLog(log *models.AuditLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	log.CreatedAt = time.Now()

	_, err := r.db.Exec("INSERT INTO audit_logs (id, user_id, username, action, resource, details, ip_address, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		log.ID, log.UserID, log.Username, log.Action, log.Resource, log.Details, log.IPAddress, log.CreatedAt.Format(time.RFC3339))
	return err
}

func (r *AuditLogRepository) GetLogs(limit int, page int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	skip := (page - 1) * limit

	rows, err := r.db.Query("SELECT id, user_id, username, action, resource, details, ip_address, created_at FROM audit_logs ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		var created string
		if err := rows.Scan(&l.ID, &l.UserID, &l.Username, &l.Action, &l.Resource, &l.Details, &l.IPAddress, &created); err != nil {
			return nil, err
		}
		l.CreatedAt, _ = time.Parse(time.RFC3339, created)
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *AuditLogRepository) GetLogCount() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM audit_logs").Scan(&count)
	return count, err
}
