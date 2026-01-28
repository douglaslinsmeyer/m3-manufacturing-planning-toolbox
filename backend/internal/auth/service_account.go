package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pinggolf/m3-planning-tools/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// ServiceAccountTokenManager manages OAuth tokens for service accounts (background workers)
// Uses client credentials flow instead of authorization code flow
type ServiceAccountTokenManager struct {
	trnConfig *clientcredentials.Config
	prdConfig *clientcredentials.Config
	trnToken  *oauth2.Token
	prdToken  *oauth2.Token
	trnMutex  sync.RWMutex
	prdMutex  sync.RWMutex
	config    *config.Config
}

// NewServiceAccountTokenManager creates a new service account token manager
func NewServiceAccountTokenManager(cfg *config.Config) *ServiceAccountTokenManager {
	// Configure client credentials for TRN environment
	trnConfig := &clientcredentials.Config{
		ClientID:     cfg.TRNClientID,
		ClientSecret: cfg.TRNClientSecret,
		TokenURL:     cfg.TRNTokenEndpoint,
		Scopes:       []string{}, // Client credentials typically don't need scopes
	}

	// Configure client credentials for PRD environment
	prdConfig := &clientcredentials.Config{
		ClientID:     cfg.PRDClientID,
		ClientSecret: cfg.PRDClientSecret,
		TokenURL:     cfg.PRDTokenEndpoint,
		Scopes:       []string{}, // Client credentials typically don't need scopes
	}

	return &ServiceAccountTokenManager{
		trnConfig: trnConfig,
		prdConfig: prdConfig,
		config:    cfg,
	}
}

// GetToken returns a valid access token for the specified environment
// Refreshes the token automatically if expired
func (m *ServiceAccountTokenManager) GetToken(environment string) (string, error) {
	switch environment {
	case "TRN":
		return m.getTRNToken()
	case "PRD":
		return m.getPRDToken()
	default:
		return "", fmt.Errorf("invalid environment: %s", environment)
	}
}

// getTRNToken gets or refreshes the TRN environment token
func (m *ServiceAccountTokenManager) getTRNToken() (string, error) {
	m.trnMutex.RLock()
	token := m.trnToken
	m.trnMutex.RUnlock()

	// Check if token is valid
	if token != nil && token.Valid() {
		return token.AccessToken, nil
	}

	// Token is expired or doesn't exist, acquire lock to refresh
	m.trnMutex.Lock()
	defer m.trnMutex.Unlock()

	// Double-check after acquiring write lock (another goroutine may have refreshed)
	if m.trnToken != nil && m.trnToken.Valid() {
		return m.trnToken.AccessToken, nil
	}

	// Fetch new token
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	newToken, err := m.trnConfig.Token(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get TRN token: %w", err)
	}

	m.trnToken = newToken
	fmt.Printf("Service account token obtained for TRN (expires: %v)\n", newToken.Expiry)

	return newToken.AccessToken, nil
}

// getPRDToken gets or refreshes the PRD environment token
func (m *ServiceAccountTokenManager) getPRDToken() (string, error) {
	m.prdMutex.RLock()
	token := m.prdToken
	m.prdMutex.RUnlock()

	// Check if token is valid
	if token != nil && token.Valid() {
		return token.AccessToken, nil
	}

	// Token is expired or doesn't exist, acquire lock to refresh
	m.prdMutex.Lock()
	defer m.prdMutex.Unlock()

	// Double-check after acquiring write lock (another goroutine may have refreshed)
	if m.prdToken != nil && m.prdToken.Valid() {
		return m.prdToken.AccessToken, nil
	}

	// Fetch new token
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	newToken, err := m.prdConfig.Token(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get PRD token: %w", err)
	}

	m.prdToken = newToken
	fmt.Printf("Service account token obtained for PRD (expires: %v)\n", newToken.Expiry)

	return newToken.AccessToken, nil
}
