package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/pinggolf/m3-planning-tools/internal/config"
	"golang.org/x/oauth2"
)

// EntraUser is the identity extracted from a verified Entra ID token.
type EntraUser struct {
	ObjectID string   // Entra object id (oid claim) — stable user identifier
	Email    string   // preferred_username / email claim
	Name     string   // display name
	Roles    []string // app roles assigned to the user (roles claim)
}

// HasRole reports whether the user carries the given app role value.
func (u *EntraUser) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// EntraAuthenticator implements the Entra ID OIDC authorization-code flow
// for user sign-in. M3 data access is NOT tied to the user token — it goes
// through the ION API service account (see IONTokenManager).
type EntraAuthenticator struct {
	cfg *config.Config

	initOnce sync.Once
	initErr  error
	oauth    *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

// NewEntraAuthenticator creates the authenticator. OIDC discovery against
// login.microsoftonline.com is deferred to first use so the process can boot
// (and worker-only pods can run) without immediate network access.
func NewEntraAuthenticator(cfg *config.Config) *EntraAuthenticator {
	return &EntraAuthenticator{cfg: cfg}
}

func (e *EntraAuthenticator) init() error {
	e.initOnce.Do(func() {
		if e.cfg.EntraTenantID == "" || e.cfg.EntraClientID == "" || e.cfg.EntraClientSecret == "" {
			e.initErr = fmt.Errorf("Entra ID is not configured (ENTRA_TENANT_ID / ENTRA_CLIENT_ID / ENTRA_CLIENT_SECRET)")
			return
		}

		issuer := fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", e.cfg.EntraTenantID)
		provider, err := oidc.NewProvider(context.Background(), issuer)
		if err != nil {
			e.initErr = fmt.Errorf("entra OIDC discovery failed: %w", err)
			// Allow a retry on the next call instead of caching a transient
			// network failure forever.
			e.initOnce = sync.Once{}
			return
		}

		e.oauth = &oauth2.Config{
			ClientID:     e.cfg.EntraClientID,
			ClientSecret: e.cfg.EntraClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  e.cfg.OAuthRedirectURI,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "offline_access"},
		}
		e.verifier = provider.Verifier(&oidc.Config{ClientID: e.cfg.EntraClientID})
	})
	return e.initErr
}

// AuthorizationURL returns the Entra consent/login URL for the given
// CSRF state value.
func (e *EntraAuthenticator) AuthorizationURL(state string) (string, error) {
	if err := e.init(); err != nil {
		return "", err
	}
	return e.oauth.AuthCodeURL(state), nil
}

// ExchangeCode redeems the authorization code, verifies the ID token
// signature and audience against the tenant JWKS, and extracts the user.
func (e *EntraAuthenticator) ExchangeCode(ctx context.Context, code string) (*EntraUser, error) {
	if err := e.init(); err != nil {
		return nil, err
	}

	token, err := e.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, fmt.Errorf("token response missing id_token")
	}

	idToken, err := e.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("ID token verification failed: %w", err)
	}

	var claims struct {
		OID               string   `json:"oid"`
		PreferredUsername string   `json:"preferred_username"`
		Email             string   `json:"email"`
		Name              string   `json:"name"`
		Roles             []string `json:"roles"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse ID token claims: %w", err)
	}

	email := claims.Email
	if email == "" {
		email = claims.PreferredUsername
	}

	return &EntraUser{
		ObjectID: claims.OID,
		Email:    email,
		Name:     claims.Name,
		Roles:    claims.Roles,
	}, nil
}

// GenerateState returns a cryptographically random CSRF state value.
func GenerateState() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
