package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/go-ctf-platform/backend/internal/config"
	"github.com/go-ctf-platform/backend/internal/handlers"
	"github.com/go-ctf-platform/backend/internal/middleware"
	"github.com/go-ctf-platform/backend/internal/repositories"
	"github.com/go-ctf-platform/backend/internal/services"
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
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

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

	// Services
	authService := services.NewAuthService(userRepo, cfg)
	challengeService := services.NewChallengeService(challengeRepo, submissionRepo)
	scoreboardService := services.NewScoreboardService(userRepo, submissionRepo, challengeRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	challengeHandler := handlers.NewChallengeHandler(challengeService)
	scoreboardHandler := handlers.NewScoreboardHandler(scoreboardService)

	// Public Routes
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.GET("/scoreboard", scoreboardHandler.GetScoreboard)

	// Protected Routes
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		protected.GET("/challenges", challengeHandler.GetAllChallenges)
		protected.GET("/challenges/:id", challengeHandler.GetChallengeByID)
		protected.POST("/challenges/:id/submit", challengeHandler.SubmitFlag)

		// Admin Routes
		admin := protected.Group("/")
		admin.Use(middleware.AdminMiddleware())
		{
			admin.POST("/challenges", challengeHandler.CreateChallenge)
		}
	}

	return r
}
