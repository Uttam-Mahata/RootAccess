package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/cache"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories/interfaces"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminTeamHandler struct {
	teamRepo       interfaces.TeamRepository
	userRepo       interfaces.UserRepository
	submissionRepo interfaces.SubmissionRepository
	invitationRepo interfaces.TeamInvitationRepository
	adjustmentRepo interfaces.ScoreAdjustmentRepository
	cache          cache.CacheProvider
}

func NewAdminTeamHandler(
	teamRepo interfaces.TeamRepository,
	userRepo interfaces.UserRepository,
	submissionRepo interfaces.SubmissionRepository,
	invitationRepo interfaces.TeamInvitationRepository,
	adjustmentRepo interfaces.ScoreAdjustmentRepository,
	cp cache.CacheProvider,
) *AdminTeamHandler {
	return &AdminTeamHandler{
		teamRepo:       teamRepo,
		userRepo:       userRepo,
		submissionRepo: submissionRepo,
		invitationRepo: invitationRepo,
		adjustmentRepo: adjustmentRepo,
		cache:          cp,
	}
}

// AdminTeamResponse represents a team with additional admin info
type AdminTeamResponse struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Avatar      string              `json:"avatar,omitempty"`
	LeaderID    string              `json:"leader_id"`
	LeaderName  string              `json:"leader_name"`
	MemberCount int                 `json:"member_count"`
	Members     []TeamMemberInfo    `json:"members"`
	InviteCode  string              `json:"invite_code"`
	Score       int                 `json:"score"`
	// Future: we could expose effective_score including adjustments if needed
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
}

// AdjustTeamScoreRequest represents a manual team score adjustment
type AdjustTeamScoreRequest struct {
	Delta  int    `json:"delta" binding:"required"`
	Reason string `json:"reason"`
}

// AdjustTeamScore allows admins to add or deduct points from a team's score.
// @Summary Adjust team score
// @Description Apply a manual score delta (positive or negative) to a team.
// @Tags Admin Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param request body AdjustTeamScoreRequest true "Score adjustment"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/teams/{id}/score-adjust [post]
func (h *AdminTeamHandler) AdjustTeamScore(c *gin.Context) {
	teamID := c.Param("id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Team ID is required"})
		return
	}

	var req AdjustTeamScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	if req.Delta == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Delta must be non-zero"})
		return
	}

	// Ensure team exists
	team, err := h.teamRepo.FindTeamByID(teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Update the team's stored score so admin views and analytics see it.
	if err := h.teamRepo.UpdateTeamScore(teamID, req.Delta); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team score"})
		return
	}

	// Record a score adjustment so scoreboard/analytics can apply it
	if h.adjustmentRepo != nil {
		adminIDStr, _ := c.Get("user_id")
		adminIDHex, _ := adminIDStr.(string)
		adminID, _ := primitive.ObjectIDFromHex(adminIDHex)

		adj := &models.ScoreAdjustment{
			TargetType: models.ScoreAdjustmentTargetTeam,
			TargetID:   team.ID,
			Delta:      req.Delta,
			Reason:     req.Reason,
			CreatedBy:  adminID,
		}
		_ = h.adjustmentRepo.Create(adj)
	}

	// Clear relevant scoreboard caches to reflect changes quickly
	if h.cache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = h.cache.Del(ctx,
			"team_scoreboard",
			"team_scoreboard_frozen",
		)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team score adjusted successfully"})
}

// TeamMemberInfo represents a team member's info
type TeamMemberInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsLeader bool   `json:"is_leader"`
}

// ListTeams returns all teams with detailed information for admin
// @Summary List all teams
// @Description Retrieve a list of all teams with detailed administrative information.
// @Tags Admin Teams
// @Produce json
// @Success 200 {array} AdminTeamResponse
// @Security ApiKeyAuth
// @Router /admin/teams [get]
func (h *AdminTeamHandler) ListTeams(c *gin.Context) {
	teams, err := h.teamRepo.GetAllTeams()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	result := make([]AdminTeamResponse, 0, len(teams))
	for _, team := range teams {
		// Get leader info
		leaderName := ""
		leader, err := h.userRepo.FindByID(team.LeaderID.Hex())
		if err == nil {
			leaderName = leader.Username
		}

		// Get member info
		members := make([]TeamMemberInfo, 0, len(team.MemberIDs))
		for _, memberID := range team.MemberIDs {
			user, err := h.userRepo.FindByID(memberID.Hex())
			if err == nil {
				members = append(members, TeamMemberInfo{
					ID:       user.ID.Hex(),
					Username: user.Username,
					Email:    user.Email,
					IsLeader: memberID == team.LeaderID,
				})
			}
		}

		result = append(result, AdminTeamResponse{
			ID:          team.ID.Hex(),
			Name:        team.Name,
			Description: team.Description,
			Avatar:      team.Avatar,
			LeaderID:    team.LeaderID.Hex(),
			LeaderName:  leaderName,
			MemberCount: len(team.MemberIDs),
			Members:     members,
			InviteCode:  team.InviteCode,
			Score:       team.Score,
			CreatedAt:   team.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   team.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, result)
}

// GetTeam returns detailed information about a specific team
// @Summary Get team (admin)
// @Description Retrieve detailed information about a specific team by its ID.
// @Tags Admin Teams
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {object} AdminTeamResponse
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/teams/{id} [get]
func (h *AdminTeamHandler) GetTeam(c *gin.Context) {
	teamID := c.Param("id")
	
	team, err := h.teamRepo.FindTeamByID(teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Get leader info
	leaderName := ""
	leader, err := h.userRepo.FindByID(team.LeaderID.Hex())
	if err == nil {
		leaderName = leader.Username
	}

	// Get member info
	members := make([]TeamMemberInfo, 0, len(team.MemberIDs))
	for _, memberID := range team.MemberIDs {
		user, err := h.userRepo.FindByID(memberID.Hex())
		if err == nil {
			members = append(members, TeamMemberInfo{
				ID:       user.ID.Hex(),
				Username: user.Username,
				Email:    user.Email,
				IsLeader: memberID == team.LeaderID,
			})
		}
	}

	result := AdminTeamResponse{
		ID:          team.ID.Hex(),
		Name:        team.Name,
		Description: team.Description,
		Avatar:      team.Avatar,
		LeaderID:    team.LeaderID.Hex(),
		LeaderName:  leaderName,
		MemberCount: len(team.MemberIDs),
		Members:     members,
		InviteCode:  team.InviteCode,
		Score:       team.Score,
		CreatedAt:   team.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   team.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, result)
}

// AdminUpdateTeamRequest represents admin team update request
type AdminUpdateTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateTeam updates team details (admin only)
// @Summary Update team (admin)
// @Description Update a team's name or description.
// @Tags Admin Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param request body AdminUpdateTeamRequest true "Team details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/teams/{id} [put]
func (h *AdminTeamHandler) UpdateTeam(c *gin.Context) {
	teamID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var req AdminUpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Check if team exists
	_, err = h.teamRepo.FindTeamByID(teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	update := map[string]interface{}{}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Description != "" {
		update["description"] = req.Description
	}

	if len(update) > 0 {
		if err := h.teamRepo.UpdateTeamFields(objID, update); err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team updated successfully"})
}

// UpdateTeamLeaderRequest represents request to change team leader
type UpdateTeamLeaderRequest struct {
	NewLeaderID string `json:"new_leader_id" binding:"required"`
}

// UpdateTeamLeader changes the team leader (admin only)
// @Summary Change team leader
// @Description Transfer team leadership to another member of the team.
// @Tags Admin Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param request body UpdateTeamLeaderRequest true "New leader details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/teams/{id}/leader [put]
func (h *AdminTeamHandler) UpdateTeamLeader(c *gin.Context) {
	teamID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var req UpdateTeamLeaderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	newLeaderID, err := primitive.ObjectIDFromHex(req.NewLeaderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid new leader ID"})
		return
	}

	// Check if team exists
	team, err := h.teamRepo.FindTeamByID(teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Verify new leader is a member of the team
	isMember := false
	for _, memberID := range team.MemberIDs {
		if memberID == newLeaderID {
			isMember = true
			break
		}
	}

	if !isMember {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New leader must be a member of the team"})
		return
	}

	if err := h.teamRepo.AdminUpdateTeamLeader(objID, newLeaderID); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team leader updated successfully"})
}

// RemoveMember removes a member from a team (admin only)
// @Summary Remove team member (admin)
// @Description Remove a member from a team. The leader cannot be removed.
// @Tags Admin Teams
// @Produce json
// @Param id path string true "Team ID"
// @Param memberId path string true "Member ID to remove"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/teams/{id}/members/{memberId} [delete]
func (h *AdminTeamHandler) RemoveMember(c *gin.Context) {
	teamID := c.Param("id")
	memberID := c.Param("memberId")

	// Check if team exists
	team, err := h.teamRepo.FindTeamByID(teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	memberObjID, err := primitive.ObjectIDFromHex(memberID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID"})
		return
	}

	// Check if member is the leader
	if team.LeaderID == memberObjID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove the team leader. Change leader first or delete the team."})
		return
	}

	// Check if member is in the team
	isMember := false
	for _, id := range team.MemberIDs {
		if id == memberObjID {
			isMember = true
			break
		}
	}

	if !isMember {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not a member of this team"})
		return
	}

	if err := h.teamRepo.RemoveMemberFromTeam(teamID, memberID); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed from team"})
}

// DeleteTeam deletes a team (admin only, bypasses normal restrictions)
// @Summary Delete team (admin)
// @Description Permanently delete a team and all its invitations.
// @Tags Admin Teams
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/teams/{id} [delete]
func (h *AdminTeamHandler) DeleteTeam(c *gin.Context) {
	teamID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	// Check if team exists
	_, err = h.teamRepo.FindTeamByID(teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Delete all invitations for this team
	if h.invitationRepo != nil {
		_ = h.invitationRepo.DeleteInvitationsByTeam(teamID)
	}

	if err := h.teamRepo.AdminDeleteTeam(objID); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team deleted successfully"})
}
