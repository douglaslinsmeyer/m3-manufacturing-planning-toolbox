package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/pinggolf/m3-planning-tools/internal/auth"
	"github.com/pinggolf/m3-planning-tools/internal/infor"
	"github.com/pinggolf/m3-planning-tools/internal/m3api"
	"github.com/pinggolf/m3-planning-tools/internal/services"
)

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

// handleLogin initiates the Entra ID OIDC login flow. The M3 environment is
// fixed per deployment, so no environment selection happens at login.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")

	state, err := auth.GenerateState()
	if err != nil {
		http.Error(w, "Failed to generate login state", http.StatusInternalServerError)
		return
	}

	session.Values["oauth_state"] = state
	session.Values["environment"] = s.config.M3Env
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	authURL, err := s.entra.AuthorizationURL(state)
	if err != nil {
		log.Printf("ERROR: failed to build Entra authorization URL: %v", err)
		http.Error(w, "Failed to generate authorization URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{AuthURL: authURL})
}

// handleAuthCallback handles the Entra ID OAuth callback
func (s *Server) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")

	// CSRF protection: the state must match what login stored in the session
	expectedState, _ := session.Values["oauth_state"].(string)
	if expectedState == "" || r.URL.Query().Get("state") != expectedState {
		http.Error(w, "Invalid login state", http.StatusBadRequest)
		return
	}
	delete(session.Values, "oauth_state")

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	user, err := s.entra.ExchangeCode(r.Context(), code)
	if err != nil {
		log.Printf("ERROR: Entra code exchange failed: %v", err)
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	// Authenticated session: identity comes from Entra ID; M3 data access
	// uses the ION API service account and is not tied to this user.
	session.Values["authenticated"] = true
	session.Values["environment"] = s.config.M3Env
	session.Values["user_profile_id"] = user.ObjectID
	session.Values["user_id"] = user.Email
	session.Values["user_email"] = user.Email
	session.Values["user_full_name"] = user.Name
	session.Values["user_roles"] = strings.Join(user.Roles, ",")

	// Cache the profile in Postgres so /auth/status and the settings
	// handlers keep working against the same store as before.
	if err := s.userProfileService.SetProfile(r.Context(), entraProfile(user)); err != nil {
		log.Printf("WARNING: Failed to cache user profile in Postgres: %v", err)
	}

	// Prime the M3 context cache (companies/divisions/facilities/warehouses)
	// in the background using the service-account client. Per-user M3
	// defaults are no longer fetched at login — context selection comes from
	// the user's saved settings or manual selection.
	if m3Client, err := s.getM3APIClient(r); err != nil {
		log.Printf("WARNING: Failed to initialize M3 API client during auth: %v", err)
		session.Values["context_load_error"] = err.Error()
	} else {
		go s.primeContextCache(s.config.M3Env, m3Client)
		delete(session.Values, "context_load_error")
	}

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, s.config.FrontendURL, http.StatusFound)
}

// entraProfile synthesizes the cached profile record from Entra ID claims.
// App roles are stored as groups so the existing role plumbing keeps working.
func entraProfile(user *auth.EntraUser) *infor.CombinedUserProfile {
	groups := make([]infor.Group, len(user.Roles))
	for i, role := range user.Roles {
		groups[i] = infor.Group{Value: role, Display: role, Type: "App Role"}
	}
	return &infor.CombinedUserProfile{
		UserProfile: infor.UserProfile{
			ID:          user.ObjectID,
			UserName:    user.Email,
			DisplayName: user.Name,
			Emails:      []infor.Email{{Value: user.Email, Type: "work", Primary: true}},
			Groups:      groups,
		},
	}
}

// handleLogout logs out the user
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")

	oldEnvironment, _ := session.Values["environment"].(string)

	// Delete user profile from Postgres cache
	if userProfileID, ok := session.Values["user_profile_id"].(string); ok && userProfileID != "" {
		if err := s.userProfileService.DeleteProfile(r.Context(), userProfileID); err != nil {
			log.Printf("WARNING: Failed to delete user profile from cache: %v", err)
		}
	}

	// Clear session
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to clear session", http.StatusInternalServerError)
		return
	}

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
		json.NewEncoder(w).Encode(AuthStatusResponse{Authenticated: false})
		return
	}

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

	// Get user profile from Postgres cache; if the cache TTL has lapsed,
	// rebuild it from the session's Entra claims (no external call needed).
	var userProfile *UserProfileResponse
	if userProfileID, ok := session.Values["user_profile_id"].(string); ok && userProfileID != "" {
		profile, err := s.userProfileService.GetProfile(r.Context(), userProfileID)
		if profile == nil || err != nil {
			if err != nil {
				log.Printf("WARNING: Failed to get cached user profile: %v, rebuilding from session", err)
			}
			rebuilt := s.profileFromSession(session)
			if rebuilt != nil {
				if cacheErr := s.userProfileService.SetProfile(r.Context(), rebuilt); cacheErr != nil {
					log.Printf("WARNING: Failed to re-cache user profile: %v", cacheErr)
				}
				profile = rebuilt
			}
		}

		if profile != nil {
			primaryEmail := ""
			for _, email := range profile.Emails {
				if email.Primary {
					primaryEmail = email.Value
					break
				}
			}

			groups := make([]UserProfileGroupResponse, len(profile.Groups))
			for i, g := range profile.Groups {
				groups[i] = UserProfileGroupResponse{Value: g.Value, Display: g.Display, Type: g.Type}
			}

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
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthStatusResponse{
		Authenticated: true,
		Environment:   environment,
		UserContext:   userContext,
		UserProfile:   userProfile,
	})
}

// profileFromSession rebuilds the cached profile from session claims.
func (s *Server) profileFromSession(session *sessions.Session) *infor.CombinedUserProfile {
	oid := getSessionString(session, "user_profile_id")
	if oid == "" {
		return nil
	}
	var roles []string
	if raw := getSessionString(session, "user_roles"); raw != "" {
		roles = strings.Split(raw, ",")
	}
	return entraProfile(&auth.EntraUser{
		ObjectID: oid,
		Email:    getSessionString(session, "user_email"),
		Name:     getSessionString(session, "user_full_name"),
		Roles:    roles,
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

	// 3. Use bulk API to prime ALL company-scoped entities in single call
	if err := repo.RefreshAllContextBulk(ctx, companies); err != nil {
		fmt.Printf("ERROR: Bulk context refresh failed: %v\n", err)
		return
	}

	fmt.Printf("Context cache priming completed for %s using bulk API calls\n", environment)
}

// handleRefreshProfile rebuilds the cached profile from the session's Entra
// claims. Identity lives in Entra ID; there is no external profile API to
// re-fetch under the service-account model.
func (s *Server) handleRefreshProfile(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "m3-session")

	authenticated, ok := session.Values["authenticated"].(bool)
	if !ok || !authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	profile := s.profileFromSession(session)
	if profile == nil {
		http.Error(w, "No profile in session", http.StatusInternalServerError)
		return
	}

	if err := s.userProfileService.SetProfile(r.Context(), profile); err != nil {
		log.Printf("ERROR: Failed to cache profile in Postgres: %v", err)
		http.Error(w, "Failed to cache profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}
