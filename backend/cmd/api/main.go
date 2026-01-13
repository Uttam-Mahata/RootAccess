package main

import (
	"log"

	"github.com/go-ctf-platform/backend/internal/config"
	"github.com/go-ctf-platform/backend/internal/database"
	"github.com/go-ctf-platform/backend/internal/routes"
)

func main() {
	cfg := config.LoadConfig()

	// Connect to Database
	database.ConnectDB(cfg.MongoURI, cfg.DBName)

	r := routes.SetupRouter(cfg)

	log.Printf("Server running on port %s", cfg.Port)
	r.Run(":" + cfg.Port)
}
