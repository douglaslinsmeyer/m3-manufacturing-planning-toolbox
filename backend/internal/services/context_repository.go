package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

// GetManufacturingOrderTypes returns cached manufacturing order types or fetches from M3
func (r *ContextRepository) GetManufacturingOrderTypes(ctx context.Context, companyNumber string, forceRefresh bool) ([]compass.M3ManufacturingOrderType, error) {
	// Check cache first unless forceRefresh is true
	if !forceRefresh {
		cached, err := r.getCachedManufacturingOrderTypes(ctx, companyNumber)
		if err == nil && len(cached) > 0 && r.isCacheFresh(cached[0].CachedAt) {
			return r.convertCachedManufacturingOrderTypes(cached), nil
		}
	}

	// Fetch from M3 API
	orderTypes, err := compass.ListManufacturingOrderTypes(ctx, r.m3Client, companyNumber)
	if err != nil {
		// Try cache as fallback even if stale
		cached, cacheErr := r.getCachedManufacturingOrderTypes(ctx, companyNumber)
		if cacheErr == nil && len(cached) > 0 {
			fmt.Printf("Warning: M3 API failed, using stale cache: %v\n", err)
			return r.convertCachedManufacturingOrderTypes(cached), nil
		}
		return nil, fmt.Errorf("failed to fetch manufacturing order types from M3 and no cache available: %w", err)
	}

	// Update cache
	if err := r.cacheManufacturingOrderTypes(ctx, orderTypes); err != nil {
		fmt.Printf("Warning: Failed to update manufacturing order types cache: %v\n", err)
	}

	return orderTypes, nil
}

// GetCustomerOrderTypes returns cached customer order types or fetches from M3
func (r *ContextRepository) GetCustomerOrderTypes(ctx context.Context, companyNumber string, forceRefresh bool) ([]compass.M3CustomerOrderType, error) {
	// Check cache first unless forceRefresh is true
	if !forceRefresh {
		cached, err := r.getCachedCustomerOrderTypes(ctx, companyNumber)
		if err == nil && len(cached) > 0 && r.isCacheFresh(cached[0].CachedAt) {
			return r.convertCachedCustomerOrderTypes(cached), nil
		}
	}

	// Fetch from M3 API
	orderTypes, err := compass.ListCustomerOrderTypes(ctx, r.m3Client, companyNumber)
	if err != nil {
		// Try cache as fallback even if stale
		cached, cacheErr := r.getCachedCustomerOrderTypes(ctx, companyNumber)
		if cacheErr == nil && len(cached) > 0 {
			fmt.Printf("Warning: M3 API failed, using stale cache: %v\n", err)
			return r.convertCachedCustomerOrderTypes(cached), nil
		}
		return nil, fmt.Errorf("failed to fetch customer order types from M3 and no cache available: %w", err)
	}

	// Update cache
	if err := r.cacheCustomerOrderTypes(ctx, orderTypes); err != nil {
		fmt.Printf("Warning: Failed to update customer order types cache: %v\n", err)
	}

	return orderTypes, nil
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

type cachedManufacturingOrderType struct {
	CompanyNumber        string
	OrderType            string
	OrderTypeDescription string
	LanguageCode         string
	CachedAt             time.Time
}

type cachedCustomerOrderType struct {
	CompanyNumber        string
	OrderType            string
	OrderTypeDescription string
	LanguageCode         string
	CachedAt             time.Time
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

func (r *ContextRepository) getCachedManufacturingOrderTypes(ctx context.Context, companyNumber string) ([]cachedManufacturingOrderType, error) {
	query := `SELECT company_number, order_type, order_type_description, language_code, cached_at
	          FROM m3_manufacturing_order_types
	          WHERE environment = $1 AND company_number = $2
	          ORDER BY order_type`

	rows, err := r.db.DB().QueryContext(ctx, query, r.environment, companyNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orderTypes []cachedManufacturingOrderType
	for rows.Next() {
		var ot cachedManufacturingOrderType
		if err := rows.Scan(&ot.CompanyNumber, &ot.OrderType, &ot.OrderTypeDescription, &ot.LanguageCode, &ot.CachedAt); err != nil {
			return nil, err
		}
		orderTypes = append(orderTypes, ot)
	}

	return orderTypes, rows.Err()
}

func (r *ContextRepository) cacheManufacturingOrderTypes(ctx context.Context, orderTypes []compass.M3ManufacturingOrderType) error {
	tx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO m3_manufacturing_order_types (environment, company_number, order_type, order_type_description, language_code, cached_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (environment, company_number, order_type, language_code)
		DO UPDATE SET
			order_type_description = EXCLUDED.order_type_description,
			cached_at = NOW()
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, ot := range orderTypes {
		langCode := ot.LanguageCode
		if langCode == "" {
			langCode = "GB" // Default to English
		}
		_, err := stmt.ExecContext(ctx, r.environment, ot.CompanyNumber, ot.OrderType, ot.OrderTypeDescription, langCode)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ContextRepository) getCachedCustomerOrderTypes(ctx context.Context, companyNumber string) ([]cachedCustomerOrderType, error) {
	query := `SELECT company_number, order_type, order_type_description, language_code, cached_at
	          FROM m3_customer_order_types
	          WHERE environment = $1 AND company_number = $2
	          ORDER BY order_type`

	rows, err := r.db.DB().QueryContext(ctx, query, r.environment, companyNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orderTypes []cachedCustomerOrderType
	for rows.Next() {
		var ot cachedCustomerOrderType
		if err := rows.Scan(&ot.CompanyNumber, &ot.OrderType, &ot.OrderTypeDescription, &ot.LanguageCode, &ot.CachedAt); err != nil {
			return nil, err
		}
		orderTypes = append(orderTypes, ot)
	}

	return orderTypes, rows.Err()
}

func (r *ContextRepository) cacheCustomerOrderTypes(ctx context.Context, orderTypes []compass.M3CustomerOrderType) error {
	tx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO m3_customer_order_types (environment, company_number, order_type, order_type_description, language_code, cached_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (environment, company_number, order_type, language_code)
		DO UPDATE SET
			order_type_description = EXCLUDED.order_type_description,
			cached_at = NOW()
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, ot := range orderTypes {
		langCode := ot.LanguageCode
		if langCode == "" {
			langCode = "GB" // Default to English
		}
		_, err := stmt.ExecContext(ctx, r.environment, ot.CompanyNumber, ot.OrderType, ot.OrderTypeDescription, langCode)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ContextRepository) convertCachedManufacturingOrderTypes(cached []cachedManufacturingOrderType) []compass.M3ManufacturingOrderType {
	result := make([]compass.M3ManufacturingOrderType, len(cached))
	for i, ot := range cached {
		result[i] = compass.M3ManufacturingOrderType{
			CompanyNumber:        ot.CompanyNumber,
			OrderType:            ot.OrderType,
			OrderTypeDescription: ot.OrderTypeDescription,
			LanguageCode:         ot.LanguageCode,
		}
	}
	return result
}

func (r *ContextRepository) convertCachedCustomerOrderTypes(cached []cachedCustomerOrderType) []compass.M3CustomerOrderType {
	result := make([]compass.M3CustomerOrderType, len(cached))
	for i, ot := range cached {
		result[i] = compass.M3CustomerOrderType{
			CompanyNumber:        ot.CompanyNumber,
			OrderType:            ot.OrderType,
			OrderTypeDescription: ot.OrderTypeDescription,
			LanguageCode:         ot.LanguageCode,
		}
	}
	return result
}

// RefreshAllContextBulk refreshes all context entities using bulk operations
// Returns error only for complete failures; partial failures are logged as warnings
func (r *ContextRepository) RefreshAllContextBulk(ctx context.Context, companies []compass.M3Company) error {
	if len(companies) == 0 {
		return nil
	}

	fmt.Printf("  %s: Building bulk request for %d companies...\n", r.environment, len(companies))

	// Build bulk request payload with all company-scoped operations
	// For each company: LstDivisions, LstWarehouses, LstOrderType (MO), LstOrderTypes (CO)
	requests := make([]m3api.BulkRequestItem, 0, len(companies)*4)

	for _, company := range companies {
		// Divisions: MNS100MI/LstDivisions
		requests = append(requests, m3api.BulkRequestItem{
			Program:     "MNS100MI",
			Transaction: "LstDivisions",
			Record: map[string]string{
				"CONO": company.CompanyNumber,
			},
		})

		// Warehouses: MMS005MI/LstWarehouses
		requests = append(requests, m3api.BulkRequestItem{
			Program:     "MMS005MI",
			Transaction: "LstWarehouses",
			Record: map[string]string{
				"CONO": company.CompanyNumber,
			},
		})

		// Manufacturing Order Types: PMS120MI/LstOrderType
		requests = append(requests, m3api.BulkRequestItem{
			Program:     "PMS120MI",
			Transaction: "LstOrderType",
			Record: map[string]string{
				"CONO": company.CompanyNumber,
			},
		})

		// Customer Order Types: OIS100MI/LstOrderTypes
		requests = append(requests, m3api.BulkRequestItem{
			Program:     "OIS100MI",
			Transaction: "LstOrderTypes",
			Record: map[string]string{
				"CONO": company.CompanyNumber,
			},
		})
	}

	fmt.Printf("  %s: Executing bulk API call with %d operations...\n", r.environment, len(requests))

	// Execute single bulk API call
	bulkResp, err := r.m3Client.ExecuteBulk(ctx, requests)
	if err != nil {
		// Check if this is a complete failure or partial success
		if bulkErr, ok := err.(*m3api.BulkOperationError); ok {
			if bulkErr.IsPartialSuccess() {
				fmt.Printf("  %s: Bulk refresh partial success - %d/%d succeeded, %d failed\n",
					r.environment, bulkErr.SuccessCount, bulkErr.TotalRequests, bulkErr.FailureCount)

				// Log individual failures
				for _, failed := range bulkErr.FailedItems {
					fmt.Printf("  WARNING: Failed transaction %s: %s\n",
						failed.Transaction, getErrorMessage(&failed))
				}

				// Continue with successful results - don't return error
			} else {
				// Complete failure - return error
				return fmt.Errorf("bulk context refresh failed completely: %w", err)
			}
		} else {
			// Network or other error - return error
			return fmt.Errorf("bulk context refresh error: %w", err)
		}
	}

	// Parse bulk response and group by entity type
	var (
		allDivisions       []compass.M3Division
		allWarehouses      []compass.M3Warehouse
		allMfgOrderTypes   []compass.M3ManufacturingOrderType
		allCustOrderTypes  []compass.M3CustomerOrderType
	)

	for _, result := range bulkResp.Results {
		if result.ErrorMessage != "" || result.NotProcessed {
			continue // Skip failed operations
		}

		// Extract company number from parameters (if present)
		companyNumber := ""
		if result.Parameters != nil {
			if cono, ok := result.Parameters["CONO"]; ok {
				companyNumber = cono
			}
		}

		// Parse records based on transaction type
		switch result.Transaction {
		case "LstDivisions":
			divisions := parseDivisions(result.Records)
			allDivisions = append(allDivisions, divisions...)

		case "LstWarehouses":
			warehouses := parseWarehouses(result.Records)
			allWarehouses = append(allWarehouses, warehouses...)

		case "LstOrderType":
			orderTypes := parseManufacturingOrderTypes(result.Records, companyNumber)
			allMfgOrderTypes = append(allMfgOrderTypes, orderTypes...)

		case "LstOrderTypes":
			orderTypes := parseCustomerOrderTypes(result.Records, companyNumber)
			allCustOrderTypes = append(allCustOrderTypes, orderTypes...)
		}
	}

	// Update cache in single database transaction (atomic)
	fmt.Printf("  %s: Updating cache with %d divisions, %d warehouses, %d mfg order types, %d cust order types...\n",
		r.environment, len(allDivisions), len(allWarehouses), len(allMfgOrderTypes), len(allCustOrderTypes))

	// Cache all entities
	if len(allDivisions) > 0 {
		if err := r.cacheDivisions(ctx, allDivisions); err != nil {
			fmt.Printf("  WARNING: Failed to cache divisions: %v\n", err)
		}
	}

	if len(allWarehouses) > 0 {
		if err := r.cacheWarehouses(ctx, allWarehouses); err != nil {
			fmt.Printf("  WARNING: Failed to cache warehouses: %v\n", err)
		}
	}

	if len(allMfgOrderTypes) > 0 {
		if err := r.cacheManufacturingOrderTypes(ctx, allMfgOrderTypes); err != nil {
			fmt.Printf("  WARNING: Failed to cache manufacturing order types: %v\n", err)
		}
	}

	if len(allCustOrderTypes) > 0 {
		if err := r.cacheCustomerOrderTypes(ctx, allCustOrderTypes); err != nil {
			fmt.Printf("  WARNING: Failed to cache customer order types: %v\n", err)
		}
	}

	return nil
}

// Helper functions to parse bulk response records

func parseDivisions(records []map[string]interface{}) []compass.M3Division {
	divisions := make([]compass.M3Division, 0, len(records))
	for _, record := range records {
		division := compass.M3Division{}

		if val, ok := record["CONO"].(string); ok {
			division.CompanyNumber = strings.TrimSpace(val)
		}
		if val, ok := record["DIVI"].(string); ok {
			division.Division = strings.TrimSpace(val)
		}
		if val, ok := record["TX15"].(string); ok {
			division.DivisionName = strings.TrimSpace(val)
		}
		if val, ok := record["FACI"].(string); ok {
			division.Facility = strings.TrimSpace(val)
		}
		if val, ok := record["WHLO"].(string); ok {
			division.Warehouse = strings.TrimSpace(val)
		}

		// Skip system divisions and empty divisions
		if division.Division != "" && division.Division != "991" && division.Division != "992" {
			divisions = append(divisions, division)
		}
	}
	return divisions
}

func parseWarehouses(records []map[string]interface{}) []compass.M3Warehouse {
	warehouses := make([]compass.M3Warehouse, 0, len(records))
	for _, record := range records {
		warehouse := compass.M3Warehouse{}

		if val, ok := record["CONO"].(string); ok {
			warehouse.CompanyNumber = strings.TrimSpace(val)
		}
		if val, ok := record["WHLO"].(string); ok {
			warehouse.Warehouse = strings.TrimSpace(val)
		}
		if val, ok := record["WHNM"].(string); ok {
			warehouse.WarehouseName = strings.TrimSpace(val)
		}
		if val, ok := record["DIVI"].(string); ok {
			warehouse.Division = strings.TrimSpace(val)
		}
		if val, ok := record["FACI"].(string); ok {
			warehouse.Facility = strings.TrimSpace(val)
		}

		// Skip empty warehouses
		if warehouse.Warehouse != "" {
			warehouses = append(warehouses, warehouse)
		}
	}
	return warehouses
}

func parseManufacturingOrderTypes(records []map[string]interface{}, fallbackCompanyNumber string) []compass.M3ManufacturingOrderType {
	orderTypes := make([]compass.M3ManufacturingOrderType, 0, len(records))
	for _, record := range records {
		orderType := compass.M3ManufacturingOrderType{}

		// Get company number from record (PMS120MI includes CONO in response)
		if val, ok := record["CONO"].(string); ok {
			orderType.CompanyNumber = strings.TrimSpace(val)
		} else {
			// Fallback to parameter if not in record
			orderType.CompanyNumber = fallbackCompanyNumber
		}

		// PMS120MI uses ORTY field
		if val, ok := record["ORTY"].(string); ok {
			orderType.OrderType = strings.TrimSpace(val)
		}
		if val, ok := record["TX40"].(string); ok {
			orderType.OrderTypeDescription = strings.TrimSpace(val)
		}
		if val, ok := record["LNCD"].(string); ok {
			orderType.LanguageCode = strings.TrimSpace(val)
		} else {
			orderType.LanguageCode = "GB" // Default language
		}

		// Skip empty order types
		if orderType.OrderType != "" && orderType.CompanyNumber != "" {
			orderTypes = append(orderTypes, orderType)
		}
	}
	return orderTypes
}

func parseCustomerOrderTypes(records []map[string]interface{}, fallbackCompanyNumber string) []compass.M3CustomerOrderType {
	orderTypes := make([]compass.M3CustomerOrderType, 0, len(records))
	for _, record := range records {
		orderType := compass.M3CustomerOrderType{}

		// Get company number from record if present, otherwise use fallback
		if val, ok := record["CONO"].(string); ok {
			orderType.CompanyNumber = strings.TrimSpace(val)
		} else {
			// Fallback to parameter if not in record (OIS100MI doesn't return CONO)
			orderType.CompanyNumber = fallbackCompanyNumber
		}

		// OIS100MI uses ORTP field (not ORTY)
		if val, ok := record["ORTP"].(string); ok {
			orderType.OrderType = strings.TrimSpace(val)
		}
		if val, ok := record["TX40"].(string); ok {
			orderType.OrderTypeDescription = strings.TrimSpace(val)
		}
		if val, ok := record["LNCD"].(string); ok {
			orderType.LanguageCode = strings.TrimSpace(val)
		} else {
			orderType.LanguageCode = "GB" // Default language
		}

		// Skip empty order types
		if orderType.OrderType != "" && orderType.CompanyNumber != "" {
			orderTypes = append(orderTypes, orderType)
		}
	}
	return orderTypes
}

func getErrorMessage(result *m3api.BulkResultItem) string {
	if result == nil {
		return "unknown error"
	}
	if result.ErrorMessage != "" {
		return result.ErrorMessage
	}
	if result.NotProcessed {
		return "transaction not processed"
	}
	return "unknown error"
}
