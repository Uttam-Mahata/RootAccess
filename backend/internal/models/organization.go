package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Organization represents a tenant/customer on the RootAccess SaaS platform.
// Each organization can host one or more CTF events.
type Organization struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string             `bson:"name" json:"name"`
	Slug           string             `bson:"slug" json:"slug"` // URL-safe unique identifier
	OwnerEmail     string             `bson:"owner_email" json:"owner_email"`
	OwnerName      string             `bson:"owner_name" json:"owner_name"`
	APIKeyHash     string             `bson:"api_key_hash" json:"-"` // bcrypt hash of the API key
	APIKeyPrefix   string             `bson:"api_key_prefix" json:"api_key_prefix"` // e.g. "ra_org_abc1" for display
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

// S3Config holds object-storage credentials for a CTF event.
// Organizers can point challenges at their own S3-compatible bucket.
type S3Config struct {
	Endpoint  string `bson:"endpoint" json:"endpoint"`   // e.g. "s3.amazonaws.com" or custom MinIO
	Bucket    string `bson:"bucket" json:"bucket"`
	Region    string `bson:"region" json:"region"`
	AccessKey string `bson:"access_key" json:"access_key"`
	SecretKey string `bson:"secret_key" json:"-"` // never returned in API responses
	PublicURL string `bson:"public_url" json:"public_url"` // CDN / public bucket URL for file links
}

// Event represents a single CTF competition hosted by an Organization.
// It extends the concept of ContestConfig with multi-tenant and storage fields.
type Event struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID                primitive.ObjectID `bson:"org_id" json:"org_id"`
	Name                 string             `bson:"name" json:"name"`
	Slug                 string             `bson:"slug" json:"slug"` // unique within org
	Description          string             `bson:"description,omitempty" json:"description,omitempty"`
	StartTime            time.Time          `bson:"start_time" json:"start_time"`
	EndTime              time.Time          `bson:"end_time" json:"end_time"`
	FreezeTime           *time.Time         `bson:"freeze_time,omitempty" json:"freeze_time,omitempty"`
	IsActive             bool               `bson:"is_active" json:"is_active"`
	IsPaused             bool               `bson:"is_paused" json:"is_paused"`
	ScoreboardVisibility string             `bson:"scoreboard_visibility" json:"scoreboard_visibility"` // "public","private","hidden"
	FrontendURL          string             `bson:"frontend_url,omitempty" json:"frontend_url,omitempty"`
	CustomMongoURI       string             `bson:"custom_mongo_uri,omitempty" json:"-"` // optional custom DB; never in JSON
	S3Config             *S3Config          `bson:"s3_config,omitempty" json:"s3_config,omitempty"`
	EventTokenHash       string             `bson:"event_token_hash" json:"-"`   // bcrypt hash of the event token
	EventTokenPrefix     string             `bson:"event_token_prefix" json:"event_token_prefix"` // e.g. "evt_abc1" for display
	CreatedAt            time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time          `bson:"updated_at" json:"updated_at"`
}
