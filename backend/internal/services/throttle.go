package services

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/pinggolf/m3-planning-tools/internal/db"
	"golang.org/x/time/rate"
)

// RateLimiterService provides global API throttling across all workers
type RateLimiterService struct {
	mu           sync.RWMutex
	limiters     map[string]*rate.Limiter // key: environment
	settingsRepo *db.Queries
}

// NewRateLimiterService creates a new rate limiter service
func NewRateLimiterService(settingsRepo *db.Queries) *RateLimiterService {
	return &RateLimiterService{
		limiters:     make(map[string]*rate.Limiter),
		settingsRepo: settingsRepo,
	}
}

// GetLimiter returns or creates rate limiter for environment
func (s *RateLimiterService) GetLimiter(ctx context.Context, env string) (*rate.Limiter, error) {
	s.mu.RLock()
	limiter, exists := s.limiters[env]
	s.mu.RUnlock()

	if exists {
		return limiter, nil
	}

	// Load limiter with write lock
	return s.loadLimiter(ctx, env)
}

// loadLimiter loads rate limiter settings from database and creates a new limiter
func (s *RateLimiterService) loadLimiter(ctx context.Context, env string) (*rate.Limiter, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := s.limiters[env]; exists {
		return limiter, nil
	}

	// Fetch settings from database
	settings, err := s.settingsRepo.GetSystemSettings(ctx, env)
	if err != nil {
		return nil, fmt.Errorf("failed to get system settings for environment %s: %w", env, err)
	}

	// Parse throttle settings with defaults
	requestsPerSec := parseIntSetting(settings, "api_throttle_requests_per_second", 10)
	burstSize := parseIntSetting(settings, "api_throttle_burst_size", 5)

	// Create rate limiter with token bucket algorithm
	// rate.Limit is requests per second, burst is max tokens in bucket
	limiter := rate.NewLimiter(rate.Limit(requestsPerSec), burstSize)
	s.limiters[env] = limiter

	return limiter, nil
}

// Wait blocks until request is allowed under rate limit
// This is the main method workers should call before making M3 API requests
func (s *RateLimiterService) Wait(ctx context.Context, env string) error {
	limiter, err := s.GetLimiter(ctx, env)
	if err != nil {
		return err
	}
	return limiter.Wait(ctx)
}

// Allow checks if request is allowed without blocking
// Returns true if the request can proceed immediately
func (s *RateLimiterService) Allow(ctx context.Context, env string) (bool, error) {
	limiter, err := s.GetLimiter(ctx, env)
	if err != nil {
		return false, err
	}
	return limiter.Allow(), nil
}

// ReloadSettings refreshes rate limiters when settings change
// This can be called by API handlers when admin updates throttle settings
func (s *RateLimiterService) ReloadSettings(ctx context.Context, env string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove old limiter
	delete(s.limiters, env)

	// Load new limiter (will create with updated settings)
	_, err := s.loadLimiter(ctx, env)
	return err
}

// parseIntSetting extracts an integer setting value with a default fallback
func parseIntSetting(settings []db.SystemSetting, key string, defaultValue int) int {
	for _, setting := range settings {
		if setting.SettingKey == key {
			if val, err := strconv.Atoi(setting.SettingValue); err == nil {
				return val
			}
		}
	}
	return defaultValue
}
