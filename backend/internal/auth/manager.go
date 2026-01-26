package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/sessions"
	"github.com/pinggolf/m3-planning-tools/internal/config"
	"golang.org/x/oauth2"
)

// Manager handles authentication and token management
type Manager struct {
	config       *config.Config
	sessionStore sessions.Store
	trnOAuth     *oauth2.Config
	prdOAuth     *oauth2.Config
}

// NewManager creates a new auth manager
func NewManager(cfg *config.Config, store sessions.Store) *Manager {
	// Configure OAuth for TRN environment
	trnOAuth := &oauth2.Config{
		ClientID:     cfg.TRNClientID,
		ClientSecret: cfg.TRNClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.TRNAuthEndpoint,
			TokenURL: cfg.TRNTokenEndpoint,
		},
		RedirectURL: cfg.OAuthRedirectURI,
		Scopes:      []string{"openid", "profile"},
	}

	// Configure OAuth for PRD environment
	prdOAuth := &oauth2.Config{
		ClientID:     cfg.PRDClientID,
		ClientSecret: cfg.PRDClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.PRDAuthEndpoint,
			TokenURL: cfg.PRDTokenEndpoint,
		},
		RedirectURL: cfg.OAuthRedirectURI,
		Scopes:      []string{"openid", "profile"},
	}

	return &Manager{
		config:       cfg,
		sessionStore: store,
		trnOAuth:     trnOAuth,
		prdOAuth:     prdOAuth,
	}
}

// GetAuthorizationURL generates the OAuth authorization URL for the specified environment
func (m *Manager) GetAuthorizationURL(environment string) (string, error) {
	oauthConfig, err := m.getOAuthConfig(environment)
	if err != nil {
		return "", err
	}

	// Generate a random state for CSRF protection
	state := generateRandomState()

	// Generate authorization URL
	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	return authURL, nil
}

// ExchangeCodeForTokens exchanges an authorization code for access and refresh tokens
func (m *Manager) ExchangeCodeForTokens(ctx context.Context, environment, code string) (*oauth2.Token, error) {
	oauthConfig, err := m.getOAuthConfig(environment)
	if err != nil {
		return nil, err
	}

	// Exchange authorization code for token
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}

// RefreshTokenIfNeeded checks if the token needs refreshing and refreshes it if necessary
// Returns (true, nil) if token was refreshed, (false, nil) if still valid, (false, error) on failure
func (m *Manager) RefreshTokenIfNeeded(session *sessions.Session) (bool, error) {
	// Get token expiry from session
	expiryUnix, ok := session.Values["token_expiry"].(int64)
	if !ok {
		return false, fmt.Errorf("invalid token expiry in session")
	}

	expiry := time.Unix(expiryUnix, 0)
	timeUntilExpiry := time.Until(expiry)

	// Check if token is expiring within the refresh buffer (default 5 minutes)
	if timeUntilExpiry > m.config.TokenRefreshBuffer {
		return false, nil // Token is still valid, no refresh needed
	}

	// Get refresh token and environment
	refreshToken, ok := session.Values["refresh_token"].(string)
	if !ok || refreshToken == "" {
		return false, fmt.Errorf("no refresh token available")
	}

	environment, ok := session.Values["environment"].(string)
	if !ok {
		return false, fmt.Errorf("no environment in session")
	}

	fmt.Printf("Token refresh triggered - expires in %v (env: %s)\n", timeUntilExpiry, environment)

	// Get OAuth config
	oauthConfig, err := m.getOAuthConfig(environment)
	if err != nil {
		return false, err
	}

	// Create token source for refresh
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := oauthConfig.TokenSource(context.Background(), token)

	// Get fresh token
	newToken, err := tokenSource.Token()
	if err != nil {
		return false, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update session with new token
	session.Values["access_token"] = newToken.AccessToken
	if newToken.RefreshToken != "" {
		session.Values["refresh_token"] = newToken.RefreshToken
	}
	session.Values["token_expiry"] = newToken.Expiry.Unix()

	fmt.Printf("Token refreshed successfully - new expiry: %v\n", newToken.Expiry)

	return true, nil // Token was successfully refreshed
}

// GetAccessToken retrieves the access token from the session
func (m *Manager) GetAccessToken(session *sessions.Session) (string, error) {
	token, ok := session.Values["access_token"].(string)
	if !ok || token == "" {
		return "", fmt.Errorf("no access token in session")
	}
	return token, nil
}

// getOAuthConfig returns the OAuth config for the specified environment
func (m *Manager) getOAuthConfig(environment string) (*oauth2.Config, error) {
	switch environment {
	case "TRN":
		return m.trnOAuth, nil
	case "PRD":
		return m.prdOAuth, nil
	default:
		return nil, fmt.Errorf("invalid environment: %s", environment)
	}
}

// generateRandomState generates a random state string for CSRF protection
func generateRandomState() string {
	// For production, use crypto/rand to generate a secure random string
	// For now, using a simple timestamp-based state
	return fmt.Sprintf("state-%d", time.Now().UnixNano())
}

// Note: User profiles are cached in Postgres (user_profiles table) with 15-min TTL
// Session only stores user_profile_id for quick lookups
// This avoids cookie size limits and enables horizontal scaling
