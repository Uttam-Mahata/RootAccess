package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/config"
	"github.com/go-ctf-platform/backend/internal/handlers"
	"github.com/go-ctf-platform/backend/internal/middleware"
	"github.com/go-ctf-platform/backend/internal/repositories"
	"github.com/go-ctf-platform/backend/internal/services"
	"github.com/golang-jwt/jwt/v5"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
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

	// Services
	emailService := services.NewEmailService(cfg)
	authService := services.NewAuthService(userRepo, emailService, cfg)
	challengeService := services.NewChallengeService(challengeRepo, submissionRepo, teamRepo)
	scoreboardService := services.NewScoreboardService(userRepo, submissionRepo, challengeRepo, teamRepo)
	teamService := services.NewTeamService(teamRepo, teamInvitationRepo, userRepo, emailService, submissionRepo, challengeRepo)
	notificationService := services.NewNotificationService(notificationRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	challengeHandler := handlers.NewChallengeHandler(challengeService)
	scoreboardHandler := handlers.NewScoreboardHandler(scoreboardService)
	teamHandler := handlers.NewTeamHandler(teamService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	profileHandler := handlers.NewProfileHandler(userRepo, submissionRepo, challengeRepo, teamRepo)

	// Public Routes - Authentication
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)
	r.POST("/auth/logout", authHandler.Logout)
	r.GET("/auth/verify-email", authHandler.VerifyEmail)
	r.POST("/auth/verify-email", authHandler.VerifyEmail)
	r.POST("/auth/resend-verification", authHandler.ResendVerification)
	r.POST("/auth/forgot-password", authHandler.ForgotPassword)
	r.POST("/auth/reset-password", authHandler.ResetPassword)

	// Public Routes - Scoreboard (team scoreboard)
	r.GET("/scoreboard", scoreboardHandler.GetScoreboard)
	r.GET("/scoreboard/teams", scoreboardHandler.GetTeamScoreboard)

	// Public Routes - Notifications (active notifications only)
	r.GET("/notifications", notificationHandler.GetActiveNotifications)

	// Public Routes - User Profiles
	r.GET("/users/:username/profile", profileHandler.GetUserProfile)

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

	// Protected Routes
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// User Routes
		protected.POST("/auth/change-password", authHandler.ChangePassword)
		protected.GET("/challenges", challengeHandler.GetAllChallenges)
		protected.GET("/challenges/:id", challengeHandler.GetChallengeByID)
		// Flag submission with rate limiting (5 attempts per minute per challenge)
		protected.POST("/challenges/:id/submit", middleware.RateLimitMiddleware(5, time.Minute), challengeHandler.SubmitFlag)

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
		{
			// Challenge management
			admin.GET("/challenges", challengeHandler.GetAllChallengesWithFlags)
			admin.POST("/challenges", challengeHandler.CreateChallenge)
			admin.PUT("/challenges/:id", challengeHandler.UpdateChallenge)
			admin.DELETE("/challenges/:id", challengeHandler.DeleteChallenge)

			// Notification management
			admin.GET("/notifications", notificationHandler.GetAllNotifications)
			admin.POST("/notifications", notificationHandler.CreateNotification)
			admin.PUT("/notifications/:id", notificationHandler.UpdateNotification)
			admin.DELETE("/notifications/:id", notificationHandler.DeleteNotification)
			admin.POST("/notifications/:id/toggle", notificationHandler.ToggleNotificationActive)
		}
	}

	return r
}
