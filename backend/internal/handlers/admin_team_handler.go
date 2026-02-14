package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminTeamHandler struct {
	teamRepo       *repositories.TeamRepository
	userRepo       *repositories.UserRepository
	submissionRepo *repositories.SubmissionRepository
	invitationRepo *repositories.TeamInvitationRepository
}

func NewAdminTeamHandler(
	teamRepo *repositories.TeamRepository,
	userRepo *repositories.UserRepository,
	submissionRepo *repositories.SubmissionRepository,
	invitationRepo *repositories.TeamInvitationRepository,
) *AdminTeamHandler {
	return &AdminTeamHandler{
		teamRepo:       teamRepo,
		userRepo:       userRepo,
		submissionRepo: submissionRepo,
		invitationRepo: invitationRepo,
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
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
}

// TeamMemberInfo represents a team member's info
type TeamMemberInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsLeader bool   `json:"is_leader"`
}

// ListTeams returns all teams with detailed information for admin
func (h *AdminTeamHandler) ListTeams(c *gin.Context) {
	teams, err := h.teamRepo.GetAllTeams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
func (h *AdminTeamHandler) UpdateTeam(c *gin.Context) {
	teamID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var req AdminUpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if team exists
	_, err = h.teamRepo.FindTeamByID(teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	update := bson.M{}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Description != "" {
		update["description"] = req.Description
	}

	if len(update) > 0 {
		if err := h.teamRepo.UpdateTeamFields(objID, update); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
func (h *AdminTeamHandler) UpdateTeamLeader(c *gin.Context) {
	teamID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var req UpdateTeamLeaderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team leader updated successfully"})
}

// RemoveMember removes a member from a team (admin only)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed from team"})
}

// DeleteTeam deletes a team (admin only, bypasses normal restrictions)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team deleted successfully"})
}
