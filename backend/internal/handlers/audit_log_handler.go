package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
)

type AuditLogHandler struct {
	auditService *services.AuditLogService
}

func NewAuditLogHandler(auditService *services.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{
		auditService: auditService,
	}
}

// GetAuditLogs returns paginated audit logs (admin only)
// @Summary Get audit logs
// @Description Retrieve paginated logs of administrative actions.
// @Tags Audit
// @Produce json
// @Param limit query int false "Number of logs per page" default(50)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/audit-logs [get]
func (h *AuditLogHandler) GetAuditLogs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	pageStr := c.DefaultQuery("page", "1")

	limit, _ := strconv.Atoi(limitStr)
	page, _ := strconv.Atoi(pageStr)

	logs, total, err := h.auditService.GetLogs(limit, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
