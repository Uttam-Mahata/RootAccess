package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/go-ctf-platform/backend/internal/database"
	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TeamService struct {
	teamRepo       *repositories.TeamRepository
	invitationRepo *repositories.TeamInvitationRepository
	userRepo       *repositories.UserRepository
	emailService   *EmailService
	submissionRepo *repositories.SubmissionRepository
	challengeRepo  *repositories.ChallengeRepository
}

func NewTeamService(
	teamRepo *repositories.TeamRepository,
	invitationRepo *repositories.TeamInvitationRepository,
	userRepo *repositories.UserRepository,
	emailService *EmailService,
	submissionRepo *repositories.SubmissionRepository,
	challengeRepo *repositories.ChallengeRepository,
) *TeamService {
	return &TeamService{
		teamRepo:       teamRepo,
		invitationRepo: invitationRepo,
		userRepo:       userRepo,
		emailService:   emailService,
		submissionRepo: submissionRepo,
		challengeRepo:  challengeRepo,
	}
}

func (s *TeamService) invalidateScoreboardCache() {
	if database.RDB != nil {
		ctx := context.Background()
		database.RDB.Del(ctx, "scoreboard")
		database.RDB.Del(ctx, "team_scoreboard")
	}
}

// generateInviteCode creates a unique invite code for the team
func (s *TeamService) generateInviteCode() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateTeam creates a new team with the user as leader
func (s *TeamService) CreateTeam(leaderID, name, description string) (*models.Team, error) {
	// Get user to check if email is verified
	user, err := s.userRepo.FindByID(leaderID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !user.EmailVerified {
		return nil, errors.New("please verify your email before creating a team")
	}

	// Check if user is already in a team
	existingTeam, _ := s.teamRepo.FindTeamByMemberID(leaderID)
	if existingTeam != nil {
		return nil, errors.New("you are already a member of a team")
	}

	// Check if team name already exists
	existingName, _ := s.teamRepo.FindTeamByName(name)
	if existingName != nil {
		return nil, errors.New("team name already exists")
	}

	// Generate invite code
	inviteCode, err := s.generateInviteCode()
	if err != nil {
		return nil, errors.New("failed to generate invite code")
	}

	leaderObjID, err := primitive.ObjectIDFromHex(leaderID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	team := &models.Team{
		Name:        name,
		Description: description,
		LeaderID:    leaderObjID,
		MemberIDs:   []primitive.ObjectID{leaderObjID}, // Leader is also a member
		InviteCode:  inviteCode,
		Score:       0,
	}

	if err := s.teamRepo.CreateTeam(team); err != nil {
		return nil, err
	}

	return team, nil
}

// Helper to calculate real-time dynamic score for a team
func (s *TeamService) calculateTeamScore(teamID primitive.ObjectID) int {
	// Get all correct submissions for the team
	submissions, err := s.submissionRepo.GetTeamSubmissions(teamID)
	if err != nil {
		return 0
	}

	totalScore := 0
	seenChallenges := make(map[string]bool)

	for _, sub := range submissions {
		challengeID := sub.ChallengeID.Hex()
		if seenChallenges[challengeID] {
			continue
		}
		seenChallenges[challengeID] = true

		// Get current challenge state (with up-to-date solve count)
		challenge, err := s.challengeRepo.GetChallengeByID(challengeID)
		if err == nil {
			// Add the CURRENT (decayed) value of the challenge
			totalScore += challenge.CurrentPoints()
		}
	}
	return totalScore
}

// GetTeamByID returns a team by its ID with dynamic score
func (s *TeamService) GetTeamByID(teamID string) (*models.Team, error) {
	team, err := s.teamRepo.FindTeamByID(teamID)
	if err != nil {
		return nil, err
	}
	
	// Recalculate score dynamically
	team.Score = s.calculateTeamScore(team.ID)
	
	return team, nil
}

// GetUserTeam returns the team that a user belongs to with dynamic score
func (s *TeamService) GetUserTeam(userID string) (*models.Team, error) {
	team, err := s.teamRepo.FindTeamByMemberID(userID)
	if err != nil {
		return nil, err
	}

	// Recalculate score dynamically
	team.Score = s.calculateTeamScore(team.ID)

	return team, nil
}

// UpdateTeam updates team name and description (leader only)
func (s *TeamService) UpdateTeam(teamID, leaderID, name, description string) (*models.Team, error) {
	team, err := s.teamRepo.FindTeamByID(teamID)
	if err != nil {
		return nil, errors.New("team not found")
	}

	if team.LeaderID.Hex() != leaderID {
		return nil, errors.New("only the team leader can update the team")
	}

	// Check if new name is unique (if changed)
	if team.Name != name {
		existingName, _ := s.teamRepo.FindTeamByName(name)
		if existingName != nil {
			return nil, errors.New("team name already exists")
		}
	}

	team.Name = name
	team.Description = description

	if err := s.teamRepo.UpdateTeam(team); err != nil {
		return nil, err
	}

	s.invalidateScoreboardCache()
	return team, nil
}

// DeleteTeam deletes a team (leader only, must have fewer than 2 members)
func (s *TeamService) DeleteTeam(teamID, leaderID string) error {
	team, err := s.teamRepo.FindTeamByID(teamID)
	if err != nil {
		return errors.New("team not found")
	}

	if team.LeaderID.Hex() != leaderID {
		return errors.New("only the team leader can delete the team")
	}

	if len(team.MemberIDs) >= models.MinTeamSize {
		return errors.New("cannot delete team with 2 or more members")
	}

	// Delete all invitations for this team
	if err := s.invitationRepo.DeleteInvitationsByTeam(teamID); err != nil {
		return err
	}

	return s.teamRepo.DeleteTeam(teamID)
}

// InviteByUsername invites a user to the team by their username
func (s *TeamService) InviteByUsername(teamID, inviterID, username string) (*models.TeamInvitation, error) {
	team, err := s.teamRepo.FindTeamByID(teamID)
	if err != nil {
		return nil, errors.New("team not found")
	}

	if team.LeaderID.Hex() != inviterID {
		return nil, errors.New("only the team leader can invite members")
	}

	// Check team size
	if len(team.MemberIDs) >= models.MaxTeamSize {
		return nil, errors.New("team is already at maximum capacity")
	}

	// Find user by username
	invitee, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if user is already in a team
	existingTeam, _ := s.teamRepo.FindTeamByMemberID(invitee.ID.Hex())
	if existingTeam != nil {
		return nil, errors.New("user is already in a team")
	}

	// Check if there's already a pending invitation
	hasPending, _ := s.invitationRepo.HasPendingInvitation(teamID, invitee.ID.Hex(), "")
	if hasPending {
		return nil, errors.New("invitation already sent to this user")
	}

	// Get inviter info
	inviter, _ := s.userRepo.FindByID(inviterID)

	// Generate invitation token
	token, err := s.emailService.GenerateVerificationToken()
	if err != nil {
		return nil, errors.New("failed to generate invitation token")
	}

	invitation := &models.TeamInvitation{
		TeamID:        team.ID,
		TeamName:      team.Name,
		InviterID:     team.LeaderID,
		InviterName:   inviter.Username,
		InviteeUserID: invitee.ID,
		Token:         token,
		Status:        models.InvitationStatusPending,
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.invitationRepo.CreateInvitation(invitation); err != nil {
		return nil, err
	}

	return invitation, nil
}

// InviteByEmail invites a user to the team by their email
func (s *TeamService) InviteByEmail(teamID, inviterID, email string) (*models.TeamInvitation, error) {
	team, err := s.teamRepo.FindTeamByID(teamID)
	if err != nil {
		return nil, errors.New("team not found")
	}

	if team.LeaderID.Hex() != inviterID {
		return nil, errors.New("only the team leader can invite members")
	}

	// Check team size
	if len(team.MemberIDs) >= models.MaxTeamSize {
		return nil, errors.New("team is already at maximum capacity")
	}

	// Check if user with email exists
	invitee, _ := s.userRepo.FindByEmail(email)
	if invitee != nil {
		// Check if user is already in a team
		existingTeam, _ := s.teamRepo.FindTeamByMemberID(invitee.ID.Hex())
		if existingTeam != nil {
			return nil, errors.New("user is already in a team")
		}
	}

	// Check if there's already a pending invitation
	hasPending, _ := s.invitationRepo.HasPendingInvitation(teamID, "", email)
	if hasPending {
		return nil, errors.New("invitation already sent to this email")
	}

	// Get inviter info
	inviter, _ := s.userRepo.FindByID(inviterID)

	// Generate invitation token
	token, err := s.emailService.GenerateVerificationToken()
	if err != nil {
		return nil, errors.New("failed to generate invitation token")
	}

	invitation := &models.TeamInvitation{
		TeamID:       team.ID,
		TeamName:     team.Name,
		InviterID:    team.LeaderID,
		InviterName:  inviter.Username,
		InviteeEmail: email,
		Token:        token,
		Status:       models.InvitationStatusPending,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if invitee != nil {
		invitation.InviteeUserID = invitee.ID
	}

	if err := s.invitationRepo.CreateInvitation(invitation); err != nil {
		return nil, err
	}

	// Send invitation email
	if err := s.emailService.SendTeamInvitationEmail(email, team.Name, inviter.Username, token); err != nil {
		// Log error but don't fail - invitation is created
		// User can still see it in their dashboard if they're registered
	}

	return invitation, nil
}

// JoinByInviteCode allows a user to join a team using the invite code
func (s *TeamService) JoinByInviteCode(userID, inviteCode string) (*models.Team, error) {
	// Get user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !user.EmailVerified {
		return nil, errors.New("please verify your email before joining a team")
	}

	// Check if user is already in a team
	existingTeam, _ := s.teamRepo.FindTeamByMemberID(userID)
	if existingTeam != nil {
		return nil, errors.New("you are already a member of a team")
	}

	// Find team by invite code
	team, err := s.teamRepo.FindTeamByInviteCode(inviteCode)
	if err != nil {
		return nil, errors.New("invalid invite code")
	}

	// Check team size
	if len(team.MemberIDs) >= models.MaxTeamSize {
		return nil, errors.New("team is already at maximum capacity")
	}

	// Add user to team
	if err := s.teamRepo.AddMemberToTeam(team.ID.Hex(), userID); err != nil {
		return nil, err
	}

	s.invalidateScoreboardCache()

	// Refresh team data
	return s.teamRepo.FindTeamByID(team.ID.Hex())
}

// GetPendingInvitations returns all pending invitations for a user
func (s *TeamService) GetPendingInvitations(userID, email string) ([]models.TeamInvitation, error) {
	// Clean up expired invitations first
	s.invitationRepo.DeleteExpiredInvitations()

	return s.invitationRepo.FindPendingInvitationsForUser(userID, email)
}

// AcceptInvitation accepts a team invitation
func (s *TeamService) AcceptInvitation(invitationID, userID string) (*models.Team, error) {
	invitation, err := s.invitationRepo.FindInvitationByID(invitationID)
	if err != nil {
		return nil, errors.New("invitation not found")
	}

	// Verify user matches invitation
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !invitation.InviteeUserID.IsZero() && invitation.InviteeUserID.Hex() != userID {
		if invitation.InviteeEmail != user.Email {
			return nil, errors.New("invitation is not for you")
		}
	} else if invitation.InviteeEmail != "" && invitation.InviteeEmail != user.Email {
		return nil, errors.New("invitation is not for you")
	}

	if invitation.Status != models.InvitationStatusPending {
		return nil, errors.New("invitation is no longer valid")
	}

	if time.Now().After(invitation.ExpiresAt) {
		s.invitationRepo.UpdateInvitationStatus(invitationID, models.InvitationStatusExpired)
		return nil, errors.New("invitation has expired")
	}

	// Check if user is already in a team
	existingTeam, _ := s.teamRepo.FindTeamByMemberID(userID)
	if existingTeam != nil {
		return nil, errors.New("you are already a member of a team")
	}

	// Get team and check capacity
	team, err := s.teamRepo.FindTeamByID(invitation.TeamID.Hex())
	if err != nil {
		return nil, errors.New("team no longer exists")
	}

	if len(team.MemberIDs) >= models.MaxTeamSize {
		return nil, errors.New("team is already at maximum capacity")
	}

	// Add user to team
	if err := s.teamRepo.AddMemberToTeam(invitation.TeamID.Hex(), userID); err != nil {
		return nil, err
	}

	// Update invitation status
	if err := s.invitationRepo.UpdateInvitationStatus(invitationID, models.InvitationStatusAccepted); err != nil {
		return nil, err
	}

	s.invalidateScoreboardCache()

	// Refresh team data
	return s.teamRepo.FindTeamByID(invitation.TeamID.Hex())
}

// RejectInvitation rejects a team invitation
func (s *TeamService) RejectInvitation(invitationID, userID string) error {
	invitation, err := s.invitationRepo.FindInvitationByID(invitationID)
	if err != nil {
		return errors.New("invitation not found")
	}

	// Verify user matches invitation
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !invitation.InviteeUserID.IsZero() && invitation.InviteeUserID.Hex() != userID {
		if invitation.InviteeEmail != user.Email {
			return errors.New("invitation is not for you")
		}
	} else if invitation.InviteeEmail != "" && invitation.InviteeEmail != user.Email {
		return errors.New("invitation is not for you")
	}

	if invitation.Status != models.InvitationStatusPending {
		return errors.New("invitation is no longer valid")
	}

	return s.invitationRepo.UpdateInvitationStatus(invitationID, models.InvitationStatusRejected)
}

// RemoveMember removes a member from the team (leader only)
func (s *TeamService) RemoveMember(teamID, leaderID, memberID string) error {
	team, err := s.teamRepo.FindTeamByID(teamID)
	if err != nil {
		return errors.New("team not found")
	}

	if team.LeaderID.Hex() != leaderID {
		return errors.New("only the team leader can remove members")
	}

	if leaderID == memberID {
		return errors.New("leader cannot remove themselves, use LeaveTeam instead")
	}

	// Verify member is in the team
	memberObjID, err := primitive.ObjectIDFromHex(memberID)
	if err != nil {
		return errors.New("invalid member ID")
	}

	found := false
	for _, id := range team.MemberIDs {
		if id == memberObjID {
			found = true
			break
		}
	}
	if !found {
		return errors.New("user is not a member of this team")
	}

	err = s.teamRepo.RemoveMemberFromTeam(teamID, memberID)
	if err == nil {
		s.invalidateScoreboardCache()
	}
	return err
}

// LeaveTeam allows a member to leave the team
func (s *TeamService) LeaveTeam(teamID, userID string) error {
	team, err := s.teamRepo.FindTeamByID(teamID)
	if err != nil {
		return errors.New("team not found")
	}

	// Check if user is the leader
	if team.LeaderID.Hex() == userID {
		if len(team.MemberIDs) > 1 {
			return errors.New("leader cannot leave team with other members. Transfer leadership or remove members first")
		}
		// Leader is the only member, delete the team
		return s.DeleteTeam(teamID, userID)
	}

	// Verify user is in the team
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	found := false
	for _, id := range team.MemberIDs {
		if id == userObjID {
			found = true
			break
		}
	}
	if !found {
		return errors.New("you are not a member of this team")
	}

	err = s.teamRepo.RemoveMemberFromTeam(teamID, userID)
	if err == nil {
		s.invalidateScoreboardCache()
	}
	return err
}

// RegenerateInviteCode generates a new invite code for the team
func (s *TeamService) RegenerateInviteCode(teamID, leaderID string) (string, error) {
	team, err := s.teamRepo.FindTeamByID(teamID)
	if err != nil {
		return "", errors.New("team not found")
	}

	if team.LeaderID.Hex() != leaderID {
		return "", errors.New("only the team leader can regenerate the invite code")
	}

	newCode, err := s.generateInviteCode()
	if err != nil {
		return "", errors.New("failed to generate new invite code")
	}

	team.InviteCode = newCode
	if err := s.teamRepo.UpdateTeam(team); err != nil {
		return "", err
	}

	return newCode, nil
}

// GetTeamMembers returns detailed information about team members
func (s *TeamService) GetTeamMembers(teamID string) ([]map[string]interface{}, error) {
	team, err := s.teamRepo.FindTeamByID(teamID)
	if err != nil {
		return nil, errors.New("team not found")
	}

	members := make([]map[string]interface{}, 0, len(team.MemberIDs))
	for _, memberID := range team.MemberIDs {
		user, err := s.userRepo.FindByID(memberID.Hex())
		if err != nil {
			continue
		}

		member := map[string]interface{}{
			"id":        user.ID.Hex(),
			"username":  user.Username,
			"is_leader": user.ID == team.LeaderID,
		}
		members = append(members, member)
	}

	return members, nil
}

// GetTeamPendingInvitations returns pending invitations sent by the team
func (s *TeamService) GetTeamPendingInvitations(teamID, leaderID string) ([]models.TeamInvitation, error) {
	team, err := s.teamRepo.FindTeamByID(teamID)
	if err != nil {
		return nil, errors.New("team not found")
	}

	if team.LeaderID.Hex() != leaderID {
		return nil, errors.New("only the team leader can view pending invitations")
	}

	return s.invitationRepo.FindPendingInvitationsByTeam(teamID)
}

// CancelInvitation cancels a pending invitation (leader only)
func (s *TeamService) CancelInvitation(invitationID, leaderID string) error {
	invitation, err := s.invitationRepo.FindInvitationByID(invitationID)
	if err != nil {
		return errors.New("invitation not found")
	}

	team, err := s.teamRepo.FindTeamByID(invitation.TeamID.Hex())
	if err != nil {
		return errors.New("team not found")
	}

	if team.LeaderID.Hex() != leaderID {
		return errors.New("only the team leader can cancel invitations")
	}

	if invitation.Status != models.InvitationStatusPending {
		return errors.New("invitation is no longer pending")
	}

	return s.invitationRepo.UpdateInvitationStatus(invitationID, models.InvitationStatusExpired)
}

// GetAllTeamsScoreboard returns all teams sorted by score
func (s *TeamService) GetAllTeamsScoreboard() ([]models.Team, error) {
	return s.teamRepo.GetAllTeamsWithScores()
}