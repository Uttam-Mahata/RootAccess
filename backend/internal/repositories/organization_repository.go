package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/database"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrganizationRepository struct {
	orgs   *mongo.Collection
	events *mongo.Collection
}

func NewOrganizationRepository() *OrganizationRepository {
	return &OrganizationRepository{
		orgs:   database.DB.Collection("organizations"),
		events: database.DB.Collection("events"),
	}
}

// ---- Organization CRUD ----

func (r *OrganizationRepository) CreateOrganization(org *models.Organization) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	org.CreatedAt = time.Now()
	org.UpdatedAt = time.Now()
	_, err := r.orgs.InsertOne(ctx, org)
	return err
}

func (r *OrganizationRepository) GetOrganizationByID(id string) (*models.Organization, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var org models.Organization
	if err := r.orgs.FindOne(ctx, bson.M{"_id": oid}).Decode(&org); err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *OrganizationRepository) GetOrganizationBySlug(slug string) (*models.Organization, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var org models.Organization
	if err := r.orgs.FindOne(ctx, bson.M{"slug": slug}).Decode(&org); err != nil {
		return nil, err
	}
	return &org, nil
}

// GetOrganizationByAPIKeyPrefix fetches the org record for key-prefix lookup so the
// caller can then bcrypt-verify the full key against APIKeyHash.
func (r *OrganizationRepository) GetOrganizationByAPIKeyPrefix(prefix string) (*models.Organization, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var org models.Organization
	if err := r.orgs.FindOne(ctx, bson.M{"api_key_prefix": prefix}).Decode(&org); err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *OrganizationRepository) SlugExists(slug string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	count, err := r.orgs.CountDocuments(ctx, bson.M{"slug": slug})
	return count > 0, err
}

// ---- Event CRUD ----

func (r *OrganizationRepository) CreateEvent(event *models.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()
	_, err := r.events.InsertOne(ctx, event)
	return err
}

func (r *OrganizationRepository) GetEventByID(id string) (*models.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var event models.Event
	if err := r.events.FindOne(ctx, bson.M{"_id": oid}).Decode(&event); err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *OrganizationRepository) ListEventsByOrg(orgID string) ([]models.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	oid, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return nil, err
	}
	cursor, err := r.events.Find(ctx, bson.M{"org_id": oid})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var events []models.Event
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *OrganizationRepository) UpdateEvent(id string, update *models.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update.UpdatedAt = time.Now()
	set := bson.M{
		"name":                   update.Name,
		"description":            update.Description,
		"start_time":             update.StartTime,
		"end_time":               update.EndTime,
		"freeze_time":            update.FreezeTime,
		"is_active":              update.IsActive,
		"is_paused":              update.IsPaused,
		"scoreboard_visibility":  update.ScoreboardVisibility,
		"frontend_url":           update.FrontendURL,
		"s3_config":              update.S3Config,
		"updated_at":             update.UpdatedAt,
	}
	if update.CustomMongoURI != "" {
		set["custom_mongo_uri"] = update.CustomMongoURI
	}
	_, err = r.events.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": set})
	return err
}

// GetEventByTokenPrefix fetches the event for token-prefix lookup so the caller
// can then bcrypt-verify the full token against EventTokenHash.
func (r *OrganizationRepository) GetEventByTokenPrefix(prefix string) (*models.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var event models.Event
	if err := r.events.FindOne(ctx, bson.M{"event_token_prefix": prefix}).Decode(&event); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("invalid event token")
		}
		return nil, err
	}
	return &event, nil
}

func (r *OrganizationRepository) EventSlugExistsForOrg(orgID, slug string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	oid, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return false, err
	}
	count, err := r.events.CountDocuments(ctx, bson.M{"org_id": oid, "slug": slug})
	return count > 0, err
}
