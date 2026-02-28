package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/utils"
)

type ScoreboardHandler struct {
	scoreboardService *services.ScoreboardService
	contestEntityRepo *repositories.ContestEntityRepository
	contestRepo       *repositories.ContestRepository
}

func NewScoreboardHandler(scoreboardService *services.ScoreboardService, contestEntityRepo *repositories.ContestEntityRepository, contestRepo *repositories.ContestRepository) *ScoreboardHandler {
	return &ScoreboardHandler{
		scoreboardService: scoreboardService,
		contestEntityRepo: contestEntityRepo,
		contestRepo:       contestRepo,
	}
}

// GetScoreboard returns the individual scoreboard for a contest
// @Summary Get individual scoreboard
// @Description Retrieve the current leaderboard for individual users in a contest, sorted by points.
// @Tags Scoreboard
// @Produce json
// @Param contest_id query string true "Contest ID"
// @Success 200 {array} services.UserScore
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /scoreboard [get]
func (h *ScoreboardHandler) GetScoreboard(c *gin.Context) {
	contestID := c.Query("contest_id")
	if contestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contest_id is required"})
		return
	}

	if h.contestEntityRepo != nil {
		contest, err := h.contestEntityRepo.FindByID(contestID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "contest not found"})
			return
		}
		if contest.GetScoreboardVisibility() == "hidden" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Scoreboard is currently hidden"})
			return
		}
	}

	scores, err := h.scoreboardService.GetScoreboard(contestID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, scores)
}

// GetTeamScoreboard returns the team scoreboard for a contest
// @Summary Get team scoreboard
// @Description Retrieve the current leaderboard for teams in a contest, sorted by points.
// @Tags Scoreboard
// @Produce json
// @Param contest_id query string true "Contest ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /scoreboard/teams [get]
func (h *ScoreboardHandler) GetTeamScoreboard(c *gin.Context) {
	contestID := c.Query("contest_id")
	if contestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contest_id is required"})
		return
	}

	if h.contestEntityRepo != nil {
		contest, err := h.contestEntityRepo.FindByID(contestID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "contest not found"})
			return
		}
		if contest.GetScoreboardVisibility() == "hidden" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Scoreboard is currently hidden"})
			return
		}
	}

	scores, err := h.scoreboardService.GetTeamScoreboard(contestID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"teams": scores,
	})
}

// GetTeamStatistics returns team score progression statistics for a contest
// @Summary Get team statistics
// @Description Retrieve team score progression over time for charts
// @Tags Scoreboard
// @Produce json
// @Param contest_id query string true "Contest ID"
// @Param days query int false "Number of days (default: 30, max: 90)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /scoreboard/teams/statistics [get]
func (h *ScoreboardHandler) GetTeamStatistics(c *gin.Context) {
	contestID := c.Query("contest_id")
	if contestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contest_id is required"})
		return
	}

	if h.contestEntityRepo != nil {
		contest, err := h.contestEntityRepo.FindByID(contestID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "contest not found"})
			return
		}
		if contest.GetScoreboardVisibility() == "hidden" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Scoreboard is currently hidden"})
			return
		}
	}

	days := 30
	if daysParam := c.Query("days"); daysParam != "" {
		if parsedDays := parseInt(daysParam); parsedDays > 0 {
			days = parsedDays
		}
	}

	stats, err := h.scoreboardService.GetTeamScoreProgression(days, contestID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"progressions": stats,
	})
}

// GetScoreboardContests returns contests that should appear on the scoreboard
// @Summary Get scoreboard contests
// @Description Retrieve contests eligible for scoreboard display (running + ended)
// @Tags Scoreboard
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /contests/active [get]
func (h *ScoreboardHandler) GetScoreboardContests(c *gin.Context) {
	if h.contestEntityRepo == nil {
		c.JSON(http.StatusOK, gin.H{"contests": []interface{}{}})
		return
	}

	contests, err := h.contestEntityRepo.GetScoreboardContests()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	// Also include the contest referenced by the active ContestConfig, even if its
	// is_active flag was not set (e.g. activated before the is_active fix was deployed).
	if h.contestRepo != nil {
		if cfg, err := h.contestRepo.GetActiveContest(); err == nil && cfg != nil && !cfg.ContestID.IsZero() {
			alreadyIncluded := false
			for _, c := range contests {
				if c.ID == cfg.ContestID {
					alreadyIncluded = true
					break
				}
			}
			if !alreadyIncluded {
				if activeContest, err := h.contestEntityRepo.FindByID(cfg.ContestID.Hex()); err == nil && activeContest != nil {
					contests = append(contests, *activeContest)
				}
			}
		}
	}

	now := time.Now()
	result := make([]gin.H, 0, len(contests))
	for _, contest := range contests {
		if contest.GetScoreboardVisibility() == "hidden" {
			continue
		}
		result = append(result, gin.H{
			"id":                    contest.ID.Hex(),
			"name":                  contest.Name,
			"description":           contest.Description,
			"start_time":            contest.StartTime,
			"end_time":              contest.EndTime,
			"status":                contest.Status(now),
			"scoreboard_visibility": contest.GetScoreboardVisibility(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"contests": result})
}

func parseInt(s string) int {
	var result int
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		} else {
			return 0
		}
	}
	return result
}
