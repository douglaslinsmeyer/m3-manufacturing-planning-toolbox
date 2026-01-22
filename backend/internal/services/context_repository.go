package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pinggolf/m3-planning-tools/internal/compass"
	"github.com/pinggolf/m3-planning-tools/internal/db"
	"github.com/pinggolf/m3-planning-tools/internal/m3api"
)

const (
	cacheDuration    = 30 * 24 * time.Hour // 30 days
	refreshThreshold = 7 * 24 * time.Hour  // 7 days - refresh if older than this
)

// ContextRepository handles caching and retrieval of M3 organizational hierarchy
type ContextRepository struct {
	db         *db.Queries
	m3Client   *m3api.Client
	environment string
}

// NewContextRepository creates a new context repository
func NewContextRepository(queries *db.Queries, m3Client *m3api.Client, environment string) *ContextRepository {
	return &ContextRepository{
		db:         queries,
		m3Client:   m3Client,
		environment: environment,
	}
}

// GetCompanies returns cached companies or fetches from M3
func (r *ContextRepository) GetCompanies(ctx context.Context, forceRefresh bool) ([]compass.M3Company, error) {
	// Check cache first unless forceRefresh is true
	if !forceRefresh {
		cached, err := r.getCachedCompanies(ctx)
		if err == nil && len(cached) > 0 && r.isCacheFresh(cached[0].CachedAt) {
			return r.convertCachedCompanies(cached), nil
		}
	}

	// Fetch from M3 API
	companies, err := compass.ListCompanies(ctx, r.m3Client)
	if err != nil {
		// Try cache as fallback even if stale
		cached, cacheErr := r.getCachedCompanies(ctx)
		if cacheErr == nil && len(cached) > 0 {
			fmt.Printf("Warning: M3 API failed, using stale cache: %v\n", err)
			return r.convertCachedCompanies(cached), nil
		}
		return nil, fmt.Errorf("failed to fetch companies from M3 and no cache available: %w", err)
	}

	// Update cache
	if err := r.cacheCompanies(ctx, companies); err != nil {
		// Log error but don't fail - we have the data from M3
		fmt.Printf("Warning: Failed to update companies cache: %v\n", err)
	}

	return companies, nil
}

// GetDivisions returns cached divisions or fetches from M3
func (r *ContextRepository) GetDivisions(ctx context.Context, companyNumber string, forceRefresh bool) ([]compass.M3Division, error) {
	// Check cache first unless forceRefresh is true
	if !forceRefresh {
		cached, err := r.getCachedDivisions(ctx, companyNumber)
		if err == nil && len(cached) > 0 && r.isCacheFresh(cached[0].CachedAt) {
			return r.convertCachedDivisions(cached), nil
		}
	}

	// Fetch from M3 API
	divisions, err := compass.ListDivisions(ctx, r.m3Client, companyNumber)
	if err != nil {
		// Try cache as fallback even if stale
		cached, cacheErr := r.getCachedDivisions(ctx, companyNumber)
		if cacheErr == nil && len(cached) > 0 {
			fmt.Printf("Warning: M3 API failed, using stale cache: %v\n", err)
			return r.convertCachedDivisions(cached), nil
		}
		return nil, fmt.Errorf("failed to fetch divisions from M3 and no cache available: %w", err)
	}

	// Update cache
	if err := r.cacheDivisions(ctx, divisions); err != nil {
		fmt.Printf("Warning: Failed to update divisions cache: %v\n", err)
	}

	return divisions, nil
}

// GetFacilities returns cached facilities or fetches from M3
func (r *ContextRepository) GetFacilities(ctx context.Context, forceRefresh bool) ([]compass.M3Facility, error) {
	// Check cache first unless forceRefresh is true
	if !forceRefresh {
		cached, err := r.getCachedFacilities(ctx)
		if err == nil && len(cached) > 0 && r.isCacheFresh(cached[0].CachedAt) {
			return r.convertCachedFacilities(cached), nil
		}
	}

	// Fetch from M3 API
	facilities, err := compass.ListFacilities(ctx, r.m3Client)
	if err != nil {
		// Try cache as fallback even if stale
		cached, cacheErr := r.getCachedFacilities(ctx)
		if cacheErr == nil && len(cached) > 0 {
			fmt.Printf("Warning: M3 API failed, using stale cache: %v\n", err)
			return r.convertCachedFacilities(cached), nil
		}
		return nil, fmt.Errorf("failed to fetch facilities from M3 and no cache available: %w", err)
	}

	// Update cache
	if err := r.cacheFacilities(ctx, facilities); err != nil {
		fmt.Printf("Warning: Failed to update facilities cache: %v\n", err)
	}

	return facilities, nil
}

// GetFilteredWarehouses returns warehouses filtered by company/division/facility
func (r *ContextRepository) GetFilteredWarehouses(ctx context.Context, companyNumber string, division, facility *string) ([]compass.M3Warehouse, error) {
	// Get all warehouses for the company
	warehouses, err := r.getWarehousesForCompany(ctx, companyNumber, false)
	if err != nil {
		return nil, err
	}

	// Apply hierarchy filters
	filtered := []compass.M3Warehouse{}
	for _, wh := range warehouses {
		// Filter by division if specified
		if division != nil && *division != "" && wh.Division != *division {
			continue
		}
		// Filter by facility if specified
		if facility != nil && *facility != "" && wh.Facility != *facility {
			continue
		}
		filtered = append(filtered, wh)
	}

	return filtered, nil
}

// getWarehousesForCompany returns cached warehouses or fetches from M3
func (r *ContextRepository) getWarehousesForCompany(ctx context.Context, companyNumber string, forceRefresh bool) ([]compass.M3Warehouse, error) {
	// Check cache first unless forceRefresh is true
	if !forceRefresh {
		cached, err := r.getCachedWarehouses(ctx, companyNumber)
		if err == nil && len(cached) > 0 && r.isCacheFresh(cached[0].CachedAt) {
			return r.convertCachedWarehouses(cached), nil
		}
	}

	// Fetch from M3 API
	warehouses, err := compass.ListWarehouses(ctx, r.m3Client, companyNumber)
	if err != nil {
		// Try cache as fallback even if stale
		cached, cacheErr := r.getCachedWarehouses(ctx, companyNumber)
		if cacheErr == nil && len(cached) > 0 {
			fmt.Printf("Warning: M3 API failed, using stale cache: %v\n", err)
			return r.convertCachedWarehouses(cached), nil
		}
		return nil, fmt.Errorf("failed to fetch warehouses from M3 and no cache available: %w", err)
	}

	// Update cache
	if err := r.cacheWarehouses(ctx, warehouses); err != nil {
		fmt.Printf("Warning: Failed to update warehouses cache: %v\n", err)
	}

	return warehouses, nil
}

// isCacheFresh checks if cached data is within the refresh threshold
func (r *ContextRepository) isCacheFresh(cachedAt time.Time) bool {
	return time.Since(cachedAt) < refreshThreshold
}

// Cache storage structures
type cachedCompany struct {
	CompanyNumber string
	CompanyName   string
	Currency      string
	CachedAt      time.Time
}

type cachedDivision struct {
	CompanyNumber string
	Division      string
	DivisionName  string
	Facility      string
	Warehouse     string
	CachedAt      time.Time
}

type cachedFacility struct {
	CompanyNumber string
	Facility      string
	FacilityName  string
	Division      string
	Warehouse     string
	CachedAt      time.Time
}

type cachedWarehouse struct {
	CompanyNumber string
	Warehouse     string
	WarehouseName string
	Division      string
	Facility      string
	CachedAt      time.Time
}

// Database query methods

func (r *ContextRepository) getCachedCompanies(ctx context.Context) ([]cachedCompany, error) {
	query := `SELECT company_number, company_name, currency, cached_at
	          FROM m3_companies
	          WHERE environment = $1
	          ORDER BY company_number`

	rows, err := r.db.DB().QueryContext(ctx, query, r.environment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []cachedCompany
	for rows.Next() {
		var c cachedCompany
		if err := rows.Scan(&c.CompanyNumber, &c.CompanyName, &c.Currency, &c.CachedAt); err != nil {
			return nil, err
		}
		companies = append(companies, c)
	}

	return companies, rows.Err()
}

func (r *ContextRepository) cacheCompanies(ctx context.Context, companies []compass.M3Company) error {
	tx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Upsert companies
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO m3_companies (environment, company_number, company_name, currency, cached_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (environment, company_number)
		DO UPDATE SET
			company_name = EXCLUDED.company_name,
			currency = EXCLUDED.currency,
			cached_at = NOW()
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, company := range companies {
		_, err := stmt.ExecContext(ctx, r.environment, company.CompanyNumber, company.CompanyName, company.Currency)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ContextRepository) getCachedDivisions(ctx context.Context, companyNumber string) ([]cachedDivision, error) {
	query := `SELECT company_number, division, division_name, facility, warehouse, cached_at
	          FROM m3_divisions
	          WHERE environment = $1 AND company_number = $2
	          ORDER BY division`

	rows, err := r.db.DB().QueryContext(ctx, query, r.environment, companyNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var divisions []cachedDivision
	for rows.Next() {
		var d cachedDivision
		var facility, warehouse sql.NullString
		if err := rows.Scan(&d.CompanyNumber, &d.Division, &d.DivisionName, &facility, &warehouse, &d.CachedAt); err != nil {
			return nil, err
		}
		d.Facility = facility.String
		d.Warehouse = warehouse.String
		divisions = append(divisions, d)
	}

	return divisions, rows.Err()
}

func (r *ContextRepository) cacheDivisions(ctx context.Context, divisions []compass.M3Division) error {
	tx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO m3_divisions (environment, company_number, division, division_name, facility, warehouse, cached_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (environment, company_number, division)
		DO UPDATE SET
			division_name = EXCLUDED.division_name,
			facility = EXCLUDED.facility,
			warehouse = EXCLUDED.warehouse,
			cached_at = NOW()
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, div := range divisions {
		_, err := stmt.ExecContext(ctx, r.environment, div.CompanyNumber, div.Division, div.DivisionName, div.Facility, div.Warehouse)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ContextRepository) getCachedFacilities(ctx context.Context) ([]cachedFacility, error) {
	query := `SELECT company_number, facility, facility_name, division, warehouse, cached_at
	          FROM m3_facilities
	          WHERE environment = $1
	          ORDER BY company_number, facility`

	rows, err := r.db.DB().QueryContext(ctx, query, r.environment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var facilities []cachedFacility
	for rows.Next() {
		var f cachedFacility
		var division, warehouse sql.NullString
		if err := rows.Scan(&f.CompanyNumber, &f.Facility, &f.FacilityName, &division, &warehouse, &f.CachedAt); err != nil {
			return nil, err
		}
		f.Division = division.String
		f.Warehouse = warehouse.String
		facilities = append(facilities, f)
	}

	return facilities, rows.Err()
}

func (r *ContextRepository) cacheFacilities(ctx context.Context, facilities []compass.M3Facility) error {
	tx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO m3_facilities (environment, company_number, facility, facility_name, division, warehouse, cached_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (environment, company_number, facility)
		DO UPDATE SET
			facility_name = EXCLUDED.facility_name,
			division = EXCLUDED.division,
			warehouse = EXCLUDED.warehouse,
			cached_at = NOW()
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, fac := range facilities {
		_, err := stmt.ExecContext(ctx, r.environment, fac.CompanyNumber, fac.Facility, fac.FacilityName, fac.Division, fac.Warehouse)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ContextRepository) getCachedWarehouses(ctx context.Context, companyNumber string) ([]cachedWarehouse, error) {
	query := `SELECT company_number, warehouse, warehouse_name, division, facility, cached_at
	          FROM m3_warehouses
	          WHERE environment = $1 AND company_number = $2
	          ORDER BY warehouse`

	rows, err := r.db.DB().QueryContext(ctx, query, r.environment, companyNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var warehouses []cachedWarehouse
	for rows.Next() {
		var w cachedWarehouse
		var division, facility sql.NullString
		if err := rows.Scan(&w.CompanyNumber, &w.Warehouse, &w.WarehouseName, &division, &facility, &w.CachedAt); err != nil {
			return nil, err
		}
		w.Division = division.String
		w.Facility = facility.String
		warehouses = append(warehouses, w)
	}

	return warehouses, rows.Err()
}

func (r *ContextRepository) cacheWarehouses(ctx context.Context, warehouses []compass.M3Warehouse) error {
	tx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO m3_warehouses (environment, company_number, warehouse, warehouse_name, division, facility, cached_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (environment, company_number, warehouse)
		DO UPDATE SET
			warehouse_name = EXCLUDED.warehouse_name,
			division = EXCLUDED.division,
			facility = EXCLUDED.facility,
			cached_at = NOW()
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, wh := range warehouses {
		_, err := stmt.ExecContext(ctx, r.environment, wh.CompanyNumber, wh.Warehouse, wh.WarehouseName, wh.Division, wh.Facility)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Conversion methods

func (r *ContextRepository) convertCachedCompanies(cached []cachedCompany) []compass.M3Company {
	result := make([]compass.M3Company, len(cached))
	for i, c := range cached {
		result[i] = compass.M3Company{
			CompanyNumber: c.CompanyNumber,
			CompanyName:   c.CompanyName,
			Currency:      c.Currency,
		}
	}
	return result
}

func (r *ContextRepository) convertCachedDivisions(cached []cachedDivision) []compass.M3Division {
	result := make([]compass.M3Division, len(cached))
	for i, d := range cached {
		result[i] = compass.M3Division{
			CompanyNumber: d.CompanyNumber,
			Division:      d.Division,
			DivisionName:  d.DivisionName,
			Facility:      d.Facility,
			Warehouse:     d.Warehouse,
		}
	}
	return result
}

func (r *ContextRepository) convertCachedFacilities(cached []cachedFacility) []compass.M3Facility {
	result := make([]compass.M3Facility, len(cached))
	for i, f := range cached {
		result[i] = compass.M3Facility{
			CompanyNumber: f.CompanyNumber,
			Facility:      f.Facility,
			FacilityName:  f.FacilityName,
			Division:      f.Division,
			Warehouse:     f.Warehouse,
		}
	}
	return result
}

func (r *ContextRepository) convertCachedWarehouses(cached []cachedWarehouse) []compass.M3Warehouse {
	result := make([]compass.M3Warehouse, len(cached))
	for i, w := range cached {
		result[i] = compass.M3Warehouse{
			CompanyNumber: w.CompanyNumber,
			Warehouse:     w.Warehouse,
			WarehouseName: w.WarehouseName,
			Division:      w.Division,
			Facility:      w.Facility,
		}
	}
	return result
}
