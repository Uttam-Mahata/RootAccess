package turso

import (
	"database/sql"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
)

const auditLogColumns = `id, user_id, username, action, resource, details, ip_address, created_at`

// AuditLogRepository implements interfaces.AuditLogRepository using database/sql.
type AuditLogRepository struct {
	db *sql.DB
}

// NewAuditLogRepository creates a new Turso-backed AuditLogRepository.
func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func scanAuditLog(s scanner) (*models.AuditLog, error) {
	var a models.AuditLog
	var id, userID, createdAt string

	err := s.Scan(&id, &userID, &a.Username, &a.Action, &a.Resource, &a.Details, &a.IPAddress, &createdAt)
	if err != nil {
		return nil, err
	}

	a.ID = oidFromHex(id)
	a.UserID = oidFromHex(userID)
	a.CreatedAt = strToTime(createdAt)
	return &a, nil
}

func (r *AuditLogRepository) CreateLog(log *models.AuditLog) error {
	id := newID()
	log.CreatedAt = time.Now()

	_, err := r.db.Exec(
		`INSERT INTO audit_logs (`+auditLogColumns+`) VALUES (?,?,?,?,?,?,?,?)`,
		id, log.UserID.Hex(), log.Username, log.Action, log.Resource,
		log.Details, log.IPAddress, timeToStr(log.CreatedAt),
	)
	if err != nil {
		return err
	}
	log.ID = oidFromHex(id)
	return nil
}

func (r *AuditLogRepository) GetLogs(limit int, page int) ([]models.AuditLog, error) {
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.Query(
		`SELECT `+auditLogColumns+` FROM audit_logs ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.AuditLog
	for rows.Next() {
		a, err := scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *a)
	}
	return results, rows.Err()
}

func (r *AuditLogRepository) GetLogCount() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM audit_logs`).Scan(&count)
	return count, err
}
