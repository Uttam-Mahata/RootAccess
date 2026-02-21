package turso

import (
	"database/sql"
	"strings"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const userColumns = `id, username, email, password_hash, role, email_verified,
	verification_token, verification_expiry, reset_password_token, reset_password_expiry,
	oauth_provider, oauth_provider_id, oauth_access_token, oauth_refresh_token, oauth_expires_at,
	status, ban_reason, suspended_until, last_ip, last_login_at, created_at, updated_at`

// validUserColumns guards against SQL injection in dynamic UpdateFields queries.
var validUserColumns = map[string]bool{
	"username": true, "email": true, "password_hash": true, "role": true,
	"email_verified": true, "verification_token": true, "verification_expiry": true,
	"reset_password_token": true, "reset_password_expiry": true,
	"status": true, "ban_reason": true, "suspended_until": true,
	"last_ip": true, "last_login_at": true,
	"oauth_provider": true, "oauth_provider_id": true,
	"oauth_access_token": true, "oauth_refresh_token": true, "oauth_expires_at": true,
}

// UserRepository implements interfaces.UserRepository using database/sql.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new Turso-backed UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...interface{}) error
}

func scanUser(s scanner) (*models.User, error) {
	var u models.User
	var (
		id            string
		emailVerified int
		pwHash, role  sql.NullString

		verToken, verExpiry     sql.NullString
		resetToken, resetExpiry sql.NullString

		oaProv, oaPID, oaAT, oaRT, oaExp sql.NullString

		status, banReason, suspUntil         sql.NullString
		lastIP, lastLogin, createdAt, updatedAt sql.NullString
	)

	err := s.Scan(
		&id, &u.Username, &u.Email, &pwHash, &role, &emailVerified,
		&verToken, &verExpiry, &resetToken, &resetExpiry,
		&oaProv, &oaPID, &oaAT, &oaRT, &oaExp,
		&status, &banReason, &suspUntil, &lastIP, &lastLogin, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	u.ID = oidFromHex(id)
	u.PasswordHash = nullStr(pwHash)
	u.Role = nullStr(role)
	u.EmailVerified = intToBool(emailVerified)
	u.VerificationToken = nullStr(verToken)
	u.VerificationExpiry = strToTime(nullStr(verExpiry))
	u.ResetPasswordToken = nullStr(resetToken)
	u.ResetPasswordExpiry = strToTime(nullStr(resetExpiry))
	u.Status = nullStr(status)
	u.BanReason = nullStr(banReason)
	u.SuspendedUntil = strToTimePtr(nullStr(suspUntil))
	u.LastIP = nullStr(lastIP)
	u.LastLoginAt = strToTimePtr(nullStr(lastLogin))
	u.CreatedAt = strToTime(nullStr(createdAt))
	u.UpdatedAt = strToTime(nullStr(updatedAt))

	if prov := nullStr(oaProv); prov != "" {
		u.OAuth = &models.OAuth{
			Provider:     prov,
			ProviderID:   nullStr(oaPID),
			AccessToken:  nullStr(oaAT),
			RefreshToken: nullStr(oaRT),
			ExpiresAt:    strToTime(nullStr(oaExp)),
		}
	}

	return &u, nil
}

func (r *UserRepository) findOne(where string, args ...interface{}) (*models.User, error) {
	row := r.db.QueryRow("SELECT "+userColumns+" FROM users WHERE "+where, args...)
	u, err := scanUser(row)
	if err != nil {
		return nil, err
	}
	u.IPHistory, err = r.loadIPHistory(u.ID.Hex())
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) queryUsers(query string, args ...interface{}) ([]models.User, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		u.IPHistory, err = r.loadIPHistory(u.ID.Hex())
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

func (r *UserRepository) loadIPHistory(userID string) ([]models.IPRecord, error) {
	rows, err := r.db.Query(
		`SELECT ip, action, timestamp FROM user_ip_history WHERE user_id = ? ORDER BY id DESC LIMIT 50`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.IPRecord
	for rows.Next() {
		var ip, action, ts sql.NullString
		if err := rows.Scan(&ip, &action, &ts); err != nil {
			return nil, err
		}
		records = append(records, models.IPRecord{
			IP:        nullStr(ip),
			Action:    nullStr(action),
			Timestamp: strToTime(nullStr(ts)),
		})
	}
	return records, rows.Err()
}

// ── Interface methods ──────────────────────────────────────────────────

func (r *UserRepository) CreateUser(user *models.User) error {
	now := time.Now()
	id := newID()
	user.CreatedAt = now
	user.UpdatedAt = now

	var oaProv, oaPID, oaAT, oaRT, oaExp string
	if user.OAuth != nil {
		oaProv = user.OAuth.Provider
		oaPID = user.OAuth.ProviderID
		oaAT = user.OAuth.AccessToken
		oaRT = user.OAuth.RefreshToken
		oaExp = timeToStr(user.OAuth.ExpiresAt)
	}

	_, err := r.db.Exec(
		`INSERT INTO users (`+userColumns+`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		id, user.Username, user.Email, user.PasswordHash, user.Role, boolToInt(user.EmailVerified),
		user.VerificationToken, timeToStr(user.VerificationExpiry),
		user.ResetPasswordToken, timeToStr(user.ResetPasswordExpiry),
		oaProv, oaPID, oaAT, oaRT, oaExp,
		user.Status, user.BanReason, timePtrToStr(user.SuspendedUntil),
		user.LastIP, timePtrToStr(user.LastLoginAt),
		timeToStr(now), timeToStr(now),
	)
	if err != nil {
		return err
	}
	user.ID = oidFromHex(id)
	return nil
}

func (r *UserRepository) UpdateUser(user *models.User) error {
	user.UpdatedAt = time.Now()

	var oaProv, oaPID, oaAT, oaRT, oaExp string
	if user.OAuth != nil {
		oaProv = user.OAuth.Provider
		oaPID = user.OAuth.ProviderID
		oaAT = user.OAuth.AccessToken
		oaRT = user.OAuth.RefreshToken
		oaExp = timeToStr(user.OAuth.ExpiresAt)
	}

	_, err := r.db.Exec(
		`UPDATE users SET username=?, email=?, password_hash=?, role=?, email_verified=?,
		verification_token=?, verification_expiry=?, reset_password_token=?, reset_password_expiry=?,
		oauth_provider=?, oauth_provider_id=?, oauth_access_token=?, oauth_refresh_token=?, oauth_expires_at=?,
		status=?, ban_reason=?, suspended_until=?, last_ip=?, last_login_at=?,
		updated_at=? WHERE id=?`,
		user.Username, user.Email, user.PasswordHash, user.Role, boolToInt(user.EmailVerified),
		user.VerificationToken, timeToStr(user.VerificationExpiry),
		user.ResetPasswordToken, timeToStr(user.ResetPasswordExpiry),
		oaProv, oaPID, oaAT, oaRT, oaExp,
		user.Status, user.BanReason, timePtrToStr(user.SuspendedUntil),
		user.LastIP, timePtrToStr(user.LastLoginAt),
		timeToStr(user.UpdatedAt), user.ID.Hex(),
	)
	return err
}

func (r *UserRepository) FindByID(userID string) (*models.User, error) {
	return r.findOne("id = ?", userID)
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	return r.findOne("username = ?", username)
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	return r.findOne("email = ?", email)
}

func (r *UserRepository) FindByVerificationToken(token string) (*models.User, error) {
	return r.findOne("verification_token = ?", token)
}

func (r *UserRepository) FindByResetToken(token string) (*models.User, error) {
	return r.findOne("reset_password_token = ?", token)
}

func (r *UserRepository) FindByProviderID(provider, providerID string) (*models.User, error) {
	return r.findOne("oauth_provider = ? AND oauth_provider_id = ?", provider, providerID)
}

func (r *UserRepository) GetAllUsers() ([]models.User, error) {
	return r.queryUsers("SELECT " + userColumns + " FROM users ORDER BY created_at DESC")
}

func (r *UserRepository) GetUsersWithDetails() ([]models.User, error) {
	return r.GetAllUsers()
}

func (r *UserRepository) UpdateFields(userID primitive.ObjectID, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	var setClauses []string
	var args []interface{}

	for key, val := range fields {
		switch key {
		case "_id", "ip_history":
			continue
		case "oauth":
			if oauth, ok := val.(*models.OAuth); ok && oauth != nil {
				setClauses = append(setClauses,
					"oauth_provider = ?", "oauth_provider_id = ?",
					"oauth_access_token = ?", "oauth_refresh_token = ?", "oauth_expires_at = ?")
				args = append(args, oauth.Provider, oauth.ProviderID,
					oauth.AccessToken, oauth.RefreshToken, timeToStr(oauth.ExpiresAt))
			} else {
				setClauses = append(setClauses,
					"oauth_provider = ?", "oauth_provider_id = ?",
					"oauth_access_token = ?", "oauth_refresh_token = ?", "oauth_expires_at = ?")
				args = append(args, "", "", "", "", "")
			}
		default:
			if !validUserColumns[key] {
				continue
			}
			setClauses = append(setClauses, key+" = ?")
			args = append(args, convertFieldValue(val))
		}
	}

	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, timeToStr(time.Now()))

	args = append(args, userID.Hex())
	query := "UPDATE users SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"
	_, err := r.db.Exec(query, args...)
	return err
}

func convertFieldValue(val interface{}) interface{} {
	switch v := val.(type) {
	case bool:
		return boolToInt(v)
	case time.Time:
		return timeToStr(v)
	case *time.Time:
		return timePtrToStr(v)
	default:
		return val
	}
}

func (r *UserRepository) CountUsers() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func (r *UserRepository) RecordUserIP(userID primitive.ObjectID, ip string, action string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`INSERT INTO user_ip_history (user_id, ip, action, timestamp) VALUES (?, ?, ?, ?)`,
		userID.Hex(), ip, action, timeToStr(now),
	)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(
		`UPDATE users SET last_ip = ?, last_login_at = ?, updated_at = ? WHERE id = ?`,
		ip, timeToStr(now), timeToStr(now), userID.Hex(),
	)
	return err
}

func (r *UserRepository) CountUsersByStatus(status string) (int64, error) {
	var count int64
	var err error
	if status == "active" || status == "" {
		err = r.db.QueryRow(
			`SELECT COUNT(*) FROM users WHERE status = 'active' OR status = '' OR status IS NULL`,
		).Scan(&count)
	} else {
		err = r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE status = ?`, status).Scan(&count)
	}
	return count, err
}

func (r *UserRepository) CountVerifiedUsers() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE email_verified = 1`).Scan(&count)
	return count, err
}

func (r *UserRepository) CountAdmins() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&count)
	return count, err
}

func (r *UserRepository) GetRecentUsers(since time.Time) ([]models.User, error) {
	return r.queryUsers(
		"SELECT "+userColumns+" FROM users WHERE created_at >= ? ORDER BY created_at DESC",
		timeToStr(since),
	)
}
