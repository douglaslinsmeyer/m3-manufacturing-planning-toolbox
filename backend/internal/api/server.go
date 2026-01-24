package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pinggolf/m3-planning-tools/internal/auth"
	"github.com/pinggolf/m3-planning-tools/internal/compass"
	"github.com/pinggolf/m3-planning-tools/internal/config"
	"github.com/pinggolf/m3-planning-tools/internal/db"
	"github.com/pinggolf/m3-planning-tools/internal/m3api"
	"github.com/pinggolf/m3-planning-tools/internal/queue"
	"github.com/pinggolf/m3-planning-tools/internal/services"
	"github.com/rs/cors"
)

// Server represents the API server
type Server struct {
	config         *config.Config
	db             *db.Queries
	router         *mux.Router
	sessionStore   sessions.Store
	authManager    *auth.Manager
	natsManager    *queue.Manager
	contextService *services.ContextService
}

// NewServer creates a new API server instance
func NewServer(cfg *config.Config, queries *db.Queries, natsManager *queue.Manager) *Server {
	// Initialize session store
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

	s := &Server{
		config:         cfg,
		db:             queries,
		router:         mux.NewRouter(),
		sessionStore:   sessionStore,
		authManager:    authManager,
		natsManager:    natsManager,
		contextService: contextService,
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
		// Refresh token if needed
		if err := s.authManager.RefreshTokenIfNeeded(session); err != nil {
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
		// Refresh token if needed
		if err := s.authManager.RefreshTokenIfNeeded(session); err != nil {
			return "", err
		}
		return s.authManager.GetAccessToken(session)
	}

	return m3api.NewClient(envConfig.APIBaseURL, getToken), nil
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

	// Snapshot management
	protected.HandleFunc("/snapshot/refresh", s.handleSnapshotRefresh).Methods("POST")
	protected.HandleFunc("/snapshot/status", s.handleSnapshotStatus).Methods("GET")
	protected.HandleFunc("/snapshot/summary", s.handleSnapshotSummary).Methods("GET")
	protected.HandleFunc("/snapshot/progress/{jobId}", s.handleSnapshotProgressSSE).Methods("GET")

	// Production orders (unified MO/MOP view)
	protected.HandleFunc("/production-orders", s.handleListProductionOrders).Methods("GET")
	protected.HandleFunc("/production-orders/{id}", s.handleGetProductionOrder).Methods("GET")

	// Manufacturing orders (full MO details)
	protected.HandleFunc("/manufacturing-orders/{id}", s.handleGetManufacturingOrder).Methods("GET")

	// Planned manufacturing orders (full MOP details)
	protected.HandleFunc("/planned-orders/{id}", s.handleGetPlannedOrder).Methods("GET")

	// Customer orders
	protected.HandleFunc("/customer-orders", s.handleListCustomerOrders).Methods("GET")
	protected.HandleFunc("/customer-orders/{id}", s.handleGetCustomerOrder).Methods("GET")

	// Deliveries
	protected.HandleFunc("/deliveries", s.handleListDeliveries).Methods("GET")
	protected.HandleFunc("/deliveries/{id}", s.handleGetDelivery).Methods("GET")

	// Analysis endpoints
	protected.HandleFunc("/analysis/inconsistencies", s.handleListInconsistencies).Methods("GET")
	protected.HandleFunc("/analysis/timeline", s.handleGetTimeline).Methods("GET")
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
		if err := s.authManager.RefreshTokenIfNeeded(session); err != nil {
			http.Error(w, "Authentication expired", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Health check handler
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
