package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/google/uuid"
)

type TeamInvitationRepository struct {
	db *sql.DB
}

func NewTeamInvitationRepository(db *sql.DB) *TeamInvitationRepository {
	return &TeamInvitationRepository{db: db}
}

func (r *TeamInvitationRepository) CreateInvitation(inv *models.TeamInvitation) error {
	if inv.ID == "" {
		inv.ID = uuid.New().String()
	}
	inv.CreatedAt = time.Now()

	query := `INSERT INTO team_invitations (id, team_id, team_name, inviter_id, inviter_name, invitee_email, invitee_user_id, token, status, expires_at, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, inv.ID, inv.TeamID, inv.TeamName, inv.InviterID, inv.InviterName, inv.InviteeEmail, inv.InviteeUserID, inv.Token, inv.Status, inv.ExpiresAt.Format(time.RFC3339), inv.CreatedAt.Format(time.RFC3339))
	return err
}

func (r *TeamInvitationRepository) scanInvitations(rows *sql.Rows) ([]models.TeamInvitation, error) {
	var invs []models.TeamInvitation
	for rows.Next() {
		var inv models.TeamInvitation
		var expiresAt, createdAt string
		if err := rows.Scan(&inv.ID, &inv.TeamID, &inv.TeamName, &inv.InviterID, &inv.InviterName, &inv.InviteeEmail, &inv.InviteeUserID, &inv.Token, &inv.Status, &expiresAt, &createdAt); err != nil {
			return nil, err
		}
		inv.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAt)
		inv.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		invs = append(invs, inv)
	}
	return invs, nil
}

func (r *TeamInvitationRepository) selectFields() string {
	return "id, team_id, team_name, inviter_id, inviter_name, invitee_email, invitee_user_id, token, status, expires_at, created_at"
}

func (r *TeamInvitationRepository) FindInvitationByID(id string) (*models.TeamInvitation, error) {
	query := fmt.Sprintf("SELECT %s FROM team_invitations WHERE id=?", r.selectFields())
	rows, err := r.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	invs, err := r.scanInvitations(rows)
	if err != nil || len(invs) == 0 {
		return nil, sql.ErrNoRows
	}
	return &invs[0], nil
}

func (r *TeamInvitationRepository) FindInvitationByToken(token string) (*models.TeamInvitation, error) {
	query := fmt.Sprintf("SELECT %s FROM team_invitations WHERE token=?", r.selectFields())
	rows, err := r.db.Query(query, token)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	invs, err := r.scanInvitations(rows)
	if err != nil || len(invs) == 0 {
		return nil, sql.ErrNoRows
	}
	return &invs[0], nil
}

func (r *TeamInvitationRepository) FindPendingInvitationsForUser(userID, email string) ([]models.TeamInvitation, error) {
	if userID != "" && email != "" {
		query := fmt.Sprintf("SELECT %s FROM team_invitations WHERE status=? AND (invitee_user_id=? OR invitee_email=?)", r.selectFields())
		rows, err := r.db.Query(query, models.InvitationStatusPending, userID, email)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return r.scanInvitations(rows)
	} else if userID != "" {
		query := fmt.Sprintf("SELECT %s FROM team_invitations WHERE status=? AND invitee_user_id=?", r.selectFields())
		rows, err := r.db.Query(query, models.InvitationStatusPending, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return r.scanInvitations(rows)
	} else if email != "" {
		query := fmt.Sprintf("SELECT %s FROM team_invitations WHERE status=? AND invitee_email=?", r.selectFields())
		rows, err := r.db.Query(query, models.InvitationStatusPending, email)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return r.scanInvitations(rows)
	}
	return []models.TeamInvitation{}, nil
}

func (r *TeamInvitationRepository) FindInvitationsByTeam(teamID string) ([]models.TeamInvitation, error) {
	query := fmt.Sprintf("SELECT %s FROM team_invitations WHERE team_id=?", r.selectFields())
	rows, err := r.db.Query(query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanInvitations(rows)
}

func (r *TeamInvitationRepository) FindPendingInvitationsByTeam(teamID string) ([]models.TeamInvitation, error) {
	query := fmt.Sprintf("SELECT %s FROM team_invitations WHERE team_id=? AND status=?", r.selectFields())
	rows, err := r.db.Query(query, teamID, models.InvitationStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanInvitations(rows)
}

func (r *TeamInvitationRepository) UpdateInvitationStatus(id, status string) error {
	_, err := r.db.Exec("UPDATE team_invitations SET status=? WHERE id=?", status, id)
	return err
}

func (r *TeamInvitationRepository) DeleteExpiredInvitations() error {
	_, err := r.db.Exec("UPDATE team_invitations SET status=? WHERE status=? AND expires_at < ?",
		models.InvitationStatusExpired, models.InvitationStatusPending, time.Now().Format(time.RFC3339))
	return err
}

func (r *TeamInvitationRepository) DeleteInvitationsByTeam(teamID string) error {
	_, err := r.db.Exec("DELETE FROM team_invitations WHERE team_id=?", teamID)
	return err
}

func (r *TeamInvitationRepository) HasPendingInvitation(teamID, userID, email string) (bool, error) {
	var count int
	var err error
	if userID != "" {
		err = r.db.QueryRow("SELECT COUNT(*) FROM team_invitations WHERE team_id=? AND status=? AND invitee_user_id=?", teamID, models.InvitationStatusPending, userID).Scan(&count)
	} else if email != "" {
		err = r.db.QueryRow("SELECT COUNT(*) FROM team_invitations WHERE team_id=? AND status=? AND invitee_email=?", teamID, models.InvitationStatusPending, email).Scan(&count)
	} else {
		return false, nil
	}
	return count > 0, err
}
