package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username            string             `bson:"username" json:"username"`
	Email               string             `bson:"email" json:"email"`
	PasswordHash        string             `bson:"password_hash" json:"-"`
	Role                string             `bson:"role" json:"role"` // "admin" or "user"
	EmailVerified       bool               `bson:"email_verified" json:"email_verified"`
	VerificationToken   string             `bson:"verification_token,omitempty" json:"-"`
	VerificationExpiry  time.Time          `bson:"verification_expiry,omitempty" json:"-"`
	ResetPasswordToken  string             `bson:"reset_password_token,omitempty" json:"-"`
	ResetPasswordExpiry time.Time          `bson:"reset_password_expiry,omitempty" json:"-"`
	OAuth               *OAuth             `bson:"oauth,omitempty" json:"oauth,omitempty"`
	Status              string             `bson:"status" json:"status"`                                   // "active", "banned", "suspended"
	BanReason           string             `bson:"ban_reason,omitempty" json:"ban_reason,omitempty"`
	SuspendedUntil      *time.Time         `bson:"suspended_until,omitempty" json:"suspended_until,omitempty"`
	CreatedAt           time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt           time.Time          `bson:"updated_at" json:"updated_at"`
}

type OAuth struct {
	Provider     string    `bson:"provider" json:"provider"` // "google", "github", etc.
	ProviderID   string    `bson:"provider_id" json:"provider_id"`
	AccessToken  string    `bson:"access_token,omitempty" json:"-"`
	RefreshToken string    `bson:"refresh_token,omitempty" json:"-"`
	ExpiresAt    time.Time `bson:"expires_at,omitempty" json:"-"`
}
