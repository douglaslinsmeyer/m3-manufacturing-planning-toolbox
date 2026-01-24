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
	repository *ContextRepository
}

// NewContextService creates a new context service
func NewContextService(repository *ContextRepository) *ContextService {
	return &ContextService{
		repository: repository,
	}
}

// EffectiveContext represents the calculated effective context
type EffectiveContext struct {
	Company   string
	Division  string
	Facility  string
	Warehouse string
}

// LoadUserDefaults fetches user defaults from M3 and stores in session
func (s *ContextService) LoadUserDefaults(ctx context.Context, session *sessions.Session, m3Client *m3api.Client) error {
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

	// Store in session as user defaults
	session.Values["user_company"] = userInfo.Company
	session.Values["user_division"] = userInfo.Division
	session.Values["user_facility"] = userInfo.Facility
	session.Values["user_warehouse"] = userInfo.Warehouse
	session.Values["user_full_name"] = userInfo.FullName

	// Debug: Verify what was stored
	fmt.Printf("DEBUG LoadUserDefaults: Stored in session:\n")
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
