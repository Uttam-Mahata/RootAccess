package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/services"
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
