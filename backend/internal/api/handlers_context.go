package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pinggolf/m3-planning-tools/internal/m3api"
	"github.com/pinggolf/m3-planning-tools/internal/services"
)

// M3 Organizational Hierarchy Response Types

// M3CompanyResponse represents a company
type M3CompanyResponse struct {
	CompanyNumber string `json:"companyNumber"`
	CompanyName   string `json:"companyName"`
	Currency      string `json:"currency"`
}

// M3DivisionResponse represents a division
type M3DivisionResponse struct {
	CompanyNumber string `json:"companyNumber"`
	Division      string `json:"division"`
	DivisionName  string `json:"divisionName"`
	Facility      string `json:"facility,omitempty"`
	Warehouse     string `json:"warehouse,omitempty"`
}

// M3FacilityResponse represents a facility
type M3FacilityResponse struct {
	CompanyNumber string `json:"companyNumber"`
	Facility      string `json:"facility"`
	FacilityName  string `json:"facilityName"`
	Division      string `json:"division,omitempty"`
	Warehouse     string `json:"warehouse,omitempty"`
}

// M3WarehouseResponse represents a warehouse
type M3WarehouseResponse struct {
	CompanyNumber string `json:"companyNumber"`
	Warehouse     string `json:"warehouse"`
	WarehouseName string `json:"warehouseName"`
	Division      string `json:"division,omitempty"`
	Facility      string `json:"facility,omitempty"`
}

// M3ManufacturingOrderTypeResponse represents a manufacturing order type
type M3ManufacturingOrderTypeResponse struct {
	CompanyNumber        string `json:"companyNumber"`
	OrderType            string `json:"orderType"`
	OrderTypeDescription string `json:"orderTypeDescription"`
	LanguageCode         string `json:"languageCode"`
}

// M3CustomerOrderTypeResponse represents a customer order type
type M3CustomerOrderTypeResponse struct {
	CompanyNumber        string `json:"companyNumber"`
	OrderType            string `json:"orderType"`
	OrderTypeDescription string `json:"orderTypeDescription"`
	LanguageCode         string `json:"languageCode"`
}

// EffectiveContextResponse represents the effective context
type EffectiveContextResponse struct {
	Environment           string                `json:"environment"`
	Company               string                `json:"company"`
	Division              string                `json:"division"`
	Facility              string                `json:"facility"`
	Warehouse             string                `json:"warehouse"`
	HasTemporaryOverrides bool                  `json:"hasTemporaryOverrides"`
	UserDefaults          *UserContextResponse  `json:"userDefaults"`
	LoadError             string                `json:"loadError,omitempty"`
}

// TemporaryOverrideRequest represents a temporary override update
type TemporaryOverrideRequest struct {
	Company   *string `json:"company,omitempty"`
	Division  *string `json:"division,omitempty"`
	Facility  *string `json:"facility,omitempty"`
	Warehouse *string `json:"warehouse,omitempty"`
}

// handleGetEffectiveContext returns the effective context (temporary overrides + user defaults)
func (s *Server) handleGetEffectiveContext(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")

	// Get environment from session
	environment, _ := session.Values["environment"].(string)
	if environment == "" {
		environment = "PRD" // Default fallback
	}

	// Get effective context
	effective := s.contextService.GetEffectiveContext(session)
	userDefaults := s.contextService.GetUserDefaults(session)
	hasOverrides := s.contextService.HasTemporaryOverrides(session)

	// Get load error if any
	loadError := ""
	if err, ok := session.Values["context_load_error"].(string); ok {
		loadError = err
	}

	response := EffectiveContextResponse{
		Environment:           environment,
		Company:               effective.Company,
		Division:              effective.Division,
		Facility:              effective.Facility,
		Warehouse:             effective.Warehouse,
		HasTemporaryOverrides: hasOverrides,
		UserDefaults: &UserContextResponse{
			Company:   userDefaults.Company,
			Division:  userDefaults.Division,
			Facility:  userDefaults.Facility,
			Warehouse: userDefaults.Warehouse,
		},
		LoadError: loadError,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSetTemporaryOverride sets a temporary context override
func (s *Server) handleSetTemporaryOverride(w http.ResponseWriter, r *http.Request) {
	var req TemporaryOverrideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session, _ := s.sessionStore.Get(r, "m3-session")

	// Update temporary overrides
	if req.Company != nil {
		s.contextService.SetTemporaryOverride(session, "company", *req.Company)
	}
	if req.Division != nil {
		s.contextService.SetTemporaryOverride(session, "division", *req.Division)
	}
	if req.Facility != nil {
		s.contextService.SetTemporaryOverride(session, "facility", *req.Facility)
	}
	if req.Warehouse != nil {
		s.contextService.SetTemporaryOverride(session, "warehouse", *req.Warehouse)
	}

	// Save session
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Return updated effective context
	s.handleGetEffectiveContext(w, r)
}

// handleClearTemporaryOverrides clears all temporary overrides
func (s *Server) handleClearTemporaryOverrides(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")

	s.contextService.ClearTemporaryOverrides(session)

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Return updated effective context
	s.handleGetEffectiveContext(w, r)
}

// handleRetryLoadContext retries loading user context from M3
func (s *Server) handleRetryLoadContext(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)

	m3Client, err := s.getM3APIClient(r)
	if err != nil {
		http.Error(w, "Failed to initialize M3 API client", http.StatusInternalServerError)
		return
	}

	if err := s.contextService.LoadUserDefaults(r.Context(), session, m3Client); err != nil {
		session.Values["context_load_error"] = err.Error()
		session.Save(r, w)
		http.Error(w, fmt.Sprintf("Failed to load context: %v", err), http.StatusInternalServerError)
		return
	}

	delete(session.Values, "context_load_error")
	session.Save(r, w)

	// Prime cache in background
	go s.primeContextCache(environment, m3Client)

	// Return effective context
	effective := s.contextService.GetEffectiveContext(session)
	userDefaults := s.contextService.GetUserDefaults(session)

	// Get environment (already fetched above on line 169)
	if environment == "" {
		environment = "PRD" // Default fallback
	}

	response := EffectiveContextResponse{
		Environment:           environment,
		Company:               effective.Company,
		Division:              effective.Division,
		Facility:              effective.Facility,
		Warehouse:             effective.Warehouse,
		HasTemporaryOverrides: s.contextService.HasTemporaryOverrides(session),
		UserDefaults: &UserContextResponse{
			Company:   userDefaults.Company,
			Division:  userDefaults.Division,
			Facility:  userDefaults.Facility,
			Warehouse: userDefaults.Warehouse,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListCompanies returns all available companies
func (s *Server) handleListCompanies(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)

	// Get context repository for this environment
	repo, err := s.getContextRepositoryForRequest(r, environment)
	if err != nil {
		http.Error(w, "Failed to get context repository", http.StatusInternalServerError)
		return
	}

	companies, err := repo.GetCompanies(r.Context(), false)
	if err != nil {
		http.Error(w, "Failed to fetch companies", http.StatusInternalServerError)
		return
	}

	// Map to response format
	response := make([]M3CompanyResponse, len(companies))
	for i, c := range companies {
		response[i] = M3CompanyResponse{
			CompanyNumber: c.CompanyNumber,
			CompanyName:   c.CompanyName,
			Currency:      c.Currency,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListDivisions returns divisions for a company
func (s *Server) handleListDivisions(w http.ResponseWriter, r *http.Request) {
	companyNumber := r.URL.Query().Get("company")
	if companyNumber == "" {
		http.Error(w, "company parameter required", http.StatusBadRequest)
		return
	}

	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)

	repo, err := s.getContextRepositoryForRequest(r, environment)
	if err != nil {
		http.Error(w, "Failed to get context repository", http.StatusInternalServerError)
		return
	}

	divisions, err := repo.GetDivisions(r.Context(), companyNumber, false)
	if err != nil {
		http.Error(w, "Failed to fetch divisions", http.StatusInternalServerError)
		return
	}

	// Map to response format
	response := make([]M3DivisionResponse, len(divisions))
	for i, d := range divisions {
		response[i] = M3DivisionResponse{
			CompanyNumber: d.CompanyNumber,
			Division:      d.Division,
			DivisionName:  d.DivisionName,
			Facility:      d.Facility,
			Warehouse:     d.Warehouse,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListFacilities returns all facilities
func (s *Server) handleListFacilities(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)

	repo, err := s.getContextRepositoryForRequest(r, environment)
	if err != nil {
		http.Error(w, "Failed to get context repository", http.StatusInternalServerError)
		return
	}

	facilities, err := repo.GetFacilities(r.Context(), false)
	if err != nil {
		http.Error(w, "Failed to fetch facilities", http.StatusInternalServerError)
		return
	}

	// Map to response format
	response := make([]M3FacilityResponse, len(facilities))
	for i, f := range facilities {
		response[i] = M3FacilityResponse{
			CompanyNumber: f.CompanyNumber,
			Facility:      f.Facility,
			FacilityName:  f.FacilityName,
			Division:      f.Division,
			Warehouse:     f.Warehouse,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListWarehouses returns warehouses filtered by company/division/facility
func (s *Server) handleListWarehouses(w http.ResponseWriter, r *http.Request) {
	companyNumber := r.URL.Query().Get("company")
	if companyNumber == "" {
		http.Error(w, "company parameter required", http.StatusBadRequest)
		return
	}

	division := r.URL.Query().Get("division")
	facility := r.URL.Query().Get("facility")

	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)

	repo, err := s.getContextRepositoryForRequest(r, environment)
	if err != nil {
		http.Error(w, "Failed to get context repository", http.StatusInternalServerError)
		return
	}

	var divPtr, faciPtr *string
	if division != "" {
		divPtr = &division
	}
	if facility != "" {
		faciPtr = &facility
	}

	warehouses, err := repo.GetFilteredWarehouses(r.Context(), companyNumber, divPtr, faciPtr)
	if err != nil {
		http.Error(w, "Failed to fetch warehouses", http.StatusInternalServerError)
		return
	}

	// Map to response format
	response := make([]M3WarehouseResponse, len(warehouses))
	for i, wh := range warehouses {
		response[i] = M3WarehouseResponse{
			CompanyNumber: wh.CompanyNumber,
			Warehouse:     wh.Warehouse,
			WarehouseName: wh.WarehouseName,
			Division:      wh.Division,
			Facility:      wh.Facility,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListManufacturingOrderTypes returns manufacturing order types for a company
func (s *Server) handleListManufacturingOrderTypes(w http.ResponseWriter, r *http.Request) {
	companyNumber := r.URL.Query().Get("company")
	if companyNumber == "" {
		http.Error(w, "company parameter required", http.StatusBadRequest)
		return
	}

	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)

	repo, err := s.getContextRepositoryForRequest(r, environment)
	if err != nil {
		http.Error(w, "Failed to get context repository", http.StatusInternalServerError)
		return
	}

	orderTypes, err := repo.GetManufacturingOrderTypes(r.Context(), companyNumber, false)
	if err != nil {
		http.Error(w, "Failed to fetch manufacturing order types", http.StatusInternalServerError)
		return
	}

	// Map to response format
	response := make([]M3ManufacturingOrderTypeResponse, len(orderTypes))
	for i, ot := range orderTypes {
		response[i] = M3ManufacturingOrderTypeResponse{
			CompanyNumber:        ot.CompanyNumber,
			OrderType:            ot.OrderType,
			OrderTypeDescription: ot.OrderTypeDescription,
			LanguageCode:         ot.LanguageCode,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListCustomerOrderTypes returns customer order types for a company
func (s *Server) handleListCustomerOrderTypes(w http.ResponseWriter, r *http.Request) {
	companyNumber := r.URL.Query().Get("company")
	if companyNumber == "" {
		http.Error(w, "company parameter required", http.StatusBadRequest)
		return
	}

	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)

	repo, err := s.getContextRepositoryForRequest(r, environment)
	if err != nil {
		http.Error(w, "Failed to get context repository", http.StatusInternalServerError)
		return
	}

	orderTypes, err := repo.GetCustomerOrderTypes(r.Context(), companyNumber, false)
	if err != nil {
		http.Error(w, "Failed to fetch customer order types", http.StatusInternalServerError)
		return
	}

	// Map to response format
	response := make([]M3CustomerOrderTypeResponse, len(orderTypes))
	for i, ot := range orderTypes {
		response[i] = M3CustomerOrderTypeResponse{
			CompanyNumber:        ot.CompanyNumber,
			OrderType:            ot.OrderType,
			OrderTypeDescription: ot.OrderTypeDescription,
			LanguageCode:         ot.LanguageCode,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper method to get context repository for an environment
// This creates a repository with an M3 API client that uses the request's session token
func (s *Server) getContextRepositoryForRequest(r *http.Request, environment string) (*services.ContextRepository, error) {
	// Get session
	session, _ := s.sessionStore.Get(r, "m3-session")

	// Get environment-specific config
	envConfig, err := s.config.GetEnvironmentConfig(environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment config: %w", err)
	}

	// Create a function to get the token from the session
	getToken := func() (string, error) {
		// Refresh token if needed (ignore refreshed flag - middleware handles persistence)
		_, err := s.authManager.RefreshTokenIfNeeded(session)
		if err != nil {
			return "", err
		}
		return s.authManager.GetAccessToken(session)
	}

	m3Client := m3api.NewClient(envConfig.APIBaseURL, getToken)
	return services.NewContextRepository(s.db, m3Client, environment), nil
}

// handleGetM3Config returns M3 portal configuration for deep linking
func (s *Server) handleGetM3Config(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")

	environment, ok := session.Values["environment"].(string)
	if !ok {
		http.Error(w, "No environment in session", http.StatusBadRequest)
		return
	}

	envConfig, err := s.config.GetEnvironmentConfig(environment)
	if err != nil {
		http.Error(w, "Failed to get environment config", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tenantId":    envConfig.TenantID,
		"instanceId":  envConfig.InstanceID,
		"environment": environment,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRefreshAllContext manually refreshes all context cache data
func (s *Server) handleRefreshAllContext(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)

	userID, err := s.getUserIDFromSession(r)
	if err != nil {
		log.Printf("ERROR: Failed to get user ID for cache refresh: %v", err)
		userID = "unknown"
	}

	startTime := time.Now()
	log.Printf("INFO: Manual context cache refresh triggered by user %s in environment %s", userID, environment)

	// Get context repository
	repo, err := s.getContextRepositoryForRequest(r, environment)
	if err != nil {
		log.Printf("ERROR: Failed to get context repository: %v", err)
		http.Error(w, "Failed to get context repository", http.StatusInternalServerError)
		return
	}

	// 1. Refresh companies (forceRefresh=true)
	companies, err := repo.GetCompanies(r.Context(), true)
	if err != nil {
		log.Printf("ERROR: Failed to refresh companies: %v", err)
		http.Error(w, "Failed to refresh companies", http.StatusInternalServerError)
		return
	}

	// 2. Refresh facilities
	facilities, err := repo.GetFacilities(r.Context(), true)
	if err != nil {
		log.Printf("ERROR: Failed to refresh facilities: %v", err)
		http.Error(w, "Failed to refresh facilities", http.StatusInternalServerError)
		return
	}

	// 3. Use bulk API to refresh ALL company-scoped entities in single call
	// This replaces the old sequential loop with 1 bulk request for:
	// - Divisions, Warehouses, MO Types, CO Types for ALL companies
	log.Printf("INFO: Using bulk API to refresh context for %d companies...", len(companies))
	if err := repo.RefreshAllContextBulk(r.Context(), companies); err != nil {
		log.Printf("ERROR: Bulk context refresh failed: %v", err)
		http.Error(w, "Failed to refresh context via bulk API", http.StatusInternalServerError)
		return
	}

	// Query database to get actual counts after bulk refresh
	divisionCount := 0
	warehouseCount := 0
	for _, company := range companies {
		divs, _ := repo.GetDivisions(r.Context(), company.CompanyNumber, false)
		divisionCount += len(divs)

		whs, _ := repo.GetFilteredWarehouses(r.Context(), company.CompanyNumber, nil, nil)
		warehouseCount += len(whs)
	}

	elapsed := time.Since(startTime)

	log.Printf("INFO: Context cache refresh completed: %d companies, %d divisions, %d facilities, %d warehouses in %dms",
		len(companies), divisionCount, len(facilities), warehouseCount, elapsed.Milliseconds())

	// Log audit trail
	if s.auditService != nil {
		s.auditService.Log(r.Context(), services.AuditParams{
			Environment: environment,
			EntityType:  "context_cache",
			Operation:   "refresh_all",
			UserID:      userID,
			Metadata: map[string]interface{}{
				"companies_count":  len(companies),
				"divisions_count":  divisionCount,
				"facilities_count": len(facilities),
				"warehouses_count": warehouseCount,
				"duration_ms":      elapsed.Milliseconds(),
			},
		})
	}

	// Return summary
	refreshResponse := map[string]interface{}{
		"status":              "success",
		"message":             "Context cache refreshed successfully",
		"companiesRefreshed":  len(companies),
		"divisionsRefreshed":  divisionCount,
		"facilitiesRefreshed": len(facilities),
		"warehousesRefreshed": warehouseCount,
		"durationMs":          elapsed.Milliseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(refreshResponse)
}

// handleGetCacheStatus returns cache status for all context resources
func (s *Server) handleGetCacheStatus(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")
	environment, _ := session.Values["environment"].(string)

	// Query cache status from database
	statuses, err := s.db.GetContextCacheStatus(r.Context(), environment)
	if err != nil {
		log.Printf("ERROR: Failed to get cache status: %v", err)
		http.Error(w, "Failed to get cache status", http.StatusInternalServerError)
		return
	}

	// Convert to API response format
	type CacheStatusResponse struct {
		ResourceType string `json:"resourceType"`
		RecordCount  int    `json:"recordCount"`
		LastRefresh  string `json:"lastRefresh,omitempty"`
		IsStale      bool   `json:"isStale"`
	}

	statusResponse := make([]CacheStatusResponse, len(statuses))
	for i, status := range statuses {
		statusResponse[i] = CacheStatusResponse{
			ResourceType: status.ResourceType,
			RecordCount:  status.RecordCount,
			IsStale:      status.IsStale,
		}
		if status.LastRefresh.Valid {
			statusResponse[i].LastRefresh = status.LastRefresh.Time.Format(time.RFC3339)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statusResponse)
}
