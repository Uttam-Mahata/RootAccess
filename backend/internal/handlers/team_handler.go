package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
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
// @Summary Create a team
// @Description Create a new team. The authenticated user becomes the team leader.
// @Tags Teams
// @Accept json
// @Produce json
// @Param request body CreateTeamRequest true "Team details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /teams [post]
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
// @Summary Get my team
// @Description Retrieve details of the team the authenticated user belongs to.
// @Tags Teams
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /teams/my-team [get]
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
// @Summary Get team details
// @Description Retrieve details of a specific team by its ID.
// @Tags Teams
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /teams/{id} [get]
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
// @Summary Update team
// @Description Update the name and description of a team. Only the team leader can perform this action.
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param request body UpdateTeamRequest true "Updated team details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /teams/{id} [put]
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
// @Summary Delete team
// @Description Permanently delete a team. Only the team leader can perform this action and only if the team has fewer than 2 members.
// @Tags Teams
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /teams/{id} [delete]
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
// @Summary Invite by username
// @Description Send a team invitation to a user by their username.
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param request body InviteByUsernameRequest true "User to invite"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /teams/{id}/invite/username [post]
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
// @Summary Invite by email
// @Description Send a team invitation to a user by their email address.
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param request body InviteByEmailRequest true "Email to invite"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /teams/{id}/invite/email [post]
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
// @Summary Join team by code
// @Description Join an existing team using its unique invite code.
// @Tags Teams
// @Accept json
// @Produce json
// @Param code path string true "Invite code"
// @Param request body JoinByCodeRequest false "Invite code in body"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /teams/join/{code} [post]
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
// @Summary Get pending invitations
// @Description Retrieve all team invitations sent to the authenticated user.
// @Tags Teams
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security ApiKeyAuth
// @Router /teams/invitations [get]
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
// @Summary Accept invitation
// @Description Accept a team invitation and join the team.
// @Tags Teams
// @Produce json
// @Param id path string true "Invitation ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /teams/invitations/{id}/accept [post]
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
// @Summary Reject invitation
// @Description Decline a team invitation.
// @Tags Teams
// @Produce json
// @Param id path string true "Invitation ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /teams/invitations/{id}/reject [post]
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
// @Summary Remove team member
// @Description Remove a specific user from the team. Only the team leader can perform this action.
// @Tags Teams
// @Produce json
// @Param id path string true "Team ID"
// @Param userId path string true "User ID to remove"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /teams/{id}/members/{userId} [delete]
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
// @Summary Leave team
// @Description Exit the current team. If the leader leaves and is the only member, the team is deleted.
// @Tags Teams
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /teams/{id}/leave [post]
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
// @Summary Regenerate invite code
// @Description Create a new unique invite code for the team, invalidating the old one.
// @Tags Teams
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /teams/{id}/regenerate-code [post]
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
// @Summary Get team's outgoing invitations
// @Description Retrieve all pending invitations sent by this team. Only the team leader can see this.
// @Tags Teams
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {array} models.TeamInvitation
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /teams/{id}/invitations [get]
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
// @Summary Cancel invitation
// @Description Withdraw a pending team invitation.
// @Tags Teams
// @Produce json
// @Param id path string true "Team ID"
// @Param invitationId path string true "Invitation ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /teams/{id}/invitations/{invitationId} [delete]
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
// @Summary Get team scoreboard
// @Description Retrieve a list of all teams sorted by their total points.
// @Tags Teams
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /teams/scoreboard [get]
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
