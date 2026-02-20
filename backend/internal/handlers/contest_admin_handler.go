package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/utils"
)

type ContestAdminHandler struct {
	contestAdminService *services.ContestAdminService
}

func NewContestAdminHandler(contestAdminService *services.ContestAdminService) *ContestAdminHandler {
	return &ContestAdminHandler{
		contestAdminService: contestAdminService,
	}
}

// ListContests returns all contests
// @Summary List all contests
// @Tags Admin Contest
// @Produce json
// @Success 200 {array} models.Contest
// @Security ApiKeyAuth
// @Router /admin/contests [get]
func (h *ContestAdminHandler) ListContests(c *gin.Context) {
	contests, err := h.contestAdminService.ListContests()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}
	c.JSON(http.StatusOK, contests)
}

// GetContest returns a contest by ID
// @Summary Get contest
// @Tags Admin Contest
// @Param id path string true "Contest ID"
// @Produce json
// @Success 200 {object} models.Contest
// @Security ApiKeyAuth
// @Router /admin/contests/{id} [get]
func (h *ContestAdminHandler) GetContest(c *gin.Context) {
	id := c.Param("id")
	contest, err := h.contestAdminService.GetContest(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contest not found"})
		return
	}
	c.JSON(http.StatusOK, contest)
}

// CreateContestRequest represents create contest request
type CreateContestRequest struct {
	Name                 string `json:"name" binding:"required"`
	Description          string `json:"description"`
	StartTime            string `json:"start_time" binding:"required"`
	EndTime              string `json:"end_time" binding:"required"`
	FreezeTime           string `json:"freeze_time"`
	ScoreboardVisibility string `json:"scoreboard_visibility"`
}

// CreateContest creates a new contest
// @Summary Create contest
// @Tags Admin Contest
// @Accept json
// @Produce json
// @Param request body CreateContestRequest true "Contest details"
// @Success 201 {object} models.Contest
// @Security ApiKeyAuth
// @Router /admin/contests [post]
func (h *ContestAdminHandler) CreateContest(c *gin.Context) {
	var req CreateContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format, use RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format, use RFC3339"})
		return
	}

	var freezeTime *time.Time
	if req.FreezeTime != "" {
		ft, err := time.Parse(time.RFC3339, req.FreezeTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid freeze_time format, use RFC3339"})
			return
		}
		freezeTime = &ft
	}

	contest, err := h.contestAdminService.CreateContest(req.Name, req.Description, startTime, endTime, freezeTime, req.ScoreboardVisibility)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	c.JSON(http.StatusCreated, contest)
}

// UpdateContestEntityRequest represents update contest entity request
type UpdateContestEntityRequest struct {
	Name                 string `json:"name" binding:"required"`
	Description          string `json:"description"`
	StartTime            string `json:"start_time" binding:"required"`
	EndTime              string `json:"end_time" binding:"required"`
	IsActive             bool   `json:"is_active"`
	FreezeTime           string `json:"freeze_time"`
	ScoreboardVisibility string `json:"scoreboard_visibility"`
}

// UpdateContest updates a contest
// @Summary Update contest
// @Tags Admin Contest
// @Accept json
// @Produce json
// @Param id path string true "Contest ID"
// @Param request body UpdateContestRequest true "Contest details"
// @Success 200 {object} models.Contest
// @Security ApiKeyAuth
// @Router /admin/contests/{id} [put]
func (h *ContestAdminHandler) UpdateContest(c *gin.Context) {
	id := c.Param("id")
	var req UpdateContestEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format, use RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format, use RFC3339"})
		return
	}

	var freezeTime *time.Time
	if req.FreezeTime != "" {
		ft, err := time.Parse(time.RFC3339, req.FreezeTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid freeze_time format, use RFC3339"})
			return
		}
		freezeTime = &ft
	}

	contest, err := h.contestAdminService.UpdateContest(id, req.Name, req.Description, startTime, endTime, req.IsActive, freezeTime, req.ScoreboardVisibility)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	c.JSON(http.StatusOK, contest)
}

// DeleteContest deletes a contest
// @Summary Delete contest
// @Tags Admin Contest
// @Param id path string true "Contest ID"
// @Success 200 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/contests/{id} [delete]
func (h *ContestAdminHandler) DeleteContest(c *gin.Context) {
	id := c.Param("id")
	if err := h.contestAdminService.DeleteContest(id); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Contest deleted"})
}

// SetActiveContestRequest represents set active contest request
type SetActiveContestRequest struct {
	ContestID string `json:"contest_id" binding:"required"`
}

// SetActiveContest sets which contest is active
// @Summary Set active contest
// @Tags Admin Contest
// @Accept json
// @Produce json
// @Param request body SetActiveContestRequest true "Contest ID"
// @Success 200 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/contests/set-active [post]
func (h *ContestAdminHandler) SetActiveContest(c *gin.Context) {
	var req SetActiveContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	if err := h.contestAdminService.SetActiveContest(req.ContestID); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Active contest updated"})
}

// ListRounds returns rounds for a contest
// @Summary List rounds
// @Tags Admin Contest Rounds
// @Param contestId path string true "Contest ID"
// @Produce json
// @Success 200 {array} models.ContestRound
// @Security ApiKeyAuth
// @Router /admin/contests/{contestId}/rounds [get]
func (h *ContestAdminHandler) ListRounds(c *gin.Context) {
	contestID := c.Param("id")
	rounds, err := h.contestAdminService.ListRounds(contestID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}
	c.JSON(http.StatusOK, rounds)
}

// CreateRoundRequest represents create round request
type CreateRoundRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	VisibleFrom string `json:"visible_from" binding:"required"`
	StartTime   string `json:"start_time" binding:"required"`
	EndTime     string `json:"end_time" binding:"required"`
}

// CreateRound creates a round
// @Summary Create round
// @Tags Admin Contest Rounds
// @Accept json
// @Produce json
// @Param contestId path string true "Contest ID"
// @Param request body CreateRoundRequest true "Round details"
// @Success 201 {object} models.ContestRound
// @Security ApiKeyAuth
// @Router /admin/contests/{contestId}/rounds [post]
func (h *ContestAdminHandler) CreateRound(c *gin.Context) {
	contestID := c.Param("id")
	var req CreateRoundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	visibleFrom, err := time.Parse(time.RFC3339, req.VisibleFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid visible_from format, use RFC3339"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format, use RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format, use RFC3339"})
		return
	}

	round, err := h.contestAdminService.CreateRound(contestID, req.Name, req.Description, req.Order, visibleFrom, startTime, endTime)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	c.JSON(http.StatusCreated, round)
}

// UpdateRoundRequest represents update round request
type UpdateRoundRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	VisibleFrom string `json:"visible_from" binding:"required"`
	StartTime   string `json:"start_time" binding:"required"`
	EndTime     string `json:"end_time" binding:"required"`
}

// UpdateRound updates a round
// @Summary Update round
// @Tags Admin Contest Rounds
// @Accept json
// @Produce json
// @Param contestId path string true "Contest ID"
// @Param roundId path string true "Round ID"
// @Param request body UpdateRoundRequest true "Round details"
// @Success 200 {object} models.ContestRound
// @Security ApiKeyAuth
// @Router /admin/contests/{contestId}/rounds/{roundId} [put]
func (h *ContestAdminHandler) UpdateRound(c *gin.Context) {
	roundID := c.Param("roundId")
	var req UpdateRoundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	visibleFrom, err := time.Parse(time.RFC3339, req.VisibleFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid visible_from format, use RFC3339"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format, use RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format, use RFC3339"})
		return
	}

	round, err := h.contestAdminService.UpdateRound(roundID, req.Name, req.Description, req.Order, visibleFrom, startTime, endTime)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	c.JSON(http.StatusOK, round)
}

// DeleteRound deletes a round
// @Summary Delete round
// @Tags Admin Contest Rounds
// @Param contestId path string true "Contest ID"
// @Param roundId path string true "Round ID"
// @Success 200 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/contests/{contestId}/rounds/{roundId} [delete]
func (h *ContestAdminHandler) DeleteRound(c *gin.Context) {
	roundID := c.Param("roundId")
	if err := h.contestAdminService.DeleteRound(roundID); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Round deleted"})
}

// AttachChallengesRequest represents attach challenges request
type AttachChallengesRequest struct {
	ChallengeIDs []string `json:"challenge_ids" binding:"required"`
}

// AttachChallenges attaches challenges to a round
// @Summary Attach challenges to round
// @Tags Admin Contest Rounds
// @Accept json
// @Produce json
// @Param contestId path string true "Contest ID"
// @Param roundId path string true "Round ID"
// @Param request body AttachChallengesRequest true "Challenge IDs"
// @Success 200 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/contests/{contestId}/rounds/{roundId}/challenges [post]
func (h *ContestAdminHandler) AttachChallenges(c *gin.Context) {
	roundID := c.Param("roundId")
	var req AttachChallengesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	if err := h.contestAdminService.AttachChallengesToRound(roundID, req.ChallengeIDs); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Challenges attached"})
}

// DetachChallengesRequest represents detach challenges request
type DetachChallengesRequest struct {
	ChallengeIDs []string `json:"challenge_ids" binding:"required"`
}

// DetachChallenges detaches challenges from a round
// @Summary Detach challenges from round
// @Tags Admin Contest Rounds
// @Accept json
// @Produce json
// @Param contestId path string true "Contest ID"
// @Param roundId path string true "Round ID"
// @Param request body DetachChallengesRequest true "Challenge IDs"
// @Success 200 {object} map[string]string
// @Security ApiKeyAuth
// @Router /admin/contests/{contestId}/rounds/{roundId}/challenges [delete]
func (h *ContestAdminHandler) DetachChallenges(c *gin.Context) {
	roundID := c.Param("roundId")
	var req DetachChallengesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	if err := h.contestAdminService.DetachChallengesFromRound(roundID, req.ChallengeIDs); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Challenges detached"})
}

// GetRoundChallenges returns challenge IDs for a round
// @Summary Get round challenges
// @Tags Admin Contest Rounds
// @Param contestId path string true "Contest ID"
// @Param roundId path string true "Round ID"
// @Produce json
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Router /admin/contests/{contestId}/rounds/{roundId}/challenges [get]
func (h *ContestAdminHandler) GetRoundChallenges(c *gin.Context) {
	roundID := c.Param("roundId")
	ids, err := h.contestAdminService.GetChallengesForRound(roundID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	strIDs := make([]string, len(ids))
	for i, id := range ids {
		strIDs[i] = id.Hex()
	}
	c.JSON(http.StatusOK, strIDs)
}
