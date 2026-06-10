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
	AppEnv        string
	AppPort       int
	FrontendURL   string
	RunMigrations bool

	// Database settings
	DatabaseURL                string
	DatabaseMaxConnections     int
	DatabaseMaxIdleConnections int
	DatabaseConnectionLifetime time.Duration

	// M3 environment (single environment per deployment — which M3 tenant
	// this instance talks to is a deploy-time decision, not a user choice)
	M3Env            string // label used in NATS subjects and DB rows: "TRN" or "PRD"
	M3TenantID       string
	M3InstanceID     string
	M3APIBaseURL     string
	M3CompassBaseURL string
	M3IONAPI         string // raw JSON content of the .ionapi service-account file

	// Entra ID user authentication (OIDC authorization-code flow)
	EntraTenantID     string
	EntraClientID     string
	EntraClientSecret string

	// OAuth/session settings
	OAuthRedirectURI string
	SessionSecret    string
	SessionDuration  time.Duration

	// CORS settings
	CORSAllowedOrigins   string
	CORSAllowCredentials bool

	// Logging
	LogLevel  string
	LogFormat string

	// NATS settings
	NATSURL string

	// Data refresh settings
	MaxQueryRecords      int
	QueryTimeout         int
	MaxConcurrentQueries int
}

// M3Environment represents TRN or PRD environment configuration
type M3Environment struct {
	TenantID       string
	InstanceID     string
	APIBaseURL     string
	CompassBaseURL string
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

		M3Env:            getEnv("M3_ENVIRONMENT", "PRD"),
		M3TenantID:       getEnv("M3_TENANT_ID", ""),
		M3InstanceID:     getEnv("M3_INSTANCE_ID", ""),
		M3APIBaseURL:     getEnv("M3_API_BASE_URL", ""),
		M3CompassBaseURL: getEnv("M3_COMPASS_BASE_URL", ""),
		M3IONAPI:         getEnv("M3_IONAPI", ""),

		EntraTenantID:     getEnv("ENTRA_TENANT_ID", ""),
		EntraClientID:     getEnv("ENTRA_CLIENT_ID", ""),
		EntraClientSecret: getEnv("ENTRA_CLIENT_SECRET", ""),

		OAuthRedirectURI: getEnv("OAUTH_REDIRECT_URI", "http://localhost:8080/api/auth/callback"),
		SessionSecret:    getEnv("SESSION_SECRET", ""),
		SessionDuration:  getEnvAsDuration("SESSION_DURATION", 24*time.Hour),

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
	// The label flows into NATS subjects and DB environment columns, which
	// are keyed on these two values.
	if c.M3Env != "TRN" && c.M3Env != "PRD" {
		return fmt.Errorf("M3_ENVIRONMENT must be TRN or PRD, got %q", c.M3Env)
	}
	// Entra ID and .ionapi credentials are validated lazily where used so
	// the app can boot in partially configured environments, but fail fast
	// in production where a misconfigured login flow is never acceptable.
	if c.AppEnv == "production" {
		if c.EntraTenantID == "" || c.EntraClientID == "" || c.EntraClientSecret == "" {
			return fmt.Errorf("ENTRA_TENANT_ID, ENTRA_CLIENT_ID and ENTRA_CLIENT_SECRET are required in production")
		}
	}
	return nil
}

// GetEnvironmentConfig returns the M3 connection configuration. The app
// supports a single environment per deployment; the env argument is kept
// for call-site compatibility and validated against the configured label.
func (c *Config) GetEnvironmentConfig(env string) (*M3Environment, error) {
	if env != "" && env != c.M3Env {
		return nil, fmt.Errorf("environment %q is not available; this deployment is configured for %s only", env, c.M3Env)
	}
	return &M3Environment{
		TenantID:       c.M3TenantID,
		InstanceID:     c.M3InstanceID,
		APIBaseURL:     c.M3APIBaseURL,
		CompassBaseURL: c.M3CompassBaseURL,
	}, nil
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
