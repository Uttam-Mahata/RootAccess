package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

var slugRe = regexp.MustCompile(`^[a-z0-9-]+$`)

// OrganizationService handles business logic for organizations and events.
type OrganizationService struct {
	repo *repositories.OrganizationRepository
}

func NewOrganizationService(repo *repositories.OrganizationRepository) *OrganizationService {
	return &OrganizationService{repo: repo}
}

// --- helpers ---

// generateSecureToken produces a cryptographically random hex token.
func generateSecureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// toSlug converts a name to a URL-safe slug.
func toSlug(s string) string {
	s = strings.ToLower(s)
	// Replace spaces/underscores/dots with hyphens
	s = strings.NewReplacer(" ", "-", "_", "-", ".", "-").Replace(s)
	// Remove any character that is not a-z, 0-9, or hyphen
	var out strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			out.WriteRune(r)
		}
	}
	result := strings.Trim(out.String(), "-")
	return result
}

// --- Organization ---

type CreateOrgInput struct {
	Name       string
	OwnerEmail string
	OwnerName  string
	// Slug is optional; derived from Name when empty
	Slug string
}

// CreateOrganizationResult carries the plain-text API key back to the caller
// (stored only once; the hash is persisted).
type CreateOrganizationResult struct {
	Org    *models.Organization
	APIKey string // plain-text; show once to the user
}

func (s *OrganizationService) CreateOrganization(input CreateOrgInput) (*CreateOrganizationResult, error) {
	slug := input.Slug
	if slug == "" {
		slug = toSlug(input.Name)
	}
	if !slugRe.MatchString(slug) {
		return nil, errors.New("slug must contain only lowercase letters, digits and hyphens")
	}

	exists, err := s.repo.SlugExists(slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("an organization with this slug already exists")
	}

	// Generate API key: format  ra_org_<32 hex chars>
	rawToken, err := generateSecureToken(16)
	if err != nil {
		return nil, err
	}
	apiKey := "ra_org_" + rawToken
	prefix := apiKey[:12] // "ra_org_" (7 chars) + first 5 hex chars of token

	hash, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	org := &models.Organization{
		Name:         input.Name,
		Slug:         slug,
		OwnerEmail:   input.OwnerEmail,
		OwnerName:    input.OwnerName,
		APIKeyHash:   string(hash),
		APIKeyPrefix: prefix,
	}

	if err := s.repo.CreateOrganization(org); err != nil {
		return nil, err
	}

	return &CreateOrganizationResult{Org: org, APIKey: apiKey}, nil
}

func (s *OrganizationService) GetOrganizationByID(id string) (*models.Organization, error) {
	return s.repo.GetOrganizationByID(id)
}

// ValidateAPIKey verifies a plain-text API key against the stored hash.
// Returns the matching organization or an error.
func (s *OrganizationService) ValidateAPIKey(apiKey string) (*models.Organization, error) {
	if len(apiKey) < 12 {
		return nil, errors.New("invalid API key")
	}
	prefix := apiKey[:12]
	org, err := s.repo.GetOrganizationByAPIKeyPrefix(prefix)
	if err != nil {
		return nil, errors.New("invalid API key")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(org.APIKeyHash), []byte(apiKey)); err != nil {
		return nil, errors.New("invalid API key")
	}
	return org, nil
}

// --- Event ---

type CreateEventInput struct {
	OrgID                string
	Name                 string
	Slug                 string // optional; derived from Name
	Description          string
	StartTime            time.Time
	EndTime              time.Time
	FreezeTime           *time.Time
	ScoreboardVisibility string
	FrontendURL          string
	CustomMongoURI       string
	S3Config             *models.S3Config
}

// CreateEventResult carries the plain-text event token back to the caller.
type CreateEventResult struct {
	Event      *models.Event
	EventToken string // plain-text; show once to the user
}

func (s *OrganizationService) CreateEvent(input CreateEventInput) (*CreateEventResult, error) {
	slug := input.Slug
	if slug == "" {
		slug = toSlug(input.Name)
	}
	if !slugRe.MatchString(slug) {
		return nil, errors.New("slug must contain only lowercase letters, digits and hyphens")
	}

	exists, err := s.repo.EventSlugExistsForOrg(input.OrgID, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("an event with this slug already exists for this organization")
	}

	visibility := input.ScoreboardVisibility
	if visibility == "" {
		visibility = "public"
	}

	// Generate event token: format  evt_<32 hex chars>
	rawToken, err := generateSecureToken(16)
	if err != nil {
		return nil, err
	}
	eventToken := "evt_" + rawToken
	prefix := eventToken[:8] // "evt_" + first 4 chars of token

	hash, err := bcrypt.GenerateFromPassword([]byte(eventToken), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	org, err := s.repo.GetOrganizationByID(input.OrgID)
	if err != nil {
		return nil, errors.New("organization not found")
	}

	event := &models.Event{
		OrgID:                org.ID,
		Name:                 input.Name,
		Slug:                 slug,
		Description:          input.Description,
		StartTime:            input.StartTime,
		EndTime:              input.EndTime,
		FreezeTime:           input.FreezeTime,
		IsActive:             false,
		IsPaused:             false,
		ScoreboardVisibility: visibility,
		FrontendURL:          input.FrontendURL,
		CustomMongoURI:       input.CustomMongoURI,
		S3Config:             input.S3Config,
		EventTokenHash:       string(hash),
		EventTokenPrefix:     prefix,
	}

	if err := s.repo.CreateEvent(event); err != nil {
		return nil, err
	}

	return &CreateEventResult{Event: event, EventToken: eventToken}, nil
}

func (s *OrganizationService) GetEventByID(id string) (*models.Event, error) {
	return s.repo.GetEventByID(id)
}

func (s *OrganizationService) ListEventsByOrg(orgID string) ([]models.Event, error) {
	return s.repo.ListEventsByOrg(orgID)
}

func (s *OrganizationService) UpdateEvent(id string, update *models.Event) error {
	return s.repo.UpdateEvent(id, update)
}

// ValidateEventToken verifies a plain-text event token and returns the matching event.
func (s *OrganizationService) ValidateEventToken(token string) (*models.Event, error) {
	if len(token) < 8 {
		return nil, errors.New("invalid event token")
	}
	prefix := token[:8]
	event, err := s.repo.GetEventByTokenPrefix(prefix)
	if err != nil {
		return nil, errors.New("invalid event token")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(event.EventTokenHash), []byte(token)); err != nil {
		return nil, errors.New("invalid event token")
	}
	return event, nil
}
