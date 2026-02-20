package main

import (
	"log"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/config"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/database"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/routes"
)

// @title RootAccess CTF API
// @version 1.0.0
// @description This is the backend API for the RootAccess CTF Platform.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/Uttam-Mahata/RootAccess/issues
// @contact.email contact@rootaccess.live

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host rootaccess.live
// @BasePath /
// @query.collection.format multi

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	// Local development or containerized deployment
	cfg := config.LoadConfig()

	// Connect to Database
	database.ConnectDB(cfg.MongoURI, cfg.DBName)
	database.ConnectRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

	r := routes.SetupRouter(cfg)

	log.Printf("Server running on port %s", cfg.Port)
	r.Run(":" + cfg.Port)
}
