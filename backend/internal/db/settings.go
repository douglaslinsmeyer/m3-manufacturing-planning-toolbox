package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// UserSettings represents user-specific preferences
type UserSettings struct {
	UserID              string
	Environment         string // M3 environment (TRN or PRD)
	DefaultWarehouse    sql.NullString
	DefaultFacility     sql.NullString
	DefaultDivision     sql.NullString
	DefaultCompany      sql.NullString
	ItemsPerPage        int32
	Theme               string
	DateFormat          string
	TimeFormat          string
	EnableNotifications bool
	NotificationSound   bool
	Preferences         json.RawMessage
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// SystemSetting represents a system-wide configuration setting
type SystemSetting struct {
	ID             int32
	Environment    string // M3 environment (TRN or PRD)
	SettingKey     string
	SettingValue   string
	SettingType    string
	Description    sql.NullString
	Category       string
	Constraints    json.RawMessage
	LastModifiedBy sql.NullString
	LastModifiedAt time.Time
	CreatedAt      time.Time
}

// GetUserSettings retrieves user settings for a specific environment
func (q *Queries) GetUserSettings(ctx context.Context, environment, userID string) (*UserSettings, error) {
	query := `
		SELECT environment, user_id, default_warehouse, default_facility, default_division, default_company,
		       items_per_page, theme, date_format, time_format,
		       enable_notifications, notification_sound, preferences,
		       created_at, updated_at
		FROM user_settings
		WHERE environment = $1 AND user_id = $2
	`
	var s UserSettings
	err := q.db.QueryRowContext(ctx, query, environment, userID).Scan(
		&s.Environment, &s.UserID, &s.DefaultWarehouse, &s.DefaultFacility, &s.DefaultDivision, &s.DefaultCompany,
		&s.ItemsPerPage, &s.Theme, &s.DateFormat, &s.TimeFormat,
		&s.EnableNotifications, &s.NotificationSound, &s.Preferences,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // No settings exist yet
	}
	return &s, err
}

// UpsertUserSettingsParams contains parameters for upserting user settings
type UpsertUserSettingsParams struct {
	Environment         string
	UserID              string
	DefaultWarehouse    sql.NullString
	DefaultFacility     sql.NullString
	DefaultDivision     sql.NullString
	DefaultCompany      sql.NullString
	ItemsPerPage        int32
	Theme               string
	DateFormat          string
	TimeFormat          string
	EnableNotifications bool
	NotificationSound   bool
	Preferences         json.RawMessage
}

// UpsertUserSettings creates or updates user settings for a specific environment
func (q *Queries) UpsertUserSettings(ctx context.Context, params UpsertUserSettingsParams) error {
	query := `
		INSERT INTO user_settings (
			environment, user_id, default_warehouse, default_facility, default_division, default_company,
			items_per_page, theme, date_format, time_format,
			enable_notifications, notification_sound, preferences, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW())
		ON CONFLICT (environment, user_id) DO UPDATE SET
			default_warehouse = EXCLUDED.default_warehouse,
			default_facility = EXCLUDED.default_facility,
			default_division = EXCLUDED.default_division,
			default_company = EXCLUDED.default_company,
			items_per_page = EXCLUDED.items_per_page,
			theme = EXCLUDED.theme,
			date_format = EXCLUDED.date_format,
			time_format = EXCLUDED.time_format,
			enable_notifications = EXCLUDED.enable_notifications,
			notification_sound = EXCLUDED.notification_sound,
			preferences = EXCLUDED.preferences,
			updated_at = NOW()
	`
	_, err := q.db.ExecContext(ctx, query,
		params.Environment, params.UserID, params.DefaultWarehouse, params.DefaultFacility,
		params.DefaultDivision, params.DefaultCompany,
		params.ItemsPerPage, params.Theme, params.DateFormat, params.TimeFormat,
		params.EnableNotifications, params.NotificationSound, params.Preferences,
	)
	return err
}

// GetSystemSettings retrieves all system settings for a specific environment
func (q *Queries) GetSystemSettings(ctx context.Context, environment string) ([]SystemSetting, error) {
	query := `
		SELECT id, environment, setting_key, setting_value, setting_type, description, category,
		       constraints, last_modified_by, last_modified_at, created_at
		FROM system_settings
		WHERE environment = $1
		ORDER BY category, setting_key
	`
	rows, err := q.db.QueryContext(ctx, query, environment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []SystemSetting
	for rows.Next() {
		var s SystemSetting
		if err := rows.Scan(
			&s.ID, &s.Environment, &s.SettingKey, &s.SettingValue, &s.SettingType, &s.Description, &s.Category,
			&s.Constraints, &s.LastModifiedBy, &s.LastModifiedAt, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, rows.Err()
}

// GetSystemSettingsByCategory retrieves system settings for a specific category and environment
func (q *Queries) GetSystemSettingsByCategory(ctx context.Context, environment, category string) ([]SystemSetting, error) {
	query := `
		SELECT id, environment, setting_key, setting_value, setting_type, description, category,
		       constraints, last_modified_by, last_modified_at, created_at
		FROM system_settings
		WHERE environment = $1 AND category = $2
		ORDER BY setting_key
	`
	rows, err := q.db.QueryContext(ctx, query, environment, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []SystemSetting
	for rows.Next() {
		var s SystemSetting
		if err := rows.Scan(
			&s.ID, &s.Environment, &s.SettingKey, &s.SettingValue, &s.SettingType, &s.Description, &s.Category,
			&s.Constraints, &s.LastModifiedBy, &s.LastModifiedAt, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, rows.Err()
}

// UpdateSystemSettingParams contains parameters for updating a system setting
type UpdateSystemSettingParams struct {
	Environment    string
	SettingKey     string
	SettingValue   string
	LastModifiedBy string
}

// UpdateSystemSetting updates a single system setting for a specific environment
func (q *Queries) UpdateSystemSetting(ctx context.Context, params UpdateSystemSettingParams) error {
	query := `
		UPDATE system_settings
		SET setting_value = $1,
		    last_modified_by = $2,
		    last_modified_at = NOW()
		WHERE environment = $3 AND setting_key = $4
	`
	_, err := q.db.ExecContext(ctx, query, params.SettingValue, params.LastModifiedBy, params.Environment, params.SettingKey)
	return err
}
