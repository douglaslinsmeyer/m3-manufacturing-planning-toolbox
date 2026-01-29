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
	"github.com/pinggolf/m3-planning-tools/internal/m3api"
	"github.com/pinggolf/m3-planning-tools/internal/queue"
	"github.com/pinggolf/m3-planning-tools/internal/services"
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

	// Initialize rate limiter service
	log.Println("Initializing rate limiter service...")
	rateLimiter := services.NewRateLimiterService(queries)
	log.Println("Rate limiter service initialized")

	// Create M3 API client factory (for bulk operation worker)
	// Note: Bulk operations use per-request tokens from job messages
	createM3Client := func(baseURL string, tokenGetter func() (string, error)) *m3api.Client {
		return m3api.NewClient(baseURL, tokenGetter)
	}

	// Initialize M3 API client for workers with placeholder (tokens provided per-job)
	// Use TRN as default environment for M3 client initialization
	trnEnvConfig, err := cfg.GetEnvironmentConfig("TRN")
	if err != nil {
		log.Printf("Warning: Could not get TRN environment config: %v", err)
	}

	// Placeholder token getter - workers will use job-specific tokens from NATS messages
	var m3Client *m3api.Client
	if trnEnvConfig != nil {
		m3Client = createM3Client(trnEnvConfig.APIBaseURL, func() (string, error) {
			return "", fmt.Errorf("token should be provided per-job via NATS message")
		})
	} else {
		log.Println("Warning: M3 client not initialized - bulk operations may not work")
	}

	// Start snapshot worker
	log.Println("Starting snapshot worker...")
	snapshotWorker := workers.NewSnapshotWorker(natsManager, queries, cfg)
	if err := snapshotWorker.Start(); err != nil {
		log.Fatalf("Failed to start snapshot worker: %v", err)
	}
	log.Println("Snapshot worker started")

	// Start bulk operation worker
	log.Println("Starting bulk operation worker...")
	bulkOpWorker := workers.NewBulkOperationWorker(natsManager, queries, m3Client, rateLimiter)
	if err := bulkOpWorker.Start(context.Background()); err != nil {
		log.Fatalf("Failed to start bulk operation worker: %v", err)
	}
	log.Println("Bulk operation worker started")

	// Initialize API server
	// Note: Context cache refresh is triggered after user login via API handlers
	// This uses user session tokens instead of service account tokens
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
