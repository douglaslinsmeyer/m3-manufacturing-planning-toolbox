package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/pinggolf/m3-planning-tools/internal/infor"
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
	Authenticated bool                 `json:"authenticated"`
	Environment   string               `json:"environment,omitempty"`
	UserContext   *UserContextResponse `json:"userContext,omitempty"`
	UserProfile   *UserProfileResponse `json:"userProfile,omitempty"`
}

// UserContextResponse represents the user's organizational context
type UserContextResponse struct {
	Company   string `json:"company,omitempty"`
	Division  string `json:"division,omitempty"`
	Facility  string `json:"facility,omitempty"`
	Warehouse string `json:"warehouse,omitempty"`
}

// UserProfileResponse represents the user's profile information for API responses
type UserProfileResponse struct {
	ID          string                     `json:"id"`
	UserName    string                     `json:"userName"`
	DisplayName string                     `json:"displayName"`
	Email       string                     `json:"email,omitempty"`
	Title       string                     `json:"title,omitempty"`
	Department  string                     `json:"department,omitempty"`
	Groups      []UserProfileGroupResponse `json:"groups,omitempty"`
	M3Info      *M3UserInfoResponse        `json:"m3Info,omitempty"`
}

// UserProfileGroupResponse represents a group/role assignment
type UserProfileGroupResponse struct {
	Value   string `json:"value"`
	Display string `json:"display"`
	Type    string `json:"type"`
}

// M3UserInfoResponse represents M3-specific user defaults
type M3UserInfoResponse struct {
	UserID           string `json:"userId"`
	FullName         string `json:"fullName"`
	DefaultCompany   string `json:"defaultCompany"`
	DefaultDivision  string `json:"defaultDivision"`
	DefaultFacility  string `json:"defaultFacility"`
	DefaultWarehouse string `json:"defaultWarehouse"`
	LanguageCode     string `json:"languageCode"`
	DateFormat       string `json:"dateFormat"`
	DateSeparator    string `json:"dateSeparator"`
	TimeSeparator    string `json:"timeSeparator"`
	TimeZone         string `json:"timeZone"`
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

	// Fetch and cache combined user profile (Infor + M3) in Postgres
	inforClient, err := s.getInforClient(r)
	if err != nil {
		log.Printf("WARNING: Failed to create Infor client: %v\n", err)
	} else {
		// Fetch Infor user management profile
		inforProfile, err := inforClient.GetUserProfile(r.Context())
		if err != nil {
			log.Printf("WARNING: Failed to fetch Infor user profile: %v\n", err)
		} else {
			// Create combined profile
			combinedProfile := &infor.CombinedUserProfile{
				UserProfile: *inforProfile,
			}

			// Fetch M3 user info (defaults and preferences)
			if m3Client, err := s.getM3APIClient(r); err == nil {
				if m3Info, err := infor.GetM3UserInfo(r.Context(), m3Client); err == nil {
					combinedProfile.M3Info = m3Info
					log.Printf("INFO: Successfully fetched M3 user info for: %s\n", m3Info.UserID)
				} else {
					log.Printf("WARNING: Failed to fetch M3 user info: %v\n", err)
				}
			}

			// Cache combined profile in Postgres with 15-min TTL
			if err := s.userProfileService.SetProfile(r.Context(), combinedProfile); err != nil {
				log.Printf("WARNING: Failed to cache user profile in Postgres: %v\n", err)
			} else {
				log.Printf("INFO: Successfully cached combined user profile for: %s (ID: %s)\n", inforProfile.DisplayName, inforProfile.ID)
				// Store user ID in session for quick lookups
				session.Values["user_profile_id"] = inforProfile.ID

				// Extract M3 defaults from profile to session (for fast context access)
				if combinedProfile.M3Info != nil {
					session.Values["user_company"] = combinedProfile.M3Info.DefaultCompany
					session.Values["user_division"] = combinedProfile.M3Info.DefaultDivision
					session.Values["user_facility"] = combinedProfile.M3Info.DefaultFacility
					session.Values["user_warehouse"] = combinedProfile.M3Info.DefaultWarehouse
					session.Values["user_full_name"] = combinedProfile.M3Info.FullName
					log.Printf("INFO: Populated session with M3 defaults from profile (Company: %s, Div: %s, Fac: %s, Whse: %s)\n",
						combinedProfile.M3Info.DefaultCompany,
						combinedProfile.M3Info.DefaultDivision,
						combinedProfile.M3Info.DefaultFacility,
						combinedProfile.M3Info.DefaultWarehouse)
				}
			}
		}
	}

	// Get M3 API client for context cache priming
	m3Client, err := s.getM3APIClient(r)
	if err != nil {
		log.Printf("ERROR: Failed to initialize M3 API client during auth: %v\n", err)
		session.Values["context_load_error"] = err.Error()
	} else {
		// ALWAYS prime the cache in the background - this populates companies/divisions/facilities/warehouses
		// so users can select them even if LoadUserDefaults fails
		go s.primeContextCache(environment, m3Client)

		// Check if M3 defaults already populated from profile (to avoid duplicate API call)
		if _, hasCompany := session.Values["user_company"].(string); !hasCompany {
			// M3 defaults not in session - call LoadUserDefaults as fallback
			log.Printf("INFO: M3 defaults not found in session, calling LoadUserDefaults\n")
			if err := s.contextService.LoadUserDefaults(r.Context(), session, m3Client); err != nil {
				// Log error but don't fail login - user can select context manually
				log.Printf("WARNING: Failed to load user defaults from M3 (user can select manually): %v\n", err)
				session.Values["context_load_error"] = err.Error()
			} else {
				// Success - clear any previous errors
				delete(session.Values, "context_load_error")
				log.Printf("INFO: Successfully loaded user defaults for environment %s\n", environment)
			}
		} else {
			// M3 defaults already in session from profile - skip LoadUserDefaults
			log.Printf("INFO: M3 defaults already in session from profile cache (skipping LoadUserDefaults)\n")
			delete(session.Values, "context_load_error")
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

	// Delete user profile from Postgres cache
	if userProfileID, ok := session.Values["user_profile_id"].(string); ok && userProfileID != "" {
		if err := s.userProfileService.DeleteProfile(r.Context(), userProfileID); err != nil {
			log.Printf("WARNING: Failed to delete user profile from cache: %v\n", err)
		}
	}

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

	// Get combined user profile from Postgres cache (15-min TTL)
	var userProfile *UserProfileResponse
	if userProfileID, ok := session.Values["user_profile_id"].(string); ok && userProfileID != "" {
		profile, err := s.userProfileService.GetProfile(r.Context(), userProfileID)

		// If profile is nil (cache expired) or error occurred, try to refresh it
		if profile == nil || err != nil {
			if err != nil {
				log.Printf("WARNING: Failed to get cached user profile: %v, attempting refresh\n", err)
			} else {
				log.Printf("INFO: User profile cache expired, refreshing from Infor API\n")
			}

			// Refresh profile from Infor and M3
			if inforClient, clientErr := s.getInforClient(r); clientErr == nil {
				if inforProfile, profileErr := inforClient.GetUserProfile(r.Context()); profileErr == nil {
					// Create combined profile
					combinedProfile := &infor.CombinedUserProfile{
						UserProfile: *inforProfile,
					}

					// Fetch M3 user info
					if m3Client, m3Err := s.getM3APIClient(r); m3Err == nil {
						if m3Info, m3InfoErr := infor.GetM3UserInfo(r.Context(), m3Client); m3InfoErr == nil {
							combinedProfile.M3Info = m3Info
							log.Printf("INFO: Refreshed M3 user info for: %s\n", m3Info.UserID)
						}
					}

					// Re-cache the refreshed profile
					if cacheErr := s.userProfileService.SetProfile(r.Context(), combinedProfile); cacheErr == nil {
						profile = combinedProfile
						log.Printf("INFO: User profile refreshed and cached for: %s\n", inforProfile.DisplayName)
					} else {
						log.Printf("WARNING: Failed to cache refreshed profile: %v\n", cacheErr)
					}
				} else {
					log.Printf("WARNING: Failed to refresh Infor profile: %v\n", profileErr)
				}
			} else {
				log.Printf("WARNING: Failed to create Infor client for profile refresh: %v\n", clientErr)
			}
		}

		if profile != nil {
			// Get primary email
			primaryEmail := ""
			for _, email := range profile.Emails {
				if email.Primary {
					primaryEmail = email.Value
					break
				}
			}

			// Convert groups
			groups := make([]UserProfileGroupResponse, len(profile.Groups))
			for i, g := range profile.Groups {
				groups[i] = UserProfileGroupResponse{
					Value:   g.Value,
					Display: g.Display,
					Type:    g.Type,
				}
			}

			// Convert M3 info if available
			var m3Info *M3UserInfoResponse
			if profile.M3Info != nil {
				m3Info = &M3UserInfoResponse{
					UserID:           profile.M3Info.UserID,
					FullName:         profile.M3Info.FullName,
					DefaultCompany:   profile.M3Info.DefaultCompany,
					DefaultDivision:  profile.M3Info.DefaultDivision,
					DefaultFacility:  profile.M3Info.DefaultFacility,
					DefaultWarehouse: profile.M3Info.DefaultWarehouse,
					LanguageCode:     profile.M3Info.LanguageCode,
					DateFormat:       profile.M3Info.DateFormat,
					DateSeparator:    profile.M3Info.DateSeparator,
					TimeSeparator:    profile.M3Info.TimeSeparator,
					TimeZone:         profile.M3Info.TimeZone,
				}
			}

			userProfile = &UserProfileResponse{
				ID:          profile.ID,
				UserName:    profile.UserName,
				DisplayName: profile.DisplayName,
				Email:       primaryEmail,
				Title:       profile.Title,
				Department:  profile.Department,
				Groups:      groups,
				M3Info:      m3Info,
			}
		} else if err != nil {
			log.Printf("WARNING: Failed to get user profile from cache: %v\n", err)
		}
	}

	response := AuthStatusResponse{
		Authenticated: true,
		Environment:   environment,
		UserContext:   userContext,
		UserProfile:   userProfile,
	}

	// Debug: Log if M3 info is included
	if userProfile != nil && userProfile.M3Info != nil {
		log.Printf("DEBUG: Auth status returning M3 info for user: %s (Company: %s)\n", userProfile.M3Info.UserID, userProfile.M3Info.DefaultCompany)
	} else if userProfile != nil {
		log.Printf("DEBUG: Auth status returning profile WITHOUT M3 info for user: %s\n", userProfile.UserName)
	} else {
		log.Printf("DEBUG: Auth status returning NO user profile\n")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

// primeContextCache populates the M3 context cache after login using bulk API operations
func (s *Server) primeContextCache(environment string, m3Client *m3api.Client) {
	ctx := context.Background()
	repo := services.NewContextRepository(s.db, m3Client, environment)

	fmt.Printf("Priming context cache for %s environment with bulk API operations...\n", environment)

	// 1. Prime companies cache (single call)
	companies, err := repo.GetCompanies(ctx, true) // Force refresh
	if err != nil {
		fmt.Printf("ERROR: Failed to prime companies cache: %v\n", err)
		return
	}
	fmt.Printf("  %s: Cached %d companies\n", environment, len(companies))

	// 2. Prime facilities cache (single call)
	facilities, err := repo.GetFacilities(ctx, true) // Force refresh
	if err != nil {
		fmt.Printf("WARNING: Failed to prime facilities cache: %v\n", err)
	} else {
		fmt.Printf("  %s: Cached %d facilities\n", environment, len(facilities))
	}

	// 3. Use NEW bulk API to prime ALL company-scoped entities in single call
	// This replaces the old sequential loop with 1 bulk request for:
	// - Divisions, Warehouses, MO Types, CO Types for ALL companies
	if err := repo.RefreshAllContextBulk(ctx, companies); err != nil {
		fmt.Printf("ERROR: Bulk context refresh failed: %v\n", err)
		return
	}

	fmt.Printf("Context cache priming completed for %s using bulk operations\n", environment)
}

// handleRefreshProfile re-fetches user profile from Infor API
func (s *Server) handleRefreshProfile(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")

	// Check authentication
	authenticated, ok := session.Values["authenticated"].(bool)
	if !ok || !authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	// Get Infor client
	inforClient, err := s.getInforClient(r)
	if err != nil {
		http.Error(w, "Failed to create Infor client", http.StatusInternalServerError)
		return
	}

	// Fetch fresh Infor profile
	inforProfile, err := inforClient.GetUserProfile(r.Context())
	if err != nil {
		log.Printf("ERROR: Failed to fetch Infor user profile: %v\n", err)
		http.Error(w, "Failed to fetch user profile", http.StatusInternalServerError)
		return
	}

	// Create combined profile
	combinedProfile := &infor.CombinedUserProfile{
		UserProfile: *inforProfile,
	}

	// Fetch M3 user info
	if m3Client, err := s.getM3APIClient(r); err == nil {
		if m3Info, err := infor.GetM3UserInfo(r.Context(), m3Client); err == nil {
			combinedProfile.M3Info = m3Info
			log.Printf("INFO: Successfully fetched M3 user info for: %s\n", m3Info.UserID)
		} else {
			log.Printf("WARNING: Failed to fetch M3 user info: %v\n", err)
		}
	}

	// Cache combined profile in Postgres with new 15-min TTL
	if err := s.userProfileService.SetProfile(r.Context(), combinedProfile); err != nil {
		log.Printf("ERROR: Failed to cache profile in Postgres: %v\n", err)
		http.Error(w, "Failed to cache profile", http.StatusInternalServerError)
		return
	}

	log.Printf("INFO: Successfully refreshed combined user profile for: %s (ID: %s)\n", inforProfile.DisplayName, inforProfile.ID)

	// Return updated combined profile
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(combinedProfile)
}
