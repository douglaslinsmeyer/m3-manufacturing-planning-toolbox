package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pinggolf/m3-planning-tools/internal/db"
	"github.com/pinggolf/m3-planning-tools/internal/m3api"
)

// ContextCacheWorker handles background refreshing of M3 context cache
type ContextCacheWorker struct {
	db          *db.Queries
	m3ClientTRN *m3api.Client
	m3ClientPRD *m3api.Client
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// NewContextCacheWorker creates a new context cache worker
func NewContextCacheWorker(queries *db.Queries, m3ClientTRN, m3ClientPRD *m3api.Client) *ContextCacheWorker {
	return &ContextCacheWorker{
		db:          queries,
		m3ClientTRN: m3ClientTRN,
		m3ClientPRD: m3ClientPRD,
		stopChan:    make(chan struct{}),
	}
}

// Start begins the background cache refresh worker
func (w *ContextCacheWorker) Start() {
	w.wg.Add(1)
	go w.run()
	fmt.Println("Context cache worker started")
}

// Stop gracefully stops the background worker
func (w *ContextCacheWorker) Stop() {
	close(w.stopChan)
	w.wg.Wait()
	fmt.Println("Context cache worker stopped")
}

// run is the main worker loop
func (w *ContextCacheWorker) run() {
	defer w.wg.Done()

	// Refresh immediately on startup
	w.refreshCache()

	// Then refresh every 6 hours
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.refreshCache()
		case <-w.stopChan:
			return
		}
	}
}

// refreshCache refreshes all M3 context data for both environments
func (w *ContextCacheWorker) refreshCache() {
	fmt.Println("Starting M3 context cache refresh...")
	start := time.Now()

	ctx := context.Background()

	// Refresh for both environments in parallel
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := w.refreshEnvironmentCache(ctx, "TRN", w.m3ClientTRN); err != nil {
			fmt.Printf("Error refreshing TRN cache: %v\n", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := w.refreshEnvironmentCache(ctx, "PRD", w.m3ClientPRD); err != nil {
			fmt.Printf("Error refreshing PRD cache: %v\n", err)
		}
	}()

	wg.Wait()

	duration := time.Since(start)
	fmt.Printf("M3 context cache refresh completed in %v\n", duration)
}

// refreshEnvironmentCache refreshes cache for a specific environment
func (w *ContextCacheWorker) refreshEnvironmentCache(ctx context.Context, environment string, m3Client *m3api.Client) error {
	repo := NewContextRepository(w.db, m3Client, environment)

	fmt.Printf("Refreshing %s context cache...\n", environment)

	// 1. Refresh companies (single call - only ~5-10 companies)
	companies, err := repo.GetCompanies(ctx, true) // forceRefresh=true
	if err != nil {
		return fmt.Errorf("failed to refresh companies for %s: %w", environment, err)
	}
	fmt.Printf("  %s: Cached %d companies\n", environment, len(companies))

	// 2. Refresh facilities (single call - not company-scoped)
	facilities, err := repo.GetFacilities(ctx, true)
	if err != nil {
		return fmt.Errorf("failed to refresh facilities for %s: %w", environment, err)
	}
	fmt.Printf("  %s: Cached %d facilities\n", environment, len(facilities))

	// 3. NEW: Single bulk call for all company-scoped entities
	err = repo.RefreshAllContextBulk(ctx, companies)
	if err != nil {
		// Log error - RefreshAllContextBulk handles partial failures internally
		return fmt.Errorf("bulk context refresh failed for %s: %w", environment, err)
	}

	return nil
}

// PrimeCache forces an immediate cache refresh (called after login)
func (w *ContextCacheWorker) PrimeCache(environment string) {
	fmt.Printf("Priming cache for %s environment...\n", environment)

	ctx := context.Background()

	var m3Client *m3api.Client
	switch environment {
	case "TRN":
		m3Client = w.m3ClientTRN
	case "PRD":
		m3Client = w.m3ClientPRD
	default:
		fmt.Printf("Unknown environment: %s\n", environment)
		return
	}

	if err := w.refreshEnvironmentCache(ctx, environment, m3Client); err != nil {
		fmt.Printf("Failed to prime cache for %s: %v\n", environment, err)
	}
}
