package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Application settings
	AppEnv      string
	AppPort     int
	FrontendURL string
	RunMigrations bool

	// Database settings
	DatabaseURL                  string
	DatabaseMaxConnections       int
	DatabaseMaxIdleConnections   int
	DatabaseConnectionLifetime   time.Duration

	// M3 TRN Environment
	TRNTenantID         string
	TRNInstanceID       string
	TRNClientID         string
	TRNClientSecret     string
	TRNAuthEndpoint     string
	TRNTokenEndpoint    string
	TRNAPIBaseURL       string
	TRNCompassBaseURL   string

	// M3 PRD Environment
	PRDTenantID         string
	PRDInstanceID       string
	PRDClientID         string
	PRDClientSecret     string
	PRDAuthEndpoint     string
	PRDTokenEndpoint    string
	PRDAPIBaseURL       string
	PRDCompassBaseURL   string

	// OAuth settings
	OAuthRedirectURI    string
	OAuthScopes         string
	SessionSecret       string
	SessionDuration     time.Duration
	TokenRefreshBuffer  time.Duration

	// CORS settings
	CORSAllowedOrigins  string
	CORSAllowCredentials bool

	// Logging
	LogLevel  string
	LogFormat string

	// NATS settings
	NATSURL string

	// Data refresh settings
	MaxQueryRecords       int
	QueryTimeout          int
	MaxConcurrentQueries  int
}

// M3Environment represents TRN or PRD environment configuration
type M3Environment struct {
	TenantID        string
	InstanceID      string
	ClientID        string
	ClientSecret    string
	AuthEndpoint    string
	TokenEndpoint   string
	APIBaseURL      string
	CompassBaseURL  string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		AppPort:     getEnvAsInt("APP_PORT", 8080),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),

		DatabaseURL:                getEnv("DATABASE_URL", ""),
		DatabaseMaxConnections:     getEnvAsInt("DATABASE_MAX_CONNECTIONS", 25),
		DatabaseMaxIdleConnections: getEnvAsInt("DATABASE_MAX_IDLE_CONNECTIONS", 5),
		DatabaseConnectionLifetime: getEnvAsDuration("DATABASE_CONNECTION_LIFETIME", 5*time.Minute),

		TRNTenantID:       getEnv("TRN_TENANT_ID", ""),
		TRNInstanceID:     getEnv("TRN_INSTANCE_ID", ""),
		TRNClientID:       getEnv("TRN_CLIENT_ID", ""),
		TRNClientSecret:   getEnv("TRN_CLIENT_SECRET", ""),
		TRNAuthEndpoint:   getEnv("TRN_AUTH_ENDPOINT", ""),
		TRNTokenEndpoint:  getEnv("TRN_TOKEN_ENDPOINT", ""),
		TRNAPIBaseURL:     getEnv("TRN_API_BASE_URL", ""),
		TRNCompassBaseURL: getEnv("TRN_COMPASS_BASE_URL", ""),

		PRDTenantID:       getEnv("PRD_TENANT_ID", ""),
		PRDInstanceID:     getEnv("PRD_INSTANCE_ID", ""),
		PRDClientID:       getEnv("PRD_CLIENT_ID", ""),
		PRDClientSecret:   getEnv("PRD_CLIENT_SECRET", ""),
		PRDAuthEndpoint:   getEnv("PRD_AUTH_ENDPOINT", ""),
		PRDTokenEndpoint:  getEnv("PRD_TOKEN_ENDPOINT", ""),
		PRDAPIBaseURL:     getEnv("PRD_API_BASE_URL", ""),
		PRDCompassBaseURL: getEnv("PRD_COMPASS_BASE_URL", ""),

		OAuthRedirectURI:   getEnv("OAUTH_REDIRECT_URI", "http://localhost:8080/api/auth/callback"),
		OAuthScopes:        getEnv("OAUTH_SCOPES", "openid profile"),
		SessionSecret:      getEnv("SESSION_SECRET", ""),
		SessionDuration:    getEnvAsDuration("SESSION_DURATION", 24*time.Hour),
		TokenRefreshBuffer: getEnvAsDuration("TOKEN_REFRESH_BUFFER", 5*time.Minute),

		CORSAllowedOrigins:   getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		CORSAllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),

		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),

		NATSURL: getEnv("NATS_URL", "nats://localhost:4222"),

		MaxQueryRecords:      getEnvAsInt("MAX_QUERY_RECORDS", 100000),
		QueryTimeout:         getEnvAsInt("QUERY_TIMEOUT", 300),
		MaxConcurrentQueries: getEnvAsInt("MAX_CONCURRENT_QUERIES", 5),

		RunMigrations: getEnvAsBool("RUN_MIGRATIONS", false),
	}

	// Validate required configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.SessionSecret == "" {
		return fmt.Errorf("SESSION_SECRET is required")
	}
	if c.TRNClientID == "" || c.TRNClientSecret == "" {
		return fmt.Errorf("TRN OAuth credentials are required")
	}
	if c.PRDClientID == "" || c.PRDClientSecret == "" {
		return fmt.Errorf("PRD OAuth credentials are required")
	}
	return nil
}

// GetEnvironmentConfig returns configuration for the specified environment
func (c *Config) GetEnvironmentConfig(env string) (*M3Environment, error) {
	switch env {
	case "TRN":
		return &M3Environment{
			TenantID:       c.TRNTenantID,
			InstanceID:     c.TRNInstanceID,
			ClientID:       c.TRNClientID,
			ClientSecret:   c.TRNClientSecret,
			AuthEndpoint:   c.TRNAuthEndpoint,
			TokenEndpoint:  c.TRNTokenEndpoint,
			APIBaseURL:     c.TRNAPIBaseURL,
			CompassBaseURL: c.TRNCompassBaseURL,
		}, nil
	case "PRD":
		return &M3Environment{
			TenantID:       c.PRDTenantID,
			InstanceID:     c.PRDInstanceID,
			ClientID:       c.PRDClientID,
			ClientSecret:   c.PRDClientSecret,
			AuthEndpoint:   c.PRDAuthEndpoint,
			TokenEndpoint:  c.PRDTokenEndpoint,
			APIBaseURL:     c.PRDAPIBaseURL,
			CompassBaseURL: c.PRDCompassBaseURL,
		}, nil
	default:
		return nil, fmt.Errorf("invalid environment: %s", env)
	}
}

// Helper functions for reading environment variables

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
