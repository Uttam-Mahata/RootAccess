package models

import "time"

type User struct {
	ID                  string    `json:"id"`
	Username            string    `json:"username"`
	Email               string    `json:"email"`
	PasswordHash        string    `json:"-"`
	Role                string    `json:"role"`
	EmailVerified       bool      `json:"email_verified"`
	VerificationToken   string    `json:"-"`
	VerificationExpiry  string    `json:"-"`
	ResetPasswordToken  string    `json:"-"`
	ResetPasswordExpiry string    `json:"-"`
	OAuthProvider       string    `json:"oauth_provider,omitempty"`
	OAuthProviderID     string    `json:"oauth_provider_id,omitempty"`
	OAuthAccessToken    string    `json:"-"`
	OAuthRefreshToken   string    `json:"-"`
	OAuthExpiresAt      string    `json:"-"`
	Status              string    `json:"status"`
	BanReason           string    `json:"ban_reason,omitempty"`
	SuspendedUntil      string    `json:"suspended_until,omitempty"`
	LastIP              string    `json:"last_ip,omitempty"`
	LastLoginAt         string    `json:"last_login_at,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type IPRecord struct {
	IP        string    `json:"ip"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action,omitempty"`
}
