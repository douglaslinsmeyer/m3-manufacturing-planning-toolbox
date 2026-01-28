package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pinggolf/m3-planning-tools/internal/db"
)

// SettingsService manages user and system settings
type SettingsService struct {
	queries      *db.Queries
	auditService *AuditService
}

// NewSettingsService creates a new settings service
func NewSettingsService(queries *db.Queries, auditService *AuditService) *SettingsService {
	return &SettingsService{
		queries:      queries,
		auditService: auditService,
	}
}

// GetUserSettings retrieves user settings for a specific environment, returning empty settings if none exist
func (s *SettingsService) GetUserSettings(ctx context.Context, environment, userID string) (*db.UserSettings, error) {
	settings, err := s.queries.GetUserSettings(ctx, environment, userID)
	if err != nil {
		return nil, err
	}

	// Return empty settings if none exist (all defaults will be null/unset)
	if settings == nil {
		settings = &db.UserSettings{
			Environment: environment,
			UserID:      userID,
		}
	}

	return settings, nil
}

// UpdateUserSettings updates user settings and logs the change
func (s *SettingsService) UpdateUserSettings(
	ctx context.Context,
	environment string,
	userID string,
	params db.UpsertUserSettingsParams,
	modifiedBy string,
) error {
	// Ensure environment is set in params
	params.Environment = environment

	// Perform update
	if err := s.queries.UpsertUserSettings(ctx, params); err != nil {
		return err
	}

	// Log audit trail
	return s.auditService.Log(ctx, AuditParams{
		Environment: environment,
		EntityType:  "user_settings",
		EntityID:    userID,
		Operation:   "update",
		UserID:      modifiedBy,
		Metadata: map[string]interface{}{
			"settings_updated": true,
		},
	})
}

// GetSystemSettings retrieves all system settings for a specific environment
func (s *SettingsService) GetSystemSettings(ctx context.Context, environment string) ([]db.SystemSetting, error) {
	return s.queries.GetSystemSettings(ctx, environment)
}

// UpdateSystemSettings updates multiple system settings (admin only) for a specific environment
func (s *SettingsService) UpdateSystemSettings(
	ctx context.Context,
	environment string,
	updates map[string]string,
	modifiedBy string,
) error {
	// Validate and update each setting
	for key, value := range updates {
		if err := s.queries.UpdateSystemSetting(ctx, db.UpdateSystemSettingParams{
			Environment:    environment,
			SettingKey:     key,
			SettingValue:   value,
			LastModifiedBy: modifiedBy,
		}); err != nil {
			return fmt.Errorf("failed to update setting %s: %w", key, err)
		}
	}

	// Log audit trail
	return s.auditService.Log(ctx, AuditParams{
		Environment: environment,
		EntityType:  "system_settings",
		Operation:   "bulk_update",
		UserID:      modifiedBy,
		Metadata: map[string]interface{}{
			"settings_count": len(updates),
			"settings_keys":  getKeys(updates),
		},
	})
}

// Helper function to extract map keys
func getKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ParseSettingValue parses a system setting value based on its type
func ParseSettingValue(setting db.SystemSetting) (interface{}, error) {
	switch setting.SettingType {
	case "string":
		return setting.SettingValue, nil
	case "integer":
		return strconv.ParseInt(setting.SettingValue, 10, 64)
	case "float":
		return strconv.ParseFloat(setting.SettingValue, 64)
	case "boolean":
		return strconv.ParseBool(setting.SettingValue)
	case "json":
		var result interface{}
		if err := json.Unmarshal([]byte(setting.SettingValue), &result); err != nil {
			return nil, err
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unknown setting type: %s", setting.SettingType)
	}
}
