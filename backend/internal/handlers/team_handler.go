package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/services"
)

type TeamHandler struct {
	teamService *services.TeamService
}

func NewTeamHandler(teamService *services.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

// Request structs
type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=50"`
	Description string `json:"description" binding:"max=500"`
}

type UpdateTeamRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=50"`
	Description string `json:"description" binding:"max=500"`
}

type InviteByUsernameRequest struct {
	Username string `json:"username" binding:"required"`
}

type InviteByEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type JoinByCodeRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

// CreateTeam creates a new team with the current user as leader
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	team, err := h.teamService.CreateTeam(userID.(string), req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get team members immediately after creation
	members, err := h.teamService.GetTeamMembers(team.ID.Hex())
	if err != nil {
		// If we can't get members, still return success with just the team
		c.JSON(http.StatusCreated, gin.H{
			"message": "Team created successfully!",
			"team":    team,
			"members": []interface{}{},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Team created successfully!",
		"team":    team,
		"members": members,
	})
}

// GetMyTeam returns the current user's team
func (h *TeamHandler) GetMyTeam(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	team, err := h.teamService.GetUserTeam(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "you are not a member of any team"})
		return
	}

	// Get team members with details
	members, err := h.teamService.GetTeamMembers(team.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get team members"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"team":    team,
		"members": members,
	})
}

// GetTeamDetails returns details about a specific team
func (h *TeamHandler) GetTeamDetails(c *gin.Context) {
	teamID := c.Param("id")

	team, err := h.teamService.GetTeamByID(teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	// Get team members with details
	members, _ := h.teamService.GetTeamMembers(teamID)

	c.JSON(http.StatusOK, gin.H{
		"team":    team,
		"members": members,
	})
}

// UpdateTeam updates the team name and description
func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	teamID := c.Param("id")

	var req UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	team, err := h.teamService.UpdateTeam(teamID, userID.(string), req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get team members after update
	members, err := h.teamService.GetTeamMembers(team.ID.Hex())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Team updated successfully!",
			"team":    team,
			"members": []interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team updated successfully!",
		"team":    team,
		"members": members,
	})
}

// DeleteTeam deletes the team
func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	teamID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.teamService.DeleteTeam(teamID, userID.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team deleted successfully!",
	})
}

// InviteByUsername invites a user to the team by their username
func (h *TeamHandler) InviteByUsername(c *gin.Context) {
	teamID := c.Param("id")

	var req InviteByUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	invitation, err := h.teamService.InviteByUsername(teamID, userID.(string), req.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Invitation sent successfully!",
		"invitation": invitation,
	})
}

// InviteByEmail invites a user to the team by their email
func (h *TeamHandler) InviteByEmail(c *gin.Context) {
	teamID := c.Param("id")

	var req InviteByEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	invitation, err := h.teamService.InviteByEmail(teamID, userID.(string), req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Invitation sent successfully!",
		"invitation": invitation,
	})
}

// JoinByCode allows a user to join a team using the invite code
func (h *TeamHandler) JoinByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		var req JoinByCodeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		code = req.InviteCode
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	team, err := h.teamService.JoinByInviteCode(userID.(string), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get team members immediately after joining
	members, err := h.teamService.GetTeamMembers(team.ID.Hex())
	if err != nil {
		// If we can't get members, still return success with just the team
		c.JSON(http.StatusOK, gin.H{
			"message": "Successfully joined the team!",
			"team":    team,
			"members": []interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully joined the team!",
		"team":    team,
		"members": members,
	})
}

// GetPendingInvitations returns all pending invitations for the current user
func (h *TeamHandler) GetPendingInvitations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	email, _ := c.Get("email")
	emailStr, _ := email.(string)

	invitations, err := h.teamService.GetPendingInvitations(userID.(string), emailStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get invitations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
	})
}

// AcceptInvitation accepts a team invitation
func (h *TeamHandler) AcceptInvitation(c *gin.Context) {
	invitationID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	team, err := h.teamService.AcceptInvitation(invitationID, userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get team members immediately after accepting invitation
	members, err := h.teamService.GetTeamMembers(team.ID.Hex())
	if err != nil {
		// If we can't get members, still return success with just the team
		c.JSON(http.StatusOK, gin.H{
			"message": "Invitation accepted! You are now a member of the team.",
			"team":    team,
			"members": []interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation accepted! You are now a member of the team.",
		"team":    team,
		"members": members,
	})
}

// RejectInvitation rejects a team invitation
func (h *TeamHandler) RejectInvitation(c *gin.Context) {
	invitationID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.teamService.RejectInvitation(invitationID, userID.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation rejected.",
	})
}

// RemoveMember removes a member from the team
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamID := c.Param("id")
	memberID := c.Param("userId")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.teamService.RemoveMember(teamID, userID.(string), memberID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member removed from team.",
	})
}

// LeaveTeam allows a member to leave the team
func (h *TeamHandler) LeaveTeam(c *gin.Context) {
	teamID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.teamService.LeaveTeam(teamID, userID.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "You have left the team.",
	})
}

// RegenerateInviteCode generates a new invite code for the team
func (h *TeamHandler) RegenerateInviteCode(c *gin.Context) {
	teamID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	newCode, err := h.teamService.RegenerateInviteCode(teamID, userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Invite code regenerated!",
		"invite_code": newCode,
	})
}

// GetTeamPendingInvitations returns pending invitations sent by the team
func (h *TeamHandler) GetTeamPendingInvitations(c *gin.Context) {
	teamID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	invitations, err := h.teamService.GetTeamPendingInvitations(teamID, userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
	})
}

// CancelInvitation cancels a pending invitation
func (h *TeamHandler) CancelInvitation(c *gin.Context) {
	invitationID := c.Param("invitationId")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.teamService.CancelInvitation(invitationID, userID.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation cancelled.",
	})
}

// GetTeamScoreboard returns all teams sorted by score
func (h *TeamHandler) GetTeamScoreboard(c *gin.Context) {
	teams, err := h.teamService.GetAllTeamsScoreboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get team scoreboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"teams": teams,
	})
}
