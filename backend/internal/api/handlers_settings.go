package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pinggolf/m3-planning-tools/internal/db"
)

// UserSettingsResponse represents the API response for user settings
type UserSettingsResponse struct {
	UserID              string                 `json:"userId"`
	DefaultWarehouse    string                 `json:"defaultWarehouse,omitempty"`
	DefaultFacility     string                 `json:"defaultFacility,omitempty"`
	DefaultDivision     string                 `json:"defaultDivision,omitempty"`
	DefaultCompany      string                 `json:"defaultCompany,omitempty"`
	ItemsPerPage        int32                  `json:"itemsPerPage"`
	Theme               string                 `json:"theme"`
	DateFormat          string                 `json:"dateFormat"`
	TimeFormat          string                 `json:"timeFormat"`
	EnableNotifications bool                   `json:"enableNotifications"`
	NotificationSound   bool                   `json:"notificationSound"`
	Preferences         map[string]interface{} `json:"preferences"`
}

// SystemSettingResponse represents the API response for a system setting
type SystemSettingResponse struct {
	Key         string                 `json:"key"`
	Value       string                 `json:"value"`
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Category    string                 `json:"category"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
}

// SystemSettingsGroupedResponse groups settings by category
type SystemSettingsGroupedResponse struct {
	Categories map[string][]SystemSettingResponse `json:"categories"`
}

// handleGetUserSettings retrieves the current user's settings
func (s *Server) handleGetUserSettings(w http.ResponseWriter, r *http.Request) {
	// Get user ID from session/profile
	userID, err := s.getUserIDFromSession(r)
	if err != nil {
		log.Printf("ERROR: Failed to get user ID from session: %v", err)
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	log.Printf("DEBUG: Getting user settings for user ID: %s", userID)

	// Get settings from service
	settings, err := s.settingsService.GetUserSettings(r.Context(), userID)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve user settings: %v", err)
		http.Error(w, "Failed to retrieve settings", http.StatusInternalServerError)
		return
	}

	log.Printf("DEBUG: Successfully retrieved user settings for: %s", userID)

	// Convert to response format
	var preferences map[string]interface{}
	if len(settings.Preferences) > 0 {
		if err := json.Unmarshal(settings.Preferences, &preferences); err != nil {
			preferences = make(map[string]interface{})
		}
	} else {
		preferences = make(map[string]interface{})
	}

	response := UserSettingsResponse{
		UserID:              settings.UserID,
		DefaultWarehouse:    settings.DefaultWarehouse.String,
		DefaultFacility:     settings.DefaultFacility.String,
		DefaultDivision:     settings.DefaultDivision.String,
		DefaultCompany:      settings.DefaultCompany.String,
		ItemsPerPage:        settings.ItemsPerPage,
		Theme:               settings.Theme,
		DateFormat:          settings.DateFormat,
		TimeFormat:          settings.TimeFormat,
		EnableNotifications: settings.EnableNotifications,
		NotificationSound:   settings.NotificationSound,
		Preferences:         preferences,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateUserSettingsRequest represents the request body for updating user settings
type UpdateUserSettingsRequest struct {
	DefaultWarehouse    string                 `json:"defaultWarehouse,omitempty"`
	DefaultFacility     string                 `json:"defaultFacility,omitempty"`
	DefaultDivision     string                 `json:"defaultDivision,omitempty"`
	DefaultCompany      string                 `json:"defaultCompany,omitempty"`
	ItemsPerPage        int32                  `json:"itemsPerPage"`
	Theme               string                 `json:"theme"`
	DateFormat          string                 `json:"dateFormat"`
	TimeFormat          string                 `json:"timeFormat"`
	EnableNotifications bool                   `json:"enableNotifications"`
	NotificationSound   bool                   `json:"notificationSound"`
	Preferences         map[string]interface{} `json:"preferences"`
}

// handleUpdateUserSettings updates the current user's settings
func (s *Server) handleUpdateUserSettings(w http.ResponseWriter, r *http.Request) {
	// Get user ID from session/profile
	userID, err := s.getUserIDFromSession(r)
	if err != nil {
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	// Parse request
	var req UpdateUserSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate theme
	if req.Theme != "light" && req.Theme != "dark" && req.Theme != "auto" {
		http.Error(w, "Invalid theme value", http.StatusBadRequest)
		return
	}

	// Validate items per page
	if req.ItemsPerPage < 1 || req.ItemsPerPage > 200 {
		http.Error(w, "Items per page must be between 1 and 200", http.StatusBadRequest)
		return
	}

	// Validate time format
	if req.TimeFormat != "12h" && req.TimeFormat != "24h" {
		http.Error(w, "Invalid time format", http.StatusBadRequest)
		return
	}

	// Marshal preferences to JSON
	var preferencesJSON []byte
	if req.Preferences != nil {
		preferencesJSON, err = json.Marshal(req.Preferences)
		if err != nil {
			http.Error(w, "Invalid preferences format", http.StatusBadRequest)
			return
		}
	} else {
		preferencesJSON = []byte("{}")
	}

	// Update settings
	params := db.UpsertUserSettingsParams{
		UserID:              userID,
		DefaultWarehouse:    sql.NullString{String: req.DefaultWarehouse, Valid: req.DefaultWarehouse != ""},
		DefaultFacility:     sql.NullString{String: req.DefaultFacility, Valid: req.DefaultFacility != ""},
		DefaultDivision:     sql.NullString{String: req.DefaultDivision, Valid: req.DefaultDivision != ""},
		DefaultCompany:      sql.NullString{String: req.DefaultCompany, Valid: req.DefaultCompany != ""},
		ItemsPerPage:        req.ItemsPerPage,
		Theme:               req.Theme,
		DateFormat:          req.DateFormat,
		TimeFormat:          req.TimeFormat,
		EnableNotifications: req.EnableNotifications,
		NotificationSound:   req.NotificationSound,
		Preferences:         preferencesJSON,
	}

	if err := s.settingsService.UpdateUserSettings(r.Context(), userID, params, userID); err != nil {
		http.Error(w, "Failed to update settings", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Settings updated successfully"})
}

// handleGetSystemSettings retrieves all system settings (admin only)
func (s *Server) handleGetSystemSettings(w http.ResponseWriter, r *http.Request) {
	// Admin check handled by middleware
	log.Printf("DEBUG: handleGetSystemSettings called")

	// Get settings from service
	settings, err := s.settingsService.GetSystemSettings(r.Context())
	if err != nil {
		log.Printf("ERROR: Failed to retrieve system settings: %v", err)
		http.Error(w, fmt.Sprintf("Failed to retrieve system settings: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("DEBUG: Retrieved %d system settings", len(settings))

	// Group by category
	grouped := make(map[string][]SystemSettingResponse)
	for _, setting := range settings {
		var constraints map[string]interface{}
		if len(setting.Constraints) > 0 {
			json.Unmarshal(setting.Constraints, &constraints)
		}

		response := SystemSettingResponse{
			Key:         setting.SettingKey,
			Value:       setting.SettingValue,
			Type:        setting.SettingType,
			Description: setting.Description.String,
			Category:    setting.Category,
			Constraints: constraints,
		}

		grouped[setting.Category] = append(grouped[setting.Category], response)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SystemSettingsGroupedResponse{Categories: grouped})
}

// UpdateSystemSettingsRequest represents the request body for updating system settings
type UpdateSystemSettingsRequest struct {
	Settings map[string]string `json:"settings"`
}

// handleUpdateSystemSettings updates system settings (admin only)
func (s *Server) handleUpdateSystemSettings(w http.ResponseWriter, r *http.Request) {
	// Admin check handled by middleware

	// Get user ID from session/profile
	userID, err := s.getUserIDFromSession(r)
	if err != nil {
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	// Parse request
	var req UpdateSystemSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Settings) == 0 {
		http.Error(w, "No settings provided", http.StatusBadRequest)
		return
	}

	// Update settings
	if err := s.settingsService.UpdateSystemSettings(r.Context(), req.Settings, userID); err != nil {
		http.Error(w, "Failed to update system settings", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "System settings updated successfully"})
}

// getUserIDFromSession extracts user ID from session
func (s *Server) getUserIDFromSession(r *http.Request) (string, error) {
	session, _ := s.sessionStore.Get(r, "m3-session")

	userProfileID, ok := session.Values["user_profile_id"].(string)
	if !ok || userProfileID == "" {
		return "", fmt.Errorf("no user profile ID in session")
	}

	return userProfileID, nil
}
