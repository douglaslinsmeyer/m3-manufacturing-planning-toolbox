package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/pinggolf/m3-planning-tools/internal/m3api"
	"github.com/pinggolf/m3-planning-tools/internal/services"
)

// LoginRequest represents the login request payload
type LoginRequest struct {
	Environment string `json:"environment"` // "TRN" or "PRD"
}

// LoginResponse represents the login response
type LoginResponse struct {
	AuthURL string `json:"authUrl"`
}

// AuthStatusResponse represents the authentication status
type AuthStatusResponse struct {
	Authenticated bool   `json:"authenticated"`
	Environment   string `json:"environment,omitempty"`
	UserContext   *UserContextResponse `json:"userContext,omitempty"`
}

// UserContextResponse represents the user's organizational context
type UserContextResponse struct {
	Company   string `json:"company,omitempty"`
	Division  string `json:"division,omitempty"`
	Facility  string `json:"facility,omitempty"`
	Warehouse string `json:"warehouse,omitempty"`
}

// handleLogin initiates the OAuth login flow
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate environment
	if req.Environment != "TRN" && req.Environment != "PRD" {
		http.Error(w, "Invalid environment. Must be TRN or PRD", http.StatusBadRequest)
		return
	}

	// Get session
	session, _ := s.sessionStore.Get(r, "m3-session")

	// Store environment in session
	session.Values["environment"] = req.Environment

	// Save session
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Generate OAuth authorization URL
	authURL, err := s.authManager.GetAuthorizationURL(req.Environment)
	if err != nil {
		http.Error(w, "Failed to generate authorization URL", http.StatusInternalServerError)
		return
	}

	// Return auth URL to frontend
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		AuthURL: authURL,
	})
}

// handleAuthCallback handles the OAuth callback
func (s *Server) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Get session
	session, _ := s.sessionStore.Get(r, "m3-session")

	// Get environment from session
	environment, ok := session.Values["environment"].(string)
	if !ok {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}

	// Extract authorization code from query parameters
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for tokens
	tokens, err := s.authManager.ExchangeCodeForTokens(r.Context(), environment, code)
	if err != nil {
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	// Store tokens and authentication status in session
	session.Values["authenticated"] = true
	session.Values["access_token"] = tokens.AccessToken
	session.Values["refresh_token"] = tokens.RefreshToken
	session.Values["token_expiry"] = tokens.Expiry.Unix()

	// Get M3 API client to load user defaults
	m3Client, err := s.getM3APIClient(r)
	if err != nil {
		log.Printf("ERROR: Failed to initialize M3 API client during auth: %v\n", err)
		session.Values["context_load_error"] = err.Error()
	} else {
		// ALWAYS prime the cache in the background - this populates companies/divisions/facilities/warehouses
		// so users can select them even if LoadUserDefaults fails
		go s.primeContextCache(environment, m3Client)

		// Try to load user defaults from M3 (this may fail but shouldn't block login)
		if err := s.contextService.LoadUserDefaults(r.Context(), session, m3Client); err != nil {
			// Log error but don't fail login - user can select context manually
			log.Printf("WARNING: Failed to load user defaults from M3 (user can select manually): %v\n", err)
			session.Values["context_load_error"] = err.Error()
		} else {
			// Success - clear any previous errors
			delete(session.Values, "context_load_error")
			log.Printf("INFO: Successfully loaded user defaults for environment %s\n", environment)
		}
	}

	// Save session with user defaults
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Redirect to frontend
	http.Redirect(w, r, s.config.FrontendURL, http.StatusFound)
}

// handleLogout logs out the user
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Get session
	session, _ := s.sessionStore.Get(r, "m3-session")

	// Store the old environment for logging purposes
	oldEnvironment, _ := session.Values["environment"].(string)

	// Clear session
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1

	// Save session
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to clear session", http.StatusInternalServerError)
		return
	}

	// TODO: Clear any cached data for this user's environment
	// This is where we would clear NATS job data, cached snapshots, etc.

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":      "logged out",
		"environment": oldEnvironment,
	})
}

// handleAuthStatus returns the current authentication status
func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")

	authenticated, ok := session.Values["authenticated"].(bool)
	if !ok || !authenticated {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthStatusResponse{
			Authenticated: false,
		})
		return
	}

	// Get environment and user context
	environment, _ := session.Values["environment"].(string)

	var userContext *UserContextResponse
	if company, ok := session.Values["user_company"].(string); ok {
		userContext = &UserContextResponse{
			Company:   company,
			Division:  getSessionString(session, "user_division"),
			Facility:  getSessionString(session, "user_facility"),
			Warehouse: getSessionString(session, "user_warehouse"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthStatusResponse{
		Authenticated: true,
		Environment:   environment,
		UserContext:   userContext,
	})
}

// handleGetContext returns the user's current organizational context
func (s *Server) handleGetContext(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")

	userContext := &UserContextResponse{
		Company:   getSessionString(session, "company"),
		Division:  getSessionString(session, "division"),
		Facility:  getSessionString(session, "facility"),
		Warehouse: getSessionString(session, "warehouse"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userContext)
}

// handleSetContext sets the user's organizational context
func (s *Server) handleSetContext(w http.ResponseWriter, r *http.Request) {
	var req UserContextResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session, _ := s.sessionStore.Get(r, "m3-session")

	// Update session with new context
	if req.Company != "" {
		session.Values["company"] = req.Company
	}
	if req.Division != "" {
		session.Values["division"] = req.Division
	}
	if req.Facility != "" {
		session.Values["facility"] = req.Facility
	}
	if req.Warehouse != "" {
		session.Values["warehouse"] = req.Warehouse
	}

	// Save session
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

// Helper function to get string from session
func getSessionString(session *sessions.Session, key string) string {
	if val, ok := session.Values[key].(string); ok {
		return val
	}
	return ""
}

// primeContextCache populates the M3 context cache after login
func (s *Server) primeContextCache(environment string, m3Client *m3api.Client) {
	ctx := context.Background()
	repo := services.NewContextRepository(s.db, m3Client, environment)

	fmt.Printf("Priming context cache for %s environment...\n", environment)

	// Prime companies cache
	companies, err := repo.GetCompanies(ctx, false) // Use cache if available
	if err != nil {
		fmt.Printf("Warning: Failed to prime companies cache: %v\n", err)
		return
	}
	fmt.Printf("Cached %d companies for %s\n", len(companies), environment)

	// Prime facilities cache
	facilities, err := repo.GetFacilities(ctx, false)
	if err != nil {
		fmt.Printf("Warning: Failed to prime facilities cache: %v\n", err)
	} else {
		fmt.Printf("Cached %d facilities for %s\n", len(facilities), environment)
	}

	// Prime divisions and warehouses for each company (limit to first 3 companies to avoid long delays)
	maxCompanies := 3
	if len(companies) > maxCompanies {
		companies = companies[:maxCompanies]
	}

	for _, company := range companies {
		// Prime divisions
		divisions, err := repo.GetDivisions(ctx, company.CompanyNumber, false)
		if err != nil {
			fmt.Printf("Warning: Failed to prime divisions cache for company %s: %v\n", company.CompanyNumber, err)
			continue
		}

		// Prime warehouses
		warehouses, err := repo.GetFilteredWarehouses(ctx, company.CompanyNumber, nil, nil)
		if err != nil {
			fmt.Printf("Warning: Failed to prime warehouses cache for company %s: %v\n", company.CompanyNumber, err)
			continue
		}

		fmt.Printf("Cached %d divisions and %d warehouses for company %s\n",
			len(divisions), len(warehouses), company.CompanyNumber)
	}

	fmt.Printf("Context cache priming completed for %s\n", environment)
}
