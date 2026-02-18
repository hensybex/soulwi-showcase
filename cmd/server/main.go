package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hensybex/soulwi_go_back/internal/config"
	"github.com/hensybex/soulwi_go_back/internal/di"
	"github.com/hensybex/soulwi_go_back/internal/migration"
	"github.com/hensybex/soulwi_go_back/internal/server/bootstrap"
	"github.com/joho/godotenv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("[Main] No .env file found or error loading it (is fine in production)", err)
	} else {
		log.Println("[Main] .env file loaded successfully")
	}

	log.Println("[Main] Application starting...")

	// Load configuration
	log.Println("[Main] Step 1: Loading configuration...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
	log.Println("[Main] Step 1: Configuration loaded.")

	// Database connection
	log.Println("[Main] Step 2: Connecting to database...")
	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("[Main] Step 2: Database connected.")

	// Apply all migrations
	log.Println("[Main] Step 3: Applying migrations...")
	if err = migration.ApplyMigrations(db); err != nil {
		log.Fatalf("Migration error: %v", err)
	}
	log.Println("[Main] Step 3: Migrations applied.")

	// Build DI container (repositories, usecases, handlers)
	log.Println("[Main] Step 4: Building DI container...")
	container := di.NewContainer(db, cfg)
	log.Println("[Main] Step 4: DI container built.")

	// Bootstrap server (apply migrations, setup router, run server)
	log.Println("[Main] Step 5: Bootstrapping server...")
	srv := bootstrap.NewServer(db, cfg, container)
	log.Printf("[Main] Step 5: Server bootstrapped. Server running on port %s\n", cfg.ApiPort)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Run server in goroutine
	go func() {
		if runErr := srv.Run(); runErr != nil {
			log.Printf("[Main] Server error: %v", runErr)
		}
	}()

	// Wait for shutdown signal
	<-quit
	log.Println("[Main] Shutdown signal received, starting graceful shutdown...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Graceful shutdown
	if err = srv.Shutdown(ctx); err != nil {
		log.Printf("[Main] Error during shutdown: %v", err)
	}

	// Close database connection
	log.Println("[Main] Closing database connection...")
	var sqlDB *sql.DB
	sqlDB, err = db.DB()
	if err != nil {
		log.Printf("[Main] Error getting database instance: %v", err)
	} else {
		if err = sqlDB.Close(); err != nil {
			log.Printf("[Main] Error closing database: %v", err)
		} else {
			log.Println("[Main] Database connection closed.")
		}
	}

	log.Println("[Main] Application shutdown complete.")
}
