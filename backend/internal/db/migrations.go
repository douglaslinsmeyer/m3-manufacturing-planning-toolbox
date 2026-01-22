package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations executes all pending SQL migrations
func RunMigrations(db *sql.DB, migrationsPath string) error {
	// Create migrations tracking table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Read migration files from directory
	migrationFiles, err := getMigrationFiles(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// Apply pending migrations
	for _, file := range migrationFiles {
		// Only process .up.sql files
		if !strings.HasSuffix(file, ".up.sql") {
			continue
		}

		// Check if already applied
		if _, applied := appliedMigrations[file]; applied {
			log.Printf("Migration %s already applied, skipping", file)
			continue
		}

		// Read migration SQL
		migrationPath := filepath.Join(migrationsPath, file)
		sqlContent, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file, err)
		}

		// Execute migration in a transaction
		log.Printf("Applying migration: %s", file)
		if err := applyMigration(db, file, string(sqlContent)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", file, err)
		}

		log.Printf("Successfully applied migration: %s", file)
	}

	log.Println("All migrations completed successfully")
	return nil
}

// createMigrationsTable creates the table to track applied migrations
func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`
	_, err := db.Exec(query)
	return err
}

// getAppliedMigrations returns a map of already applied migration files
func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// getMigrationFiles returns a sorted list of migration files
func getMigrationFiles(migrationsPath string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.sql"))
	if err != nil {
		return nil, err
	}

	// Extract just the filenames
	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, filepath.Base(file))
	}

	// Sort to ensure migrations run in order
	sort.Strings(fileNames)

	return fileNames, nil
}

// applyMigration executes a single migration within a transaction
func applyMigration(db *sql.DB, version string, sqlContent string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute the migration SQL
	if _, err := tx.Exec(sqlContent); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record that this migration has been applied
	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version) VALUES ($1)",
		version,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}
