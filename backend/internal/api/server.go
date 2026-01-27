package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pinggolf/m3-planning-tools/internal/auth"
	"github.com/pinggolf/m3-planning-tools/internal/compass"
	"github.com/pinggolf/m3-planning-tools/internal/config"
	"github.com/pinggolf/m3-planning-tools/internal/db"
	"github.com/pinggolf/m3-planning-tools/internal/infor"
	"github.com/pinggolf/m3-planning-tools/internal/m3api"
	"github.com/pinggolf/m3-planning-tools/internal/queue"
	"github.com/pinggolf/m3-planning-tools/internal/services"
	"github.com/rs/cors"
)

// Server represents the API server
type Server struct {
	config                *config.Config
	db                    *db.Queries
	router                *mux.Router
	sessionStore          sessions.Store
	authManager           *auth.Manager
	natsManager           *queue.Manager
	contextService        *services.ContextService
	auditService          *services.AuditService
	userProfileService    *services.UserProfileService
	settingsService       *services.SettingsService
	detectorConfigService *services.DetectorConfigService
}

// NewServer creates a new API server instance
func NewServer(cfg *config.Config, queries *db.Queries, natsManager *queue.Manager, database *sql.DB) *Server {
	// Initialize session store (cookie-based for auth tokens only)
	// User profiles stored in Postgres to avoid cookie size limits and enable scaling
	sessionStore := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(cfg.SessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   cfg.AppEnv == "production",
		SameSite: http.SameSiteLaxMode,
	}

	// Initialize auth manager
	authManager := auth.NewManager(cfg, sessionStore)

	// Initialize context service
	// Note: We'll create a placeholder repository here since the actual repository
	// needs a Compass client with session-specific tokens
	contextService := services.NewContextService(nil)

	// Initialize audit service
	auditService := services.NewAuditService(queries)

	// Initialize user profile service (uses raw DB for JSONB operations)
	userProfileService := services.NewUserProfileService(database)

	// Link user profile service to context service (for cache-first user defaults loading)
	contextService.SetUserProfileService(userProfileService)

	// Initialize settings service
	settingsService := services.NewSettingsService(queries, auditService)

	// Initialize detector config service
	detectorConfigService := services.NewDetectorConfigService(queries)

	s := &Server{
		config:                cfg,
		db:                    queries,
		router:                mux.NewRouter(),
		sessionStore:          sessionStore,
		authManager:           authManager,
		natsManager:           natsManager,
		contextService:        contextService,
		auditService:          auditService,
		userProfileService:    userProfileService,
		settingsService:       settingsService,
		detectorConfigService: detectorConfigService,
	}

	s.setupRoutes()
	return s
}

// getCompassClient returns a Compass client for the current user session
func (s *Server) getCompassClient(r *http.Request) (*compass.Client, error) {
	session, _ := s.sessionStore.Get(r, "m3-session")

	// Get environment
	environment, ok := session.Values["environment"].(string)
	if !ok {
		return nil, fmt.Errorf("no environment in session")
	}

	// Get environment config
	envConfig, err := s.config.GetEnvironmentConfig(environment)
	if err != nil {
		return nil, err
	}

	// Create token getter function
	getToken := func() (string, error) {
		// Refresh token if needed (ignore refreshed flag - middleware handles persistence)
		_, err := s.authManager.RefreshTokenIfNeeded(session)
		if err != nil {
			return "", err
		}
		return s.authManager.GetAccessToken(session)
	}

	return compass.NewClient(envConfig.CompassBaseURL, getToken), nil
}

// getM3APIClient returns an M3 API client for the current user session
func (s *Server) getM3APIClient(r *http.Request) (*m3api.Client, error) {
	session, _ := s.sessionStore.Get(r, "m3-session")

	// Get environment
	environment, ok := session.Values["environment"].(string)
	if !ok {
		return nil, fmt.Errorf("no environment in session")
	}

	// Get environment config
	envConfig, err := s.config.GetEnvironmentConfig(environment)
	if err != nil {
		return nil, err
	}

	// Create token getter function
	getToken := func() (string, error) {
		// Refresh token if needed (ignore refreshed flag - middleware handles persistence)
		_, err := s.authManager.RefreshTokenIfNeeded(session)
		if err != nil {
			return "", err
		}
		return s.authManager.GetAccessToken(session)
	}

	return m3api.NewClient(envConfig.APIBaseURL, getToken), nil
}

// getInforClient returns an Infor API client for the current user session
func (s *Server) getInforClient(r *http.Request) (*infor.Client, error) {
	session, _ := s.sessionStore.Get(r, "m3-session")

	// Get environment
	environment, ok := session.Values["environment"].(string)
	if !ok {
		return nil, fmt.Errorf("no environment in session")
	}

	// Get environment config
	envConfig, err := s.config.GetEnvironmentConfig(environment)
	if err != nil {
		return nil, err
	}

	// Create token getter function
	getToken := func() (string, error) {
		// Refresh token if needed (ignore refreshed flag - middleware handles persistence)
		_, err := s.authManager.RefreshTokenIfNeeded(session)
		if err != nil {
			return "", err
		}
		return s.authManager.GetAccessToken(session)
	}

	// Base URL: https://mingle-ionapi.inforcloudsuite.com/{tenant}/
	baseURL := fmt.Sprintf("https://mingle-ionapi.inforcloudsuite.com/%s/", envConfig.TenantID)

	return infor.NewClient(baseURL, getToken), nil
}

// Router returns the configured HTTP router with CORS
func (s *Server) Router() http.Handler {
	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{s.config.CORSAllowedOrigins},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: s.config.CORSAllowCredentials,
		MaxAge:           300,
	})

	return c.Handler(s.router)
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// API version prefix
	api := s.router.PathPrefix("/api").Subrouter()

	// Health check (no auth required)
	api.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Auth routes
	authRouter := api.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/login", s.handleLogin).Methods("POST")
	authRouter.HandleFunc("/callback", s.handleAuthCallback).Methods("GET")
	authRouter.HandleFunc("/logout", s.handleLogout).Methods("POST")
	authRouter.HandleFunc("/status", s.handleAuthStatus).Methods("GET")

	// User context routes (for selecting company/division/facility/warehouse)
	authRouter.HandleFunc("/context", s.handleGetContext).Methods("GET")
	authRouter.HandleFunc("/context", s.handleSetContext).Methods("POST")

	// User profile routes (require authentication)
	authRouter.HandleFunc("/profile/refresh", s.handleRefreshProfile).Methods("POST")

	// Protected routes (require authentication)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(s.authMiddleware)

	// Context management routes
	contextRouter := protected.PathPrefix("/context").Subrouter()
	contextRouter.HandleFunc("/effective", s.handleGetEffectiveContext).Methods("GET")
	contextRouter.HandleFunc("/temporary", s.handleSetTemporaryOverride).Methods("POST")
	contextRouter.HandleFunc("/temporary", s.handleClearTemporaryOverrides).Methods("DELETE")
	contextRouter.HandleFunc("/retry-load", s.handleRetryLoadContext).Methods("POST")
	contextRouter.HandleFunc("/companies", s.handleListCompanies).Methods("GET")
	contextRouter.HandleFunc("/divisions", s.handleListDivisions).Methods("GET")
	contextRouter.HandleFunc("/facilities", s.handleListFacilities).Methods("GET")
	contextRouter.HandleFunc("/warehouses", s.handleListWarehouses).Methods("GET")
	contextRouter.HandleFunc("/manufacturing-order-types", s.handleListManufacturingOrderTypes).Methods("GET")
	contextRouter.HandleFunc("/customer-order-types", s.handleListCustomerOrderTypes).Methods("GET")

	// M3 Configuration (for deep linking)
	protected.HandleFunc("/m3-config", s.handleGetM3Config).Methods("GET")

	// Snapshot management
	protected.HandleFunc("/snapshot/refresh", s.handleSnapshotRefresh).Methods("POST")
	protected.HandleFunc("/snapshot/refresh/{jobId}/cancel", s.handleCancelRefresh).Methods("POST")
	protected.HandleFunc("/snapshot/status", s.handleSnapshotStatus).Methods("GET")
	protected.HandleFunc("/snapshot/summary", s.handleSnapshotSummary).Methods("GET")
	protected.HandleFunc("/snapshot/active-job", s.handleGetActiveJob).Methods("GET")
	protected.HandleFunc("/snapshot/progress/{jobId}", s.handleSnapshotProgressSSE).Methods("GET")

	// Production orders (unified MO/MOP view)
	protected.HandleFunc("/production-orders", s.handleListProductionOrders).Methods("GET")
	protected.HandleFunc("/production-orders/{id}", s.handleGetProductionOrder).Methods("GET")

	// Manufacturing orders (full MO details)
	protected.HandleFunc("/manufacturing-orders/{id}", s.handleGetManufacturingOrder).Methods("GET")

	// Planned manufacturing orders (full MOP details)
	protected.HandleFunc("/planned-orders/{id}", s.handleGetPlannedOrder).Methods("GET")

	// Analysis endpoints
	protected.HandleFunc("/analysis/inconsistencies", s.handleListInconsistencies).Methods("GET")
	protected.HandleFunc("/analysis/timeline", s.handleGetTimeline).Methods("GET")

	// Issue detection endpoints
	protected.HandleFunc("/issues", s.handleListIssues).Methods("GET")
	protected.HandleFunc("/issues/summary", s.handleGetIssueSummary).Methods("GET")
	protected.HandleFunc("/issues/{id}", s.handleGetIssueDetail).Methods("GET")
	protected.HandleFunc("/issues/{id}/ignore", s.handleIgnoreIssue).Methods("POST")
	protected.HandleFunc("/issues/{id}/unignore", s.handleUnignoreIssue).Methods("POST")
	protected.HandleFunc("/issues/{id}/delete-mop", s.handleDeletePlannedMO).Methods("POST")
	protected.HandleFunc("/issues/{id}/delete-mo", s.handleDeleteMO).Methods("POST")
	protected.HandleFunc("/issues/{id}/close-mo", s.handleCloseMO).Methods("POST")

	// Settings routes (user settings - authenticated users only)
	protected.HandleFunc("/settings/user", s.handleGetUserSettings).Methods("GET")
	protected.HandleFunc("/settings/user", s.handleUpdateUserSettings).Methods("PUT")

	// System settings routes (admin only)
	adminRouter := protected.PathPrefix("/settings/system").Subrouter()
	adminRouter.Use(s.adminMiddleware)
	adminRouter.HandleFunc("", s.handleGetSystemSettings).Methods("GET")
	adminRouter.HandleFunc("", s.handleUpdateSystemSettings).Methods("PUT")
}

// authMiddleware checks if the user is authenticated
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := s.sessionStore.Get(r, "m3-session")

		// Check if user is authenticated
		authenticated, ok := session.Values["authenticated"].(bool)
		if !ok || !authenticated {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if token is still valid and refresh if needed
		refreshed, err := s.authManager.RefreshTokenIfNeeded(session)
		if err != nil {
			http.Error(w, "Authentication expired", http.StatusUnauthorized)
			return
		}

		// Save session if tokens were refreshed to persist new token data
		if refreshed {
			if err := session.Save(r, w); err != nil {
				log.Printf("Failed to save session after token refresh: %v", err)
				// Don't fail the request - session might still work on next call
			}
		}

		next.ServeHTTP(w, r)
	})
}

// Health check handler
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
