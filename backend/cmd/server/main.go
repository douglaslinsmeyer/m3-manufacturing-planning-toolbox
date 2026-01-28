package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pinggolf/m3-planning-tools/internal/api"
	"github.com/pinggolf/m3-planning-tools/internal/config"
	"github.com/pinggolf/m3-planning-tools/internal/db"
	"github.com/pinggolf/m3-planning-tools/internal/queue"
	"github.com/pinggolf/m3-planning-tools/internal/workers"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Check for migration command
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrations(cfg)
		return
	}

	// Initialize database connection
	database, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Configure connection pool
	database.SetMaxOpenConns(cfg.DatabaseMaxConnections)
	database.SetMaxIdleConns(cfg.DatabaseMaxIdleConnections)
	database.SetConnMaxLifetime(cfg.DatabaseConnectionLifetime)

	// Test database connection
	if err := database.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established")

	// Run database migrations (only if enabled)
	if cfg.RunMigrations {
		log.Println("Running database migrations...")
		if err := db.RunMigrations(database, "migrations"); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Database migrations completed successfully")
	} else {
		log.Println("Skipping migrations (RUN_MIGRATIONS=false)")
	}

	// Initialize database layer
	queries := db.New(database)

	// Initialize NATS connection
	log.Println("Connecting to NATS...")
	natsManager, err := queue.NewManager(cfg.NATSURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsManager.Close()
	log.Println("NATS connection established")

	// Start snapshot worker
	log.Println("Starting snapshot worker...")
	snapshotWorker := workers.NewSnapshotWorker(natsManager, queries, cfg)
	if err := snapshotWorker.Start(); err != nil {
		log.Fatalf("Failed to start snapshot worker: %v", err)
	}
	log.Println("Snapshot worker started")

	// Initialize API server
	server := api.NewServer(cfg, queries, natsManager, database)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.AppPort),
		Handler:      server.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %d (environment: %s)", cfg.AppPort, cfg.AppEnv)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

func runMigrations(cfg *config.Config) {
	// Open database connection
	database, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	log.Println("Running database migrations...")
	if err := db.RunMigrations(database, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")
}
