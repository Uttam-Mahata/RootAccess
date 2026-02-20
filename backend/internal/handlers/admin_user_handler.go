package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminUserHandler struct {
	userRepo       *repositories.UserRepository
	teamRepo       *repositories.TeamRepository
	submissionRepo *repositories.SubmissionRepository
}

func NewAdminUserHandler(userRepo *repositories.UserRepository) *AdminUserHandler {
	return &AdminUserHandler{userRepo: userRepo}
}

func NewAdminUserHandlerWithRepos(userRepo *repositories.UserRepository, teamRepo *repositories.TeamRepository, submissionRepo *repositories.SubmissionRepository) *AdminUserHandler {
	return &AdminUserHandler{
		userRepo:       userRepo,
		teamRepo:       teamRepo,
		submissionRepo: submissionRepo,
	}
}

// AdminUserResponse represents a user with detailed admin info
type AdminUserResponse struct {
	ID             string      `json:"id"`
	Username       string      `json:"username"`
	Email          string      `json:"email"`
	Role           string      `json:"role"`
	Status         string      `json:"status"`
	EmailVerified  bool        `json:"email_verified"`
	LastIP         string      `json:"last_ip,omitempty"`
	IPHistory      interface{} `json:"ip_history,omitempty"`
	LastLoginAt    string      `json:"last_login_at,omitempty"`
	TeamID         string      `json:"team_id,omitempty"`
	TeamName       string      `json:"team_name,omitempty"`
	BanReason      string      `json:"ban_reason,omitempty"`
	OAuthProvider  string      `json:"oauth_provider,omitempty"`
	CreatedAt      string      `json:"created_at"`
	UpdatedAt      string      `json:"updated_at"`
}

// ListUsers returns all users with detailed information for admin
// @Summary List all users
// @Description Retrieve a list of all users with detailed administrative information.
// @Tags Admin Users
// @Produce json
// @Success 200 {array} AdminUserResponse
// @Security ApiKeyAuth
// @Router /admin/users [get]
func (h *AdminUserHandler) ListUsers(c *gin.Context) {
	users, err := h.userRepo.GetUsersWithDetails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]AdminUserResponse, 0, len(users))
	for _, u := range users {
		response := AdminUserResponse{
			ID:            u.ID.Hex(),
			Username:      u.Username,
			Email:         u.Email,
			Role:          u.Role,
			Status:        u.Status,
			EmailVerified: u.EmailVerified,
			LastIP:        u.LastIP,
			IPHistory:     u.IPHistory,
			BanReason:     u.BanReason,
			CreatedAt:     u.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}

		if u.LastLoginAt != nil {
			response.LastLoginAt = u.LastLoginAt.Format("2006-01-02T15:04:05Z")
		}

		if u.OAuth != nil {
			response.OAuthProvider = u.OAuth.Provider
		}

		// Get team info if available
		if h.teamRepo != nil {
			team, err := h.teamRepo.FindTeamByMemberID(u.ID.Hex())
			if err == nil && team != nil {
				response.TeamID = team.ID.Hex()
				response.TeamName = team.Name
			}
		}

		result = append(result, response)
	}
	c.JSON(http.StatusOK, result)
}

// GetUser returns detailed information about a specific user
// @Summary Get user (admin)
// @Description Retrieve detailed information about a specific user by their ID.
// @Tags Admin Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} AdminUserResponse
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/users/{id} [get]
func (h *AdminUserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	response := AdminUserResponse{
		ID:            user.ID.Hex(),
		Username:      user.Username,
		Email:         user.Email,
		Role:          user.Role,
		Status:        user.Status,
		EmailVerified: user.EmailVerified,
		LastIP:        user.LastIP,
		IPHistory:     user.IPHistory,
		BanReason:     user.BanReason,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if user.LastLoginAt != nil {
		response.LastLoginAt = user.LastLoginAt.Format("2006-01-02T15:04:05Z")
	}

	if user.OAuth != nil {
		response.OAuthProvider = user.OAuth.Provider
	}

	// Get team info if available
	if h.teamRepo != nil {
		team, err := h.teamRepo.FindTeamByMemberID(user.ID.Hex())
		if err == nil && team != nil {
			response.TeamID = team.ID.Hex()
			response.TeamName = team.Name
		}
	}

	c.JSON(http.StatusOK, response)
}

type UpdateUserStatusRequest struct {
	Status    string `json:"status" binding:"required"`
	BanReason string `json:"ban_reason"`
}

// UpdateUserStatus updates user's account status
// @Summary Update user status
// @Description Change a user's account status (active, banned, or suspended).
// @Tags Admin Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body UpdateUserStatusRequest true "New status details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/users/{id}/status [put]
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

// UpdateUserRole updates user's role
// @Summary Update user role
// @Description Change a user's role (admin or user). Admins cannot change their own role.
// @Tags Admin Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body UpdateUserRoleRequest true "New role"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/users/{id}/role [put]
func (h *AdminUserHandler) UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Prevent admins from modifying their own role
	currentUserID, _ := c.Get("user_id")
	if currentUserID.(string) == id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot modify your own role"})
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

// DeleteUser deletes a user (admin only)
// @Summary Delete user (admin)
// @Description Perform a soft delete on a user's account by setting status to 'deleted'.
// @Tags Admin Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/users/{id} [delete]
func (h *AdminUserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Prevent admins from deleting themselves
	currentUserID, _ := c.Get("user_id")
	if currentUserID.(string) == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	// Check if user exists
	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Prevent deleting other admins (optional protection)
	if user.Role == "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete admin users. Demote to user first."})
		return
	}

	// Delete user from database
	// Note: This is a soft delete - the user's status is set to "deleted" rather than
	// physically removing the record. This preserves data integrity for historical records
	// like submissions and team memberships. The user will still appear in total counts.
	// For a full purge, implement data cleanup in related tables (submissions, teams, etc.)
	err = h.userRepo.UpdateFields(objID, bson.M{"status": "deleted"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
