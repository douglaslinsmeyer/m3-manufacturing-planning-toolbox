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

// GetUserSettings retrieves user settings, returning defaults if none exist
func (s *SettingsService) GetUserSettings(ctx context.Context, userID string) (*db.UserSettings, error) {
	settings, err := s.queries.GetUserSettings(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Return default settings if none exist
	if settings == nil {
		settings = &db.UserSettings{
			UserID:              userID,
			ItemsPerPage:        20,
			Theme:               "light",
			DateFormat:          "YYYY-MM-DD",
			TimeFormat:          "24h",
			EnableNotifications: true,
			NotificationSound:   false,
			Preferences:         []byte("{}"),
		}
	}

	return settings, nil
}

// UpdateUserSettings updates user settings and logs the change
func (s *SettingsService) UpdateUserSettings(
	ctx context.Context,
	userID string,
	params db.UpsertUserSettingsParams,
	modifiedBy string,
) error {
	// Perform update
	if err := s.queries.UpsertUserSettings(ctx, params); err != nil {
		return err
	}

	// Log audit trail
	return s.auditService.Log(ctx, AuditParams{
		EntityType: "user_settings",
		EntityID:   userID,
		Operation:  "update",
		UserID:     modifiedBy,
		Metadata: map[string]interface{}{
			"settings_updated": true,
		},
	})
}

// GetSystemSettings retrieves all system settings
func (s *SettingsService) GetSystemSettings(ctx context.Context) ([]db.SystemSetting, error) {
	return s.queries.GetSystemSettings(ctx)
}

// UpdateSystemSettings updates multiple system settings (admin only)
func (s *SettingsService) UpdateSystemSettings(
	ctx context.Context,
	updates map[string]string,
	modifiedBy string,
) error {
	// Validate and update each setting
	for key, value := range updates {
		if err := s.queries.UpdateSystemSetting(ctx, db.UpdateSystemSettingParams{
			SettingKey:     key,
			SettingValue:   value,
			LastModifiedBy: modifiedBy,
		}); err != nil {
			return fmt.Errorf("failed to update setting %s: %w", key, err)
		}
	}

	// Log audit trail
	return s.auditService.Log(ctx, AuditParams{
		EntityType: "system_settings",
		Operation:  "bulk_update",
		UserID:     modifiedBy,
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
