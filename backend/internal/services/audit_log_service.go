package services

import (
	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
)

type AuditLogService struct {
	auditRepo *repositories.AuditLogRepository
}

func NewAuditLogService(auditRepo *repositories.AuditLogRepository) *AuditLogService {
	return &AuditLogService{
		auditRepo: auditRepo,
	}
}

// Log records an audit event
func (s *AuditLogService) Log(userID string, username, action, resource, details, ipAddress string) {
	log := &models.AuditLog{
		UserID:    userID,
		Username:  username,
		Action:    action,
		Resource:  resource,
		Details:   details,
		IPAddress: ipAddress,
	}
	// Fire and forget - audit logging should not block operations
	go func() {
		_ = s.auditRepo.CreateLog(log)
	}()
}

// GetLogs returns paginated audit logs
func (s *AuditLogService) GetLogs(limit, page int) ([]models.AuditLog, int64, error) {
	logs, err := s.auditRepo.GetLogs(limit, page)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.auditRepo.GetLogCount()
	if err != nil {
		return nil, 0, err
	}

	return logs, count, nil
}
