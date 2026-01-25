package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pinggolf/m3-planning-tools/internal/infor"
)

// UserProfileService manages user profile caching in Postgres
type UserProfileService struct {
	db  *sql.DB
	ttl time.Duration
}

// NewUserProfileService creates a new user profile service
func NewUserProfileService(db *sql.DB) *UserProfileService {
	return &UserProfileService{
		db:  db,
		ttl: 15 * time.Minute, // Profile cache TTL: 15 minutes
	}
}

// GetProfile retrieves a cached combined user profile from Postgres
// Returns nil if not found or expired
func (s *UserProfileService) GetProfile(ctx context.Context, userID string) (*infor.CombinedUserProfile, error) {
	var profileJSON []byte
	var expiresAt time.Time

	query := `
		SELECT profile_data, expires_at
		FROM user_profiles
		WHERE user_id = $1 AND expires_at > NOW()
	`

	err := s.db.QueryRowContext(ctx, query, userID).Scan(&profileJSON, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil // Not found or expired
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query profile: %w", err)
	}

	var profile infor.CombinedUserProfile
	if err := json.Unmarshal(profileJSON, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	return &profile, nil
}

// SetProfile stores a combined user profile in Postgres with TTL
func (s *UserProfileService) SetProfile(ctx context.Context, profile *infor.CombinedUserProfile) error {
	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	expiresAt := time.Now().Add(s.ttl)

	query := `
		INSERT INTO user_profiles (user_id, profile_data, fetched_at, expires_at, updated_at)
		VALUES ($1, $2, NOW(), $3, NOW())
		ON CONFLICT (user_id)
		DO UPDATE SET
			profile_data = EXCLUDED.profile_data,
			fetched_at = NOW(),
			expires_at = EXCLUDED.expires_at,
			updated_at = NOW()
	`

	_, err = s.db.ExecContext(ctx, query, profile.ID, profileJSON, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to store profile: %w", err)
	}

	return nil
}

// DeleteProfile removes a user profile from cache
func (s *UserProfileService) DeleteProfile(ctx context.Context, userID string) error {
	query := `DELETE FROM user_profiles WHERE user_id = $1`
	_, err := s.db.ExecContext(ctx, query, userID)
	return err
}

// CleanupExpired removes expired profiles from the cache
// Should be called periodically (e.g., daily cron job)
func (s *UserProfileService) CleanupExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM user_profiles WHERE expires_at < NOW()`
	result, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// HasRole checks if a user has a specific role/group
// Useful for authorization checks
func (s *UserProfileService) HasRole(ctx context.Context, userID string, roleDisplay string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_profiles
			WHERE user_id = $1
			AND expires_at > NOW()
			AND profile_data @> jsonb_build_object('groups', jsonb_build_array(jsonb_build_object('display', $2)))
		)
	`

	var hasRole bool
	err := s.db.QueryRowContext(ctx, query, userID, roleDisplay).Scan(&hasRole)
	if err != nil {
		return false, fmt.Errorf("failed to check role: %w", err)
	}

	return hasRole, nil
}
