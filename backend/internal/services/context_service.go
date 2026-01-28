package services

import (
	"context"
	"fmt"

	"github.com/gorilla/sessions"
	"github.com/pinggolf/m3-planning-tools/internal/compass"
	"github.com/pinggolf/m3-planning-tools/internal/m3api"
)

// ContextService manages user context (defaults and temporary overrides)
type ContextService struct {
	repository         *ContextRepository
	userProfileService *UserProfileService
	settingsService    *SettingsService
}

// NewContextService creates a new context service
func NewContextService(repository *ContextRepository) *ContextService {
	return &ContextService{
		repository: repository,
	}
}

// SetUserProfileService sets the user profile service (for accessing M3 defaults from profile cache)
func (s *ContextService) SetUserProfileService(userProfileService *UserProfileService) {
	s.userProfileService = userProfileService
}

// SetSettingsService sets the settings service (for accessing user custom defaults)
func (s *ContextService) SetSettingsService(settingsService *SettingsService) {
	s.settingsService = settingsService
}

// EffectiveContext represents the calculated effective context
type EffectiveContext struct {
	Company   string
	Division  string
	Facility  string
	Warehouse string
}

// LoadUserDefaults fetches user defaults and stores in session
// Priority order: user_settings (custom) → profile cache (M3) → M3 API
func (s *ContextService) LoadUserDefaults(ctx context.Context, session *sessions.Session, m3Client *m3api.Client) error {
	// Initialize with empty defaults that will be populated from various sources
	var company, division, facility, warehouse, fullName string

	// Priority 1: Check user_settings for custom defaults
	if s.settingsService != nil {
		if environment, ok := session.Values["environment"].(string); ok {
			if userID, ok := session.Values["user_profile_id"].(string); ok && userID != "" {
				if userSettings, err := s.settingsService.GetUserSettings(ctx, environment, userID); err == nil && userSettings != nil {
					fmt.Printf("INFO: Checking user_settings for custom defaults\n")

					// Apply custom defaults if they exist (not null)
					customsFound := false
					if userSettings.DefaultCompany.Valid && userSettings.DefaultCompany.String != "" {
						company = userSettings.DefaultCompany.String
						customsFound = true
						fmt.Printf("DEBUG LoadUserDefaults: Using custom company: %s\n", company)
					}
					if userSettings.DefaultDivision.Valid && userSettings.DefaultDivision.String != "" {
						division = userSettings.DefaultDivision.String
						customsFound = true
						fmt.Printf("DEBUG LoadUserDefaults: Using custom division: %s\n", division)
					}
					if userSettings.DefaultFacility.Valid && userSettings.DefaultFacility.String != "" {
						facility = userSettings.DefaultFacility.String
						customsFound = true
						fmt.Printf("DEBUG LoadUserDefaults: Using custom facility: %s\n", facility)
					}
					if userSettings.DefaultWarehouse.Valid && userSettings.DefaultWarehouse.String != "" {
						warehouse = userSettings.DefaultWarehouse.String
						customsFound = true
						fmt.Printf("DEBUG LoadUserDefaults: Using custom warehouse: %s\n", warehouse)
					}

					if customsFound {
						fmt.Printf("INFO: Applied custom defaults from user_settings\n")
					}
				}
			}
		}
	}

	// Priority 2: Fill in any missing defaults from M3 profile cache
	if s.userProfileService != nil {
		if userProfileID, ok := session.Values["user_profile_id"].(string); ok && userProfileID != "" {
			if profile, err := s.userProfileService.GetProfile(ctx, userProfileID); err == nil && profile != nil {
				if profile.M3Info != nil {
					fmt.Printf("INFO: Filling missing defaults from profile cache\n")

					if company == "" {
						company = profile.M3Info.DefaultCompany
					}
					if division == "" {
						division = profile.M3Info.DefaultDivision
					}
					if facility == "" {
						facility = profile.M3Info.DefaultFacility
					}
					if warehouse == "" {
						warehouse = profile.M3Info.DefaultWarehouse
					}
					fullName = profile.M3Info.FullName

					fmt.Printf("DEBUG LoadUserDefaults: After M3 cache - Company: %s, Div: %s, Fac: %s, Whse: %s\n",
						company, division, facility, warehouse)
				}
			}
		}
	}

	// Priority 3: Fallback to M3 API call if any fields still missing
	if company == "" || division == "" || facility == "" || warehouse == "" || fullName == "" {
		fmt.Printf("INFO: Loading missing defaults from M3 API\n")
		userInfo, err := compass.GetUserInfo(ctx, m3Client)
		if err != nil {
			return err
		}

		// Debug: Log what we received from M3
		fmt.Printf("DEBUG LoadUserDefaults: Received from M3 GetUserInfo:\n")
		fmt.Printf("  Company: '%s'\n", userInfo.Company)
		fmt.Printf("  Division: '%s'\n", userInfo.Division)
		fmt.Printf("  Facility: '%s'\n", userInfo.Facility)
		fmt.Printf("  Warehouse: '%s'\n", userInfo.Warehouse)
		fmt.Printf("  FullName: '%s'\n", userInfo.FullName)

		// Fill in missing values
		if company == "" {
			company = userInfo.Company
		}
		if division == "" {
			division = userInfo.Division
		}
		if facility == "" {
			facility = userInfo.Facility
		}
		if warehouse == "" {
			warehouse = userInfo.Warehouse
		}
		if fullName == "" {
			fullName = userInfo.FullName
		}
	}

	// Store final effective defaults in session
	session.Values["user_company"] = company
	session.Values["user_division"] = division
	session.Values["user_facility"] = facility
	session.Values["user_warehouse"] = warehouse
	session.Values["user_full_name"] = fullName

	// Debug: Verify what was stored
	fmt.Printf("DEBUG LoadUserDefaults: Final session values:\n")
	fmt.Printf("  user_company: '%v'\n", session.Values["user_company"])
	fmt.Printf("  user_division: '%v'\n", session.Values["user_division"])
	fmt.Printf("  user_facility: '%v'\n", session.Values["user_facility"])
	fmt.Printf("  user_warehouse: '%v'\n", session.Values["user_warehouse"])

	return nil
}

// GetEffectiveContext calculates the effective context (temporary overrides → user defaults)
func (s *ContextService) GetEffectiveContext(session *sessions.Session) EffectiveContext {
	// Debug: Log session contents
	fmt.Printf("DEBUG GetEffectiveContext: Session values:\n")
	for k, v := range session.Values {
		fmt.Printf("  %v: %v\n", k, v)
	}

	effective := EffectiveContext{}

	// Company: temporary override or user default
	if temp, ok := session.Values["temp_company"].(string); ok && temp != "" {
		effective.Company = temp
		fmt.Printf("DEBUG GetEffectiveContext: Using temp_company: '%s'\n", temp)
	} else if user, ok := session.Values["user_company"].(string); ok {
		effective.Company = user
		fmt.Printf("DEBUG GetEffectiveContext: Using user_company: '%s'\n", user)
	} else {
		fmt.Printf("DEBUG GetEffectiveContext: No company found in session\n")
	}

	// Division: temporary override or user default
	if temp, ok := session.Values["temp_division"].(string); ok && temp != "" {
		effective.Division = temp
	} else if user, ok := session.Values["user_division"].(string); ok {
		effective.Division = user
	}

	// Facility: temporary override or user default
	if temp, ok := session.Values["temp_facility"].(string); ok && temp != "" {
		effective.Facility = temp
	} else if user, ok := session.Values["user_facility"].(string); ok {
		effective.Facility = user
	}

	// Warehouse: temporary override or user default
	if temp, ok := session.Values["temp_warehouse"].(string); ok && temp != "" {
		effective.Warehouse = temp
	} else if user, ok := session.Values["user_warehouse"].(string); ok {
		effective.Warehouse = user
	}

	fmt.Printf("DEBUG GetEffectiveContext: Returning effective context: %+v\n", effective)
	return effective
}

// GetUserDefaults returns the user's default context from session
func (s *ContextService) GetUserDefaults(session *sessions.Session) EffectiveContext {
	defaults := EffectiveContext{}

	if company, ok := session.Values["user_company"].(string); ok {
		defaults.Company = company
	}
	if division, ok := session.Values["user_division"].(string); ok {
		defaults.Division = division
	}
	if facility, ok := session.Values["user_facility"].(string); ok {
		defaults.Facility = facility
	}
	if warehouse, ok := session.Values["user_warehouse"].(string); ok {
		defaults.Warehouse = warehouse
	}

	return defaults
}

// SetTemporaryOverride updates a temporary context override
func (s *ContextService) SetTemporaryOverride(session *sessions.Session, field, value string) {
	key := "temp_" + field
	if value == "" {
		// Empty value means clear the override
		delete(session.Values, key)
	} else {
		session.Values[key] = value
	}
}

// ClearTemporaryOverrides removes all temporary overrides
func (s *ContextService) ClearTemporaryOverrides(session *sessions.Session) {
	delete(session.Values, "temp_company")
	delete(session.Values, "temp_division")
	delete(session.Values, "temp_facility")
	delete(session.Values, "temp_warehouse")
}

// HasTemporaryOverrides checks if any temporary overrides exist
func (s *ContextService) HasTemporaryOverrides(session *sessions.Session) bool {
	_, hasCompany := session.Values["temp_company"]
	_, hasDivision := session.Values["temp_division"]
	_, hasFacility := session.Values["temp_facility"]
	_, hasWarehouse := session.Values["temp_warehouse"]
	return hasCompany || hasDivision || hasFacility || hasWarehouse
}

// GetUserFullName returns the user's full name from session
func (s *ContextService) GetUserFullName(session *sessions.Session) string {
	if name, ok := session.Values["user_full_name"].(string); ok {
		return name
	}
	return ""
}
