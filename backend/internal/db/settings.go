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

// GetUserSettings retrieves user settings
func (q *Queries) GetUserSettings(ctx context.Context, userID string) (*UserSettings, error) {
	query := `
		SELECT user_id, default_warehouse, default_facility, default_division, default_company,
		       items_per_page, theme, date_format, time_format,
		       enable_notifications, notification_sound, preferences,
		       created_at, updated_at
		FROM user_settings
		WHERE user_id = $1
	`
	var s UserSettings
	err := q.db.QueryRowContext(ctx, query, userID).Scan(
		&s.UserID, &s.DefaultWarehouse, &s.DefaultFacility, &s.DefaultDivision, &s.DefaultCompany,
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

// UpsertUserSettings creates or updates user settings
func (q *Queries) UpsertUserSettings(ctx context.Context, params UpsertUserSettingsParams) error {
	query := `
		INSERT INTO user_settings (
			user_id, default_warehouse, default_facility, default_division, default_company,
			items_per_page, theme, date_format, time_format,
			enable_notifications, notification_sound, preferences, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
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
		params.UserID, params.DefaultWarehouse, params.DefaultFacility,
		params.DefaultDivision, params.DefaultCompany,
		params.ItemsPerPage, params.Theme, params.DateFormat, params.TimeFormat,
		params.EnableNotifications, params.NotificationSound, params.Preferences,
	)
	return err
}

// GetSystemSettings retrieves all system settings
func (q *Queries) GetSystemSettings(ctx context.Context) ([]SystemSetting, error) {
	query := `
		SELECT id, setting_key, setting_value, setting_type, description, category,
		       constraints, last_modified_by, last_modified_at, created_at
		FROM system_settings
		ORDER BY category, setting_key
	`
	rows, err := q.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []SystemSetting
	for rows.Next() {
		var s SystemSetting
		if err := rows.Scan(
			&s.ID, &s.SettingKey, &s.SettingValue, &s.SettingType, &s.Description, &s.Category,
			&s.Constraints, &s.LastModifiedBy, &s.LastModifiedAt, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, rows.Err()
}

// GetSystemSettingsByCategory retrieves system settings for a specific category
func (q *Queries) GetSystemSettingsByCategory(ctx context.Context, category string) ([]SystemSetting, error) {
	query := `
		SELECT id, setting_key, setting_value, setting_type, description, category,
		       constraints, last_modified_by, last_modified_at, created_at
		FROM system_settings
		WHERE category = $1
		ORDER BY setting_key
	`
	rows, err := q.db.QueryContext(ctx, query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []SystemSetting
	for rows.Next() {
		var s SystemSetting
		if err := rows.Scan(
			&s.ID, &s.SettingKey, &s.SettingValue, &s.SettingType, &s.Description, &s.Category,
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
	SettingKey     string
	SettingValue   string
	LastModifiedBy string
}

// UpdateSystemSetting updates a single system setting
func (q *Queries) UpdateSystemSetting(ctx context.Context, params UpdateSystemSettingParams) error {
	query := `
		UPDATE system_settings
		SET setting_value = $1,
		    last_modified_by = $2,
		    last_modified_at = NOW()
		WHERE setting_key = $3
	`
	_, err := q.db.ExecContext(ctx, query, params.SettingValue, params.LastModifiedBy, params.SettingKey)
	return err
}
