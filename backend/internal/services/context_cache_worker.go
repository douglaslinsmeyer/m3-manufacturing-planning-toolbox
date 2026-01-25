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

	// Refresh companies
	companies, err := repo.GetCompanies(ctx, true) // forceRefresh=true
	if err != nil {
		return fmt.Errorf("failed to refresh companies for %s: %w", environment, err)
	}
	fmt.Printf("  %s: Cached %d companies\n", environment, len(companies))

	// Refresh facilities (not tied to a specific company)
	facilities, err := repo.GetFacilities(ctx, true)
	if err != nil {
		return fmt.Errorf("failed to refresh facilities for %s: %w", environment, err)
	}
	fmt.Printf("  %s: Cached %d facilities\n", environment, len(facilities))

	// Refresh divisions, warehouses, and order types for each company
	var divCount, whCount, motCount, cotCount int
	for _, company := range companies {
		// Refresh divisions
		divisions, err := repo.GetDivisions(ctx, company.CompanyNumber, true)
		if err != nil {
			fmt.Printf("  Warning: Failed to refresh divisions for company %s in %s: %v\n", company.CompanyNumber, environment, err)
			continue
		}
		divCount += len(divisions)

		// Refresh warehouses
		warehouses, err := repo.getWarehousesForCompany(ctx, company.CompanyNumber, true)
		if err != nil {
			fmt.Printf("  Warning: Failed to refresh warehouses for company %s in %s: %v\n", company.CompanyNumber, environment, err)
			continue
		}
		whCount += len(warehouses)

		// Refresh manufacturing order types
		mfgOrderTypes, err := repo.GetManufacturingOrderTypes(ctx, company.CompanyNumber, true)
		if err != nil {
			fmt.Printf("  Warning: Failed to refresh manufacturing order types for company %s in %s: %v\n", company.CompanyNumber, environment, err)
			// Don't continue - try customer order types too
		} else {
			motCount += len(mfgOrderTypes)
		}

		// Refresh customer order types
		coOrderTypes, err := repo.GetCustomerOrderTypes(ctx, company.CompanyNumber, true)
		if err != nil {
			fmt.Printf("  Warning: Failed to refresh customer order types for company %s in %s: %v\n", company.CompanyNumber, environment, err)
			// Don't continue - still did other refreshes
		} else {
			cotCount += len(coOrderTypes)
		}
	}

	fmt.Printf("  %s: Cached %d divisions, %d warehouses, %d manufacturing order types, and %d customer order types\n", environment, divCount, whCount, motCount, cotCount)
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
