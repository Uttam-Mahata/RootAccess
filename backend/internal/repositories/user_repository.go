package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (
			id, username, email, password_hash, role, email_verified,
			verification_token, verification_expiry, reset_password_token,
			reset_password_expiry, oauth_provider, oauth_provider_id,
			oauth_access_token, oauth_refresh_token, oauth_expires_at,
			status, ban_reason, suspended_until, last_ip, last_login_at,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.Role, user.EmailVerified,
		user.VerificationToken, user.VerificationExpiry, user.ResetPasswordToken,
		user.ResetPasswordExpiry, user.OAuthProvider, user.OAuthProviderID,
		user.OAuthAccessToken, user.OAuthRefreshToken, user.OAuthExpiresAt,
		user.Status, user.BanReason, user.SuspendedUntil, user.LastIP, user.LastLoginAt,
		user.CreatedAt.Format(time.RFC3339), user.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *UserRepository) UpdateUser(user *models.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users SET
			username=?, email=?, password_hash=?, role=?, email_verified=?,
			verification_token=?, verification_expiry=?, reset_password_token=?,
			reset_password_expiry=?, oauth_provider=?, oauth_provider_id=?,
			oauth_access_token=?, oauth_refresh_token=?, oauth_expires_at=?,
			status=?, ban_reason=?, suspended_until=?, last_ip=?, last_login_at=?,
			updated_at=?
		WHERE id=?
	`

	_, err := r.db.Exec(query,
		user.Username, user.Email, user.PasswordHash, user.Role, user.EmailVerified,
		user.VerificationToken, user.VerificationExpiry, user.ResetPasswordToken,
		user.ResetPasswordExpiry, user.OAuthProvider, user.OAuthProviderID,
		user.OAuthAccessToken, user.OAuthRefreshToken, user.OAuthExpiresAt,
		user.Status, user.BanReason, user.SuspendedUntil, user.LastIP, user.LastLoginAt,
		user.UpdatedAt.Format(time.RFC3339), user.ID,
	)
	return err
}

func (r *UserRepository) scanUser(row *sql.Row) (*models.User, error) {
	var user models.User
	var createdAt, updatedAt string
	err := row.Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.EmailVerified,
		&user.VerificationToken, &user.VerificationExpiry, &user.ResetPasswordToken,
		&user.ResetPasswordExpiry, &user.OAuthProvider, &user.OAuthProviderID,
		&user.OAuthAccessToken, &user.OAuthRefreshToken, &user.OAuthExpiresAt,
		&user.Status, &user.BanReason, &user.SuspendedUntil, &user.LastIP, &user.LastLoginAt,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &user, nil
}

func (r *UserRepository) scanUsers(rows *sql.Rows) ([]models.User, error) {
	var users []models.User
	for rows.Next() {
		var user models.User
		var createdAt, updatedAt string
		if err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.EmailVerified,
			&user.VerificationToken, &user.VerificationExpiry, &user.ResetPasswordToken,
			&user.ResetPasswordExpiry, &user.OAuthProvider, &user.OAuthProviderID,
			&user.OAuthAccessToken, &user.OAuthRefreshToken, &user.OAuthExpiresAt,
			&user.Status, &user.BanReason, &user.SuspendedUntil, &user.LastIP, &user.LastLoginAt,
			&createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepository) selectUserFields() string {
	return `id, username, email, password_hash, role, email_verified,
			verification_token, verification_expiry, reset_password_token,
			reset_password_expiry, oauth_provider, oauth_provider_id,
			oauth_access_token, oauth_refresh_token, oauth_expires_at,
			status, ban_reason, suspended_until, last_ip, last_login_at,
			created_at, updated_at`
}

func (r *UserRepository) FindByID(userID string) (*models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE id=?", r.selectUserFields())
	return r.scanUser(r.db.QueryRow(query, userID))
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE username=?", r.selectUserFields())
	return r.scanUser(r.db.QueryRow(query, username))
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE email=?", r.selectUserFields())
	return r.scanUser(r.db.QueryRow(query, email))
}

func (r *UserRepository) FindByVerificationToken(token string) (*models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE verification_token=?", r.selectUserFields())
	return r.scanUser(r.db.QueryRow(query, token))
}

func (r *UserRepository) FindByResetToken(token string) (*models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE reset_password_token=?", r.selectUserFields())
	return r.scanUser(r.db.QueryRow(query, token))
}

func (r *UserRepository) FindByProviderID(provider, providerID string) (*models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE oauth_provider=? AND oauth_provider_id=?", r.selectUserFields())
	return r.scanUser(r.db.QueryRow(query, provider, providerID))
}

func (r *UserRepository) GetAllUsers() ([]models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users", r.selectUserFields())
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanUsers(rows)
}

func (r *UserRepository) UpdateFields(userID string, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	fields["updated_at"] = time.Now().Format(time.RFC3339)

	var setClauses []string
	var args []interface{}

	for k, v := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s=?", k))
		if b, ok := v.(bool); ok {
			if b {
				args = append(args, 1)
			} else {
				args = append(args, 0)
			}
		} else {
			args = append(args, v)
		}
	}
	args = append(args, userID)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id=?", strings.Join(setClauses, ", "))
	_, err := r.db.Exec(query, args...)
	return err
}

func (r *UserRepository) RecordUserIP(userID string, ip string, action string) error {
	now := time.Now().Format(time.RFC3339)

	updateQuery := `UPDATE users SET last_ip=?, last_login_at=?, updated_at=? WHERE id=?`
	if _, err := r.db.Exec(updateQuery, ip, now, now, userID); err != nil {
		return err
	}

	insertHist := `INSERT INTO user_ip_history (user_id, ip, action, timestamp) VALUES (?, ?, ?, ?)`
	if _, err := r.db.Exec(insertHist, userID, ip, action, now); err != nil {
		return err
	}

	deleteHist := `
		DELETE FROM user_ip_history 
		WHERE user_id = ? AND timestamp NOT IN (
			SELECT timestamp FROM user_ip_history 
			WHERE user_id = ? 
			ORDER BY timestamp DESC 
			LIMIT 50
		)
	`
	_, err := r.db.Exec(deleteHist, userID, userID)
	return err
}

func (r *UserRepository) GetUsersWithDetails() ([]models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users ORDER BY created_at DESC", r.selectUserFields())
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

func (r *UserRepository) CountUsers() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (r *UserRepository) CountUsersByStatus(status string) (int64, error) {
	var count int64
	var err error
	if status == "active" || status == "" {
		err = r.db.QueryRow("SELECT COUNT(*) FROM users WHERE status='active' OR status IS NULL OR status=''").Scan(&count)
	} else {
		err = r.db.QueryRow("SELECT COUNT(*) FROM users WHERE status=?", status).Scan(&count)
	}
	return count, err
}

func (r *UserRepository) CountVerifiedUsers() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE email_verified=1").Scan(&count)
	return count, err
}

func (r *UserRepository) CountAdmins() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE role='admin'").Scan(&count)
	return count, err
}

func (r *UserRepository) GetRecentUsers(since time.Time) ([]models.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE created_at >= ? ORDER BY created_at DESC", r.selectUserFields())
	rows, err := r.db.Query(query, since.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanUsers(rows)
}
