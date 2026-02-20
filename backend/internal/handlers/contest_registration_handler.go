package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
)

type ContestRegistrationHandler struct {
	registrationService *services.ContestRegistrationService
	teamService         *services.TeamService
}

func NewContestRegistrationHandler(
	registrationService *services.ContestRegistrationService,
	teamService *services.TeamService,
) *ContestRegistrationHandler {
	return &ContestRegistrationHandler{
		registrationService: registrationService,
		teamService:         teamService,
	}
}

// GetUpcomingContests returns all upcoming contests (public endpoint)
// @Summary Get upcoming contests
// @Description Returns all contests that haven't started yet
// @Tags contests
// @Produce json
// @Success 200 {array} models.Contest
// @Router /contests/upcoming [get]
func (h *ContestRegistrationHandler) GetUpcomingContests(c *gin.Context) {
	contests, err := h.registrationService.GetUpcomingContests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contests"})
		return
	}
	c.JSON(http.StatusOK, contests)
}

// RegisterTeamForContest registers the current user's team for a contest
// @Summary Register team for contest
// @Description Registers the authenticated user's team for a contest
// @Tags contests
// @Accept json
// @Produce json
// @Param contest_id path string true "Contest ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /contests/{contest_id}/register [post]
func (h *ContestRegistrationHandler) RegisterTeamForContest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	contestID := c.Param("contest_id")
	if contestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contest ID is required"})
		return
	}

	// Get user's team
	team, err := h.teamService.GetUserTeam(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You must be part of a team to register"})
		return
	}

	err = h.registrationService.RegisterTeamForContest(team.ID.Hex(), contestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team registered successfully"})
}

// UnregisterTeamFromContest unregisters the current user's team from a contest
// @Summary Unregister team from contest
// @Description Unregisters the authenticated user's team from a contest
// @Tags contests
// @Accept json
// @Produce json
// @Param contest_id path string true "Contest ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /contests/{contest_id}/unregister [post]
func (h *ContestRegistrationHandler) UnregisterTeamFromContest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	contestID := c.Param("contest_id")
	if contestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contest ID is required"})
		return
	}

	// Get user's team
	team, err := h.teamService.GetUserTeam(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You must be part of a team to unregister"})
		return
	}

	err = h.registrationService.UnregisterTeamFromContest(team.ID.Hex(), contestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team unregistered successfully"})
}

// GetRegisteredTeamsCount returns how many teams are registered for a contest
// @Summary Get registered teams count
// @Description Returns the number of teams registered for a contest
// @Tags contests
// @Produce json
// @Param contest_id path string true "Contest ID"
// @Success 200 {object} map[string]int64
// @Router /contests/{contest_id}/registered-count [get]
func (h *ContestRegistrationHandler) GetRegisteredTeamsCount(c *gin.Context) {
	contestID := c.Param("contest_id")
	if contestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contest ID is required"})
		return
	}

	count, err := h.registrationService.GetRegisteredTeamsCount(contestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// GetTeamRegistrationStatus checks if the current user's team is registered for a contest
// @Summary Get team registration status
// @Description Checks if the authenticated user's team is registered for a contest
// @Tags contests
// @Produce json
// @Param contest_id path string true "Contest ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /contests/{contest_id}/registration-status [get]
func (h *ContestRegistrationHandler) GetTeamRegistrationStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	contestID := c.Param("contest_id")
	if contestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contest ID is required"})
		return
	}

	// Get user's team
	team, err := h.teamService.GetUserTeam(userID.(string))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"registered": false})
		return
	}

	registered, err := h.registrationService.IsTeamRegistered(team.ID.Hex(), contestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check registration status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"registered": registered})
}
