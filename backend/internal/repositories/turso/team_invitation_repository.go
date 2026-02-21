package turso

import (
	"database/sql"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
)

const invitationColumns = `id, team_id, team_name, inviter_id, inviter_name, invitee_email, invitee_user_id, token, status, expires_at, created_at`

// TeamInvitationRepository implements interfaces.TeamInvitationRepository using database/sql.
type TeamInvitationRepository struct {
	db *sql.DB
}

// NewTeamInvitationRepository creates a new Turso-backed TeamInvitationRepository.
func NewTeamInvitationRepository(db *sql.DB) *TeamInvitationRepository {
	return &TeamInvitationRepository{db: db}
}

func scanInvitation(s scanner) (*models.TeamInvitation, error) {
	var inv models.TeamInvitation
	var (
		id, teamID, inviterID, inviteeUserID string
		teamName, inviterName                sql.NullString
		inviteeEmail                         sql.NullString
		token, status                        sql.NullString
		expiresAt, createdAt                 sql.NullString
	)

	err := s.Scan(
		&id, &teamID, &teamName, &inviterID, &inviterName,
		&inviteeEmail, &inviteeUserID, &token, &status,
		&expiresAt, &createdAt,
	)
	if err != nil {
		return nil, err
	}

	inv.ID = oidFromHex(id)
	inv.TeamID = oidFromHex(teamID)
	inv.TeamName = nullStr(teamName)
	inv.InviterID = oidFromHex(inviterID)
	inv.InviterName = nullStr(inviterName)
	inv.InviteeEmail = nullStr(inviteeEmail)
	inv.InviteeUserID = oidFromHex(inviteeUserID)
	inv.Token = nullStr(token)
	inv.Status = nullStr(status)
	inv.ExpiresAt = strToTime(nullStr(expiresAt))
	inv.CreatedAt = strToTime(nullStr(createdAt))

	return &inv, nil
}

func (r *TeamInvitationRepository) queryInvitations(query string, args ...interface{}) ([]models.TeamInvitation, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []models.TeamInvitation
	for rows.Next() {
		inv, err := scanInvitation(rows)
		if err != nil {
			return nil, err
		}
		invitations = append(invitations, *inv)
	}
	return invitations, rows.Err()
}

// ── Interface methods ──────────────────────────────────────────────────

func (r *TeamInvitationRepository) CreateInvitation(invitation *models.TeamInvitation) error {
	id := newID()
	now := time.Now()
	invitation.CreatedAt = now

	_, err := r.db.Exec(
		`INSERT INTO team_invitations (`+invitationColumns+`) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		id, invitation.TeamID.Hex(), invitation.TeamName,
		invitation.InviterID.Hex(), invitation.InviterName,
		invitation.InviteeEmail, invitation.InviteeUserID.Hex(),
		invitation.Token, invitation.Status,
		timeToStr(invitation.ExpiresAt), timeToStr(now),
	)
	if err != nil {
		return err
	}
	invitation.ID = oidFromHex(id)
	return nil
}

func (r *TeamInvitationRepository) FindInvitationByID(invitationID string) (*models.TeamInvitation, error) {
	row := r.db.QueryRow("SELECT "+invitationColumns+" FROM team_invitations WHERE id = ?", invitationID)
	return scanInvitation(row)
}

func (r *TeamInvitationRepository) FindInvitationByToken(token string) (*models.TeamInvitation, error) {
	row := r.db.QueryRow("SELECT "+invitationColumns+" FROM team_invitations WHERE token = ?", token)
	return scanInvitation(row)
}

func (r *TeamInvitationRepository) FindPendingInvitationsForUser(userID, email string) ([]models.TeamInvitation, error) {
	return r.queryInvitations(
		"SELECT "+invitationColumns+" FROM team_invitations WHERE status = 'pending' AND (invitee_user_id = ? OR invitee_email = ?)",
		userID, email,
	)
}

func (r *TeamInvitationRepository) FindInvitationsByTeam(teamID string) ([]models.TeamInvitation, error) {
	return r.queryInvitations(
		"SELECT "+invitationColumns+" FROM team_invitations WHERE team_id = ? ORDER BY created_at DESC",
		teamID,
	)
}

func (r *TeamInvitationRepository) FindPendingInvitationsByTeam(teamID string) ([]models.TeamInvitation, error) {
	return r.queryInvitations(
		"SELECT "+invitationColumns+" FROM team_invitations WHERE team_id = ? AND status = 'pending' ORDER BY created_at DESC",
		teamID,
	)
}

func (r *TeamInvitationRepository) UpdateInvitationStatus(invitationID, status string) error {
	_, err := r.db.Exec(`UPDATE team_invitations SET status = ? WHERE id = ?`, status, invitationID)
	return err
}

func (r *TeamInvitationRepository) DeleteExpiredInvitations() error {
	_, err := r.db.Exec(
		`DELETE FROM team_invitations WHERE status = 'pending' AND expires_at < ?`,
		timeToStr(time.Now()),
	)
	return err
}

func (r *TeamInvitationRepository) DeleteInvitationsByTeam(teamID string) error {
	_, err := r.db.Exec(`DELETE FROM team_invitations WHERE team_id = ?`, teamID)
	return err
}

func (r *TeamInvitationRepository) HasPendingInvitation(teamID, userID, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM team_invitations WHERE team_id = ? AND status = 'pending' AND (invitee_user_id = ? OR invitee_email = ?))`,
		teamID, userID, email,
	).Scan(&exists)
	return exists, err
}
