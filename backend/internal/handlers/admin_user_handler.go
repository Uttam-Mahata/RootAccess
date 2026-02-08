package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminUserHandler struct {
	userRepo *repositories.UserRepository
}

func NewAdminUserHandler(userRepo *repositories.UserRepository) *AdminUserHandler {
	return &AdminUserHandler{userRepo: userRepo}
}

func (h *AdminUserHandler) ListUsers(c *gin.Context) {
	users, err := h.userRepo.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Strip sensitive fields
	var result []gin.H
	for _, u := range users {
		result = append(result, gin.H{
			"id":             u.ID.Hex(),
			"username":       u.Username,
			"email":          u.Email,
			"role":           u.Role,
			"status":         u.Status,
			"email_verified": u.EmailVerified,
			"created_at":     u.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, result)
}

type UpdateUserStatusRequest struct {
	Status    string `json:"status" binding:"required"`
	BanReason string `json:"ban_reason"`
}

func (h *AdminUserHandler) UpdateUserStatus(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Status != "active" && req.Status != "banned" && req.Status != "suspended" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be 'active', 'banned', or 'suspended'"})
		return
	}
	update := bson.M{"status": req.Status, "ban_reason": req.BanReason}
	err = h.userRepo.UpdateFields(objID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User status updated"})
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

func (h *AdminUserHandler) UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role != "admin" && req.Role != "user" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be 'admin' or 'user'"})
		return
	}
	err = h.userRepo.UpdateFields(objID, bson.M{"role": req.Role})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User role updated"})
}
