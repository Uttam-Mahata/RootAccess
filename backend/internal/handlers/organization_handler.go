package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/models"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
)

// OrganizationHandler manages HTTP endpoints for the SaaS multi-tenant layer.
type OrganizationHandler struct {
	orgService *services.OrganizationService
}

func NewOrganizationHandler(orgService *services.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{orgService: orgService}
}

// --- request / response types ---

type createOrgRequest struct {
	Name       string `json:"name" binding:"required,min=2,max=100"`
	Slug       string `json:"slug"`
	OwnerEmail string `json:"owner_email" binding:"required,email"`
	OwnerName  string `json:"owner_name" binding:"required,min=2,max=100"`
}

type createEventRequest struct {
	Name                 string          `json:"name" binding:"required,min=2,max=100"`
	Slug                 string          `json:"slug"`
	Description          string          `json:"description"`
	StartTime            time.Time       `json:"start_time" binding:"required"`
	EndTime              time.Time       `json:"end_time" binding:"required"`
	FreezeTime           *time.Time      `json:"freeze_time"`
	ScoreboardVisibility string          `json:"scoreboard_visibility"`
	FrontendURL          string          `json:"frontend_url"`
	CustomMongoURI       string          `json:"custom_mongo_uri"`
	S3Config             *models.S3Config `json:"s3_config"`
}

type updateEventRequest struct {
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	StartTime            time.Time       `json:"start_time"`
	EndTime              time.Time       `json:"end_time"`
	FreezeTime           *time.Time      `json:"freeze_time"`
	IsActive             bool            `json:"is_active"`
	IsPaused             bool            `json:"is_paused"`
	ScoreboardVisibility string          `json:"scoreboard_visibility"`
	FrontendURL          string          `json:"frontend_url"`
	CustomMongoURI       string          `json:"custom_mongo_uri"`
	S3Config             *models.S3Config `json:"s3_config"`
}

// --- handlers ---

// CreateOrganization registers a new organization and returns a one-time API key.
// @Summary Register a new organization
// @Description Register an organization on the RootAccess SaaS platform and receive an API key.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param request body createOrgRequest true "Organization details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /orgs [post]
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var req createOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.orgService.CreateOrganization(services.CreateOrgInput{
		Name:       req.Name,
		Slug:       req.Slug,
		OwnerEmail: req.OwnerEmail,
		OwnerName:  req.OwnerName,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Organization created successfully. Store your API key — it will not be shown again.",
		"org":        result.Org,
		"api_key":    result.APIKey,
	})
}

// GetOrganization returns public details of an organization.
// @Summary Get organization details
// @Description Retrieve public details for an organization by ID.
// @Tags Organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} models.Organization
// @Failure 404 {object} map[string]string
// @Router /orgs/{id} [get]
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	org, err := h.orgService.GetOrganizationByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}
	c.JSON(http.StatusOK, org)
}

// CreateEvent creates a new CTF event under an organization.
// Requires X-API-Key header matching the organization's API key.
// @Summary Create a CTF event
// @Description Create a new CTF event for an organization. Returns a one-time event token.
// @Tags Events
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param request body createEventRequest true "Event details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /orgs/{id}/events [post]
func (h *OrganizationHandler) CreateEvent(c *gin.Context) {
	orgID := c.Param("id")

	// Verify the caller owns this organization
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "X-API-Key header required"})
		return
	}
	org, err := h.orgService.ValidateAPIKey(apiKey)
	if err != nil || org.ID.Hex() != orgID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or mismatched API key"})
		return
	}

	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !req.EndTime.After(req.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_time must be after start_time"})
		return
	}

	result, err := h.orgService.CreateEvent(services.CreateEventInput{
		OrgID:                orgID,
		Name:                 req.Name,
		Slug:                 req.Slug,
		Description:          req.Description,
		StartTime:            req.StartTime,
		EndTime:              req.EndTime,
		FreezeTime:           req.FreezeTime,
		ScoreboardVisibility: req.ScoreboardVisibility,
		FrontendURL:          req.FrontendURL,
		CustomMongoURI:       req.CustomMongoURI,
		S3Config:             req.S3Config,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Event created successfully. Store your event token — it will not be shown again.",
		"event":       result.Event,
		"event_token": result.EventToken,
	})
}

// ListEvents returns all events for an organization.
// @Summary List events for an organization
// @Description List all CTF events belonging to an organization.
// @Tags Events
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {array} models.Event
// @Failure 400 {object} map[string]string
// @Router /orgs/{id}/events [get]
func (h *OrganizationHandler) ListEvents(c *gin.Context) {
	events, err := h.orgService.ListEventsByOrg(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

// GetEvent returns details of a single event.
// @Summary Get event details
// @Description Retrieve details for a CTF event by ID.
// @Tags Events
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} models.Event
// @Failure 404 {object} map[string]string
// @Router /events/{id} [get]
func (h *OrganizationHandler) GetEvent(c *gin.Context) {
	event, err := h.orgService.GetEventByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}

// UpdateEvent updates a CTF event. Requires X-API-Key header.
// @Summary Update a CTF event
// @Description Update settings for an existing CTF event.
// @Tags Events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param request body updateEventRequest true "Event update fields"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /events/{id} [put]
func (h *OrganizationHandler) UpdateEvent(c *gin.Context) {
	eventID := c.Param("id")

	// Fetch the event first to check org ownership
	event, err := h.orgService.GetEventByID(eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	// Verify ownership via API key
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "X-API-Key header required"})
		return
	}
	org, err := h.orgService.ValidateAPIKey(apiKey)
	if err != nil || org.ID != event.OrgID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or mismatched API key"})
		return
	}

	var req updateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !req.StartTime.IsZero() && !req.EndTime.IsZero() && !req.EndTime.After(req.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_time must be after start_time"})
		return
	}

	updated := &models.Event{
		Name:                 req.Name,
		Description:          req.Description,
		StartTime:            req.StartTime,
		EndTime:              req.EndTime,
		FreezeTime:           req.FreezeTime,
		IsActive:             req.IsActive,
		IsPaused:             req.IsPaused,
		ScoreboardVisibility: req.ScoreboardVisibility,
		FrontendURL:          req.FrontendURL,
		CustomMongoURI:       req.CustomMongoURI,
		S3Config:             req.S3Config,
	}

	if err := h.orgService.UpdateEvent(eventID, updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event updated successfully"})
}
