package main

import (
	"log"

	"vehicle-management-system/config"
	"vehicle-management-system/database"
	"vehicle-management-system/routes"
	"vehicle-management-system/seed"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db := database.InitDB(&cfg.Database)
	defer db.Close()

	// Seed initial data
	seed.SeedAll()

	// Setup router
	router := routes.SetupRouter(cfg)

	// Start server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
