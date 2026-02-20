package routes

import (
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/config"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/database"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/handlers"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/middleware"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
	websocketPkg "github.com/Uttam-Mahata/RootAccess/backend/internal/websocket"
	"github.com/golang-jwt/jwt/v5"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Configure trusted proxies for accurate client IP detection
	// When behind a reverse proxy (nginx, load balancer), this ensures
	// c.ClientIP() returns the real client IP from X-Forwarded-For / X-Real-IP
	// instead of the proxy's IP (e.g. 127.0.0.1)
	if cfg.TrustedProxies != "" {
		proxies := strings.Split(cfg.TrustedProxies, ",")
		for i := range proxies {
			proxies[i] = strings.TrimSpace(proxies[i])
		}
		r.SetTrustedProxies(proxies)
	} else {
		// Default: only trust loopback addresses in development
		r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	}
	r.ForwardedByClientIP = true
	r.RemoteIPHeaders = []string{"X-Forwarded-For", "X-Real-IP"}

	// CORS
	allowedOrigins := map[string]bool{
		cfg.FrontendURL:                true,
		"https://rootaccessctf.web.app": true,
		"https://dev.rootaccess.live":   true,
		"https://rootaccess.live":       true,
		"https://ctf.rootaccess.live":   true,
		"https://ctfapis.rootaccess.live": true,
	}

	// Add additional origins from config
	if cfg.CORSAllowedOrigins != "" {
		additionalOrigins := strings.Split(cfg.CORSAllowedOrigins, ",")
		for _, origin := range additionalOrigins {
			allowedOrigins[strings.TrimSpace(origin)] = true
		}
	}

	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			if allowedOrigins[origin] {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			} else if cfg.Environment == "development" {
				// In development, be more permissive to allow various local dev setups
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Repositories
	userRepo := repositories.NewUserRepository()
	challengeRepo := repositories.NewChallengeRepository()
	submissionRepo := repositories.NewSubmissionRepository()
	teamRepo := repositories.NewTeamRepository()
	teamInvitationRepo := repositories.NewTeamInvitationRepository()
	notificationRepo := repositories.NewNotificationRepository()
	hintRepo := repositories.NewHintRepository()
	contestRepo := repositories.NewContestRepository()
	contestEntityRepo := repositories.NewContestEntityRepository()
	contestRoundRepo := repositories.NewContestRoundRepository()
	roundChallengeRepo := repositories.NewRoundChallengeRepository()
	writeupRepo := repositories.NewWriteupRepository()
	auditLogRepo := repositories.NewAuditLogRepository()
	achievementRepo := repositories.NewAchievementRepository()
	scoreAdjustmentRepo := repositories.NewScoreAdjustmentRepository()
	teamContestRegistrationRepo := repositories.NewTeamContestRegistrationRepository()
	if err := teamContestRegistrationRepo.CreateIndexes(); err != nil {
		log.Printf("warning: could not create team_contest_registrations indexes: %v", err)
	}

	// Services
	emailService := services.NewEmailService(cfg)
	authService := services.NewAuthService(userRepo, emailService, cfg)
	oauthService := services.NewOAuthService(userRepo, cfg)
	challengeService := services.NewChallengeService(challengeRepo, submissionRepo, teamRepo)
	scoreboardService := services.NewScoreboardService(userRepo, submissionRepo, challengeRepo, teamRepo, contestRepo, scoreAdjustmentRepo, contestEntityRepo, contestRoundRepo, roundChallengeRepo, teamContestRegistrationRepo)
	teamService := services.NewTeamService(teamRepo, teamInvitationRepo, userRepo, emailService, submissionRepo, challengeRepo)
	notificationService := services.NewNotificationService(notificationRepo)
	hintService := services.NewHintService(hintRepo, challengeRepo, teamRepo)
	contestService := services.NewContestService(contestRepo)
	contestAdminService := services.NewContestAdminService(contestRepo, contestEntityRepo, contestRoundRepo, roundChallengeRepo, challengeRepo, teamContestRegistrationRepo)
	contestRegistrationService := services.NewContestRegistrationService(contestEntityRepo, teamContestRegistrationRepo, teamRepo)
	writeupService := services.NewWriteupService(writeupRepo, submissionRepo, teamRepo)
	auditLogService := services.NewAuditLogService(auditLogRepo)
	achievementService := services.NewAchievementService(achievementRepo, submissionRepo, challengeRepo)
	analyticsService := services.NewAnalyticsService(userRepo, submissionRepo, challengeRepo, teamRepo, scoreAdjustmentRepo)
	activityService := services.NewActivityService(userRepo, submissionRepo, challengeRepo, achievementRepo, teamRepo)

	// WebSocket hub
	wsHub := websocketPkg.NewHub()
	go wsHub.Run()

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	oauthHandler := handlers.NewOAuthHandler(oauthService, database.RDB, cfg)
	challengeHandler := handlers.NewChallengeHandlerWithRepos(challengeService, achievementService, contestService, contestAdminService, wsHub, submissionRepo, userRepo, teamRepo)
	scoreboardHandler := handlers.NewScoreboardHandler(scoreboardService, contestEntityRepo)
	teamHandler := handlers.NewTeamHandler(teamService)
	notificationHandler := handlers.NewNotificationHandler(notificationService, wsHub)
	profileHandler := handlers.NewProfileHandler(userRepo, submissionRepo, challengeRepo, teamRepo)
	hintHandler := handlers.NewHintHandler(hintService)
	contestHandler := handlers.NewContestHandler(contestService)
	contestAdminHandler := handlers.NewContestAdminHandler(contestAdminService)
	contestRegistrationHandler := handlers.NewContestRegistrationHandler(contestRegistrationService, teamService)
	writeupHandler := handlers.NewWriteupHandlerWithContestAdmin(writeupService, contestAdminService)
	auditLogHandler := handlers.NewAuditLogHandler(auditLogService)
	achievementHandler := handlers.NewAchievementHandler(achievementService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	activityHandler := handlers.NewActivityHandler(activityService)
	wsHandler := handlers.NewWebSocketHandler(wsHub, cfg)
	bulkChallengeHandler := handlers.NewBulkChallengeHandler(challengeService)
	leaderboardHandler := handlers.NewLeaderboardHandler(scoreboardService)
	adminUserHandler := handlers.NewAdminUserHandlerWithRepos(userRepo, teamRepo, submissionRepo, scoreAdjustmentRepo)
	adminTeamHandler := handlers.NewAdminTeamHandler(teamRepo, userRepo, submissionRepo, teamInvitationRepo, scoreAdjustmentRepo)

	// Public Routes - Authentication (with IP rate limiting)
	r.POST("/auth/register", middleware.IPRateLimitMiddleware(10, time.Minute), authHandler.Register)
	r.POST("/auth/login", middleware.IPRateLimitMiddleware(10, time.Minute), authHandler.Login)
	r.POST("/auth/logout", authHandler.Logout)
	r.GET("/auth/verify-email", authHandler.VerifyEmail)
	r.POST("/auth/verify-email", authHandler.VerifyEmail)
	r.POST("/auth/resend-verification", middleware.IPRateLimitMiddleware(5, time.Minute), authHandler.ResendVerification)
	r.POST("/auth/forgot-password", middleware.IPRateLimitMiddleware(5, time.Minute), authHandler.ForgotPassword)
	r.POST("/auth/reset-password", middleware.IPRateLimitMiddleware(5, time.Minute), authHandler.ResetPassword)

	// OAuth Routes
	r.GET("/auth/google", oauthHandler.GoogleLogin)
	r.GET("/auth/google/callback", oauthHandler.GoogleCallback)

	// GitHub OAuth Routes
	r.GET("/auth/github", oauthHandler.GitHubLogin)
	r.GET("/auth/github/callback", oauthHandler.GitHubCallback)

	// Discord OAuth Routes
	r.GET("/auth/discord", oauthHandler.DiscordLogin)
	r.GET("/auth/discord/callback", oauthHandler.DiscordCallback)

	// Public Routes - Scoreboard (contest-scoped)
	r.GET("/scoreboard", scoreboardHandler.GetScoreboard)
	r.GET("/scoreboard/teams", scoreboardHandler.GetTeamScoreboard)
	r.GET("/scoreboard/teams/statistics", scoreboardHandler.GetTeamStatistics)

	// Public Routes - Active contests for scoreboard
	r.GET("/contests/active", scoreboardHandler.GetScoreboardContests)

	// Public Routes - Notifications (active notifications only)
	r.GET("/notifications", notificationHandler.GetActiveNotifications)

	// Public Routes - Contest Status
	r.GET("/contest/status", contestHandler.GetContestStatus)

	// Public Routes - Contest Registration (upcoming contests)
	r.GET("/contests/upcoming", contestRegistrationHandler.GetUpcomingContests)
	r.GET("/contests/:contest_id/registered-count", contestRegistrationHandler.GetRegisteredTeamsCount)

	// Authenticated Routes - Contest Registration
	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware(cfg))
	{
		auth.POST("/contests/:contest_id/register", contestRegistrationHandler.RegisterTeamForContest)
		auth.POST("/contests/:contest_id/unregister", contestRegistrationHandler.UnregisterTeamFromContest)
		auth.GET("/contests/:contest_id/registration-status", contestRegistrationHandler.GetTeamRegistrationStatus)
	}

	// Public Routes - User Profiles
	r.GET("/users/:username/profile", profileHandler.GetUserProfile)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Get current user info (checks cookie)
	r.GET("/auth/me", func(c *gin.Context) {
		tokenString, err := c.Cookie("auth_token")
		if err != nil || tokenString == "" {
			c.JSON(401, gin.H{"authenticated": false})
			return
		}

		// Parse token to get user info
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"authenticated": false})
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		c.JSON(200, gin.H{
			"authenticated": true,
			"user": gin.H{
				"id":       claims["user_id"],
				"username": claims["username"],
				"email":    claims["email"],
				"role":     claims["role"],
			},
		})
	})

	// WebSocket endpoint - Moved to protected routes
	// r.GET("/ws", wsHandler.HandleWebSocket)

	// Swagger documentation (controlled via build tags)
	registerSwagger(r)

	// Enhanced leaderboard
	r.GET("/leaderboard/category", leaderboardHandler.GetCategoryLeaderboard)
	r.GET("/leaderboard/time", leaderboardHandler.GetTimeBasedLeaderboard)

	// Protected Routes
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// WebSocket endpoint
		protected.GET("/ws", wsHandler.HandleWebSocket)

		// User Routes
		protected.POST("/auth/change-password", authHandler.ChangePassword)
		protected.GET("/challenges", challengeHandler.GetAllChallenges)
		protected.GET("/challenges/:id", challengeHandler.GetChallengeByID)
		protected.GET("/challenges/:id/solves", challengeHandler.GetChallengeSolves)
		// Flag submission with rate limiting (5 attempts per minute per challenge)
		protected.POST("/challenges/:id/submit", middleware.RateLimitMiddleware(5, time.Minute), challengeHandler.SubmitFlag)

		// Hint routes
		protected.GET("/challenges/:id/hints", hintHandler.GetHints)
		protected.POST("/challenges/:id/hints/:hintId/reveal", hintHandler.RevealHint)

		// Writeup routes
		protected.POST("/challenges/:id/writeups", writeupHandler.CreateWriteup)
		protected.GET("/challenges/:id/writeups", writeupHandler.GetWriteups)
		protected.GET("/writeups/my", writeupHandler.GetMyWriteups)

		// Activity & achievements
		protected.GET("/activity/me", activityHandler.GetMyActivity)
		protected.GET("/achievements/me", achievementHandler.GetMyAchievements)

		// Writeup enhancements
		protected.PUT("/writeups/:id", writeupHandler.UpdateWriteup)
		protected.POST("/writeups/:id/upvote", writeupHandler.ToggleUpvote)

		// Team Routes
		teams := protected.Group("/teams")
		{
			// Team creation and viewing
			teams.POST("", teamHandler.CreateTeam)
			teams.GET("/my-team", teamHandler.GetMyTeam)
			teams.GET("/:id", teamHandler.GetTeamDetails)
			teams.PUT("/:id", teamHandler.UpdateTeam)
			teams.DELETE("/:id", teamHandler.DeleteTeam)

			// Invite code join
			teams.POST("/join/:code", teamHandler.JoinByCode)

			// Invitations (for invitee)
			teams.GET("/invitations", teamHandler.GetPendingInvitations)
			teams.POST("/invitations/:id/accept", teamHandler.AcceptInvitation)
			teams.POST("/invitations/:id/reject", teamHandler.RejectInvitation)

			// Team invitations (for leader)
			teams.POST("/:id/invite/username", teamHandler.InviteByUsername)
			teams.POST("/:id/invite/email", teamHandler.InviteByEmail)
			teams.GET("/:id/invitations", teamHandler.GetTeamPendingInvitations)
			teams.DELETE("/:id/invitations/:invitationId", teamHandler.CancelInvitation)

			// Member management
			teams.DELETE("/:id/members/:userId", teamHandler.RemoveMember)
			teams.POST("/:id/leave", teamHandler.LeaveTeam)

			// Invite code regeneration
			teams.POST("/:id/regenerate-code", teamHandler.RegenerateInviteCode)
		}

		// Admin Routes
		admin := protected.Group("/admin")
		admin.Use(middleware.AdminMiddleware())
		admin.Use(middleware.AuditMiddleware(auditLogService))
		{
			// Challenge management
			admin.GET("/challenges", challengeHandler.GetAllChallengesWithFlags)
			admin.POST("/challenges", challengeHandler.CreateChallenge)
			admin.PUT("/challenges/:id", challengeHandler.UpdateChallenge)
			admin.DELETE("/challenges/:id", challengeHandler.DeleteChallenge)
			admin.PUT("/challenges/:id/official-writeup", challengeHandler.UpdateOfficialWriteup)
			admin.POST("/challenges/:id/official-writeup/publish", challengeHandler.PublishOfficialWriteup)

			// Notification management
			admin.GET("/notifications", notificationHandler.GetAllNotifications)
			admin.POST("/notifications", notificationHandler.CreateNotification)
			admin.PUT("/notifications/:id", notificationHandler.UpdateNotification)
			admin.DELETE("/notifications/:id", notificationHandler.DeleteNotification)
			admin.POST("/notifications/:id/toggle", notificationHandler.ToggleNotificationActive)

			// Contest management
			admin.GET("/contest", contestHandler.GetContestConfig)
			admin.PUT("/contest", contestHandler.UpdateContestConfig)

			// Contest entities and rounds (multi-contest support)
			admin.GET("/contest-entities", contestAdminHandler.ListContests)
			admin.POST("/contest-entities", contestAdminHandler.CreateContest)
			admin.POST("/contest-entities/set-active", contestAdminHandler.SetActiveContest)
			admin.GET("/contest-entities/:id", contestAdminHandler.GetContest)
			admin.PUT("/contest-entities/:id", contestAdminHandler.UpdateContest)
			admin.DELETE("/contest-entities/:id", contestAdminHandler.DeleteContest)
			admin.GET("/contest-entities/:id/rounds", contestAdminHandler.ListRounds)
			admin.POST("/contest-entities/:id/rounds", contestAdminHandler.CreateRound)
			admin.PUT("/contest-entities/:id/rounds/:roundId", contestAdminHandler.UpdateRound)
			admin.DELETE("/contest-entities/:id/rounds/:roundId", contestAdminHandler.DeleteRound)
			admin.GET("/contest-entities/:id/rounds/:roundId/challenges", contestAdminHandler.GetRoundChallenges)
			admin.POST("/contest-entities/:id/rounds/:roundId/challenges", contestAdminHandler.AttachChallenges)
			admin.DELETE("/contest-entities/:id/rounds/:roundId/challenges", contestAdminHandler.DetachChallenges)

			// Writeup management
			admin.GET("/writeups", writeupHandler.GetAllWriteups)
			admin.PUT("/writeups/:id/status", writeupHandler.UpdateWriteupStatus)
			admin.DELETE("/writeups/:id", writeupHandler.DeleteWriteup)

			// Audit logs
			admin.GET("/audit-logs", auditLogHandler.GetAuditLogs)

			// Analytics
			admin.GET("/analytics", analyticsHandler.GetPlatformAnalytics)

			// Bulk challenge management
			admin.POST("/challenges/import", bulkChallengeHandler.ImportChallenges)
			admin.GET("/challenges/export", bulkChallengeHandler.ExportChallenges)
			admin.POST("/challenges/:id/duplicate", bulkChallengeHandler.DuplicateChallenge)

			// User management
			admin.GET("/users", adminUserHandler.ListUsers)
			admin.GET("/users/:id", adminUserHandler.GetUser)
			admin.PUT("/users/:id/status", adminUserHandler.UpdateUserStatus)
			admin.PUT("/users/:id/role", adminUserHandler.UpdateUserRole)
			admin.POST("/users/:id/score-adjust", adminUserHandler.AdjustUserScore)
			admin.DELETE("/users/:id", adminUserHandler.DeleteUser)

			// Team management (admin)
			admin.GET("/teams", adminTeamHandler.ListTeams)
			admin.GET("/teams/:id", adminTeamHandler.GetTeam)
			admin.PUT("/teams/:id", adminTeamHandler.UpdateTeam)
			admin.PUT("/teams/:id/leader", adminTeamHandler.UpdateTeamLeader)
			admin.POST("/teams/:id/score-adjust", adminTeamHandler.AdjustTeamScore)
			admin.DELETE("/teams/:id/members/:memberId", adminTeamHandler.RemoveMember)
			admin.DELETE("/teams/:id", adminTeamHandler.DeleteTeam)
		}
	}

	return r
}
