package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// IONAPICredentials mirrors the .ionapi credentials file downloaded from
// Infor ION API > Authorized Apps for a "Backend Service" application.
// The service account (saak/sask) is exchanged for tokens via the OAuth2
// resource-owner password grant against the Infor STS.
type IONAPICredentials struct {
	TenantID     string `json:"ti"`   // e.g. XK3JRT8CJCAF9GWY_PRD
	AppName      string `json:"cn"`   // authorized app name
	ClientID     string `json:"ci"`   // OAuth client ID
	ClientSecret string `json:"cs"`   // OAuth client secret
	IONBaseURL   string `json:"iu"`   // https://mingle-ionapi.inforcloudsuite.com
	SSOBaseURL   string `json:"pu"`   // https://mingle-sso.inforcloudsuite.com:443/<tenant>/as/
	AuthPath     string `json:"oa"`   // authorization.oauth2
	TokenPath    string `json:"ot"`   // token.oauth2
	RevokePath   string `json:"or"`   // revoke_token.oauth2
	SAAccessKey  string `json:"saak"` // service account access key
	SASecretKey  string `json:"sask"` // service account secret key
}

// TokenURL joins the SSO base URL and token path from the .ionapi file.
func (c *IONAPICredentials) TokenURL() string {
	return strings.TrimRight(c.SSOBaseURL, "/") + "/" + strings.TrimLeft(c.TokenPath, "/")
}

// ParseIONAPICredentials parses the raw JSON content of a .ionapi file.
func ParseIONAPICredentials(raw string) (*IONAPICredentials, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("empty .ionapi credentials")
	}
	var creds IONAPICredentials
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return nil, fmt.Errorf("invalid .ionapi JSON: %w", err)
	}
	if creds.ClientID == "" || creds.ClientSecret == "" || creds.SAAccessKey == "" || creds.SASecretKey == "" {
		return nil, fmt.Errorf(".ionapi credentials missing ci/cs/saak/sask")
	}
	if creds.SSOBaseURL == "" || creds.TokenPath == "" {
		return nil, fmt.Errorf(".ionapi credentials missing pu/ot token endpoint parts")
	}
	return &creds, nil
}

// refreshSkew renews tokens this long before their actual expiry so a
// token never goes stale mid-request.
const refreshSkew = 60 * time.Second

// IONTokenManager provides M3/Compass access tokens for the application's
// single configured environment using the ION API service account. It
// replaces the previous per-user Infor OAuth tokens (and the
// client-credentials ServiceAccountTokenManager).
type IONTokenManager struct {
	creds *IONAPICredentials
	token *oauth2.Token
	mu    sync.Mutex
}

// NewIONTokenManager builds a manager from the raw JSON content of a
// .ionapi file (typically supplied via the M3_IONAPI env var). An empty
// raw string yields a manager whose GetToken returns a descriptive error,
// so the app can still boot in partially configured dev environments.
func NewIONTokenManager(raw string) (*IONTokenManager, error) {
	if strings.TrimSpace(raw) == "" {
		return &IONTokenManager{}, nil
	}
	creds, err := ParseIONAPICredentials(raw)
	if err != nil {
		return nil, err
	}
	return &IONTokenManager{creds: creds}, nil
}

// Configured reports whether service-account credentials are present.
func (m *IONTokenManager) Configured() bool {
	return m.creds != nil
}

// GetToken returns a valid service-account access token, fetching or
// renewing it as needed. Safe for concurrent use.
func (m *IONTokenManager) GetToken() (string, error) {
	if m.creds == nil {
		return "", fmt.Errorf("no ION API service-account credentials configured (set M3_IONAPI to the .ionapi file content)")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.token != nil && m.token.Expiry.After(time.Now().Add(refreshSkew)) {
		return m.token.AccessToken, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conf := &oauth2.Config{
		ClientID:     m.creds.ClientID,
		ClientSecret: m.creds.ClientSecret,
		Endpoint:     oauth2.Endpoint{TokenURL: m.creds.TokenURL()},
	}

	token, err := conf.PasswordCredentialsToken(ctx, m.creds.SAAccessKey, m.creds.SASecretKey)
	if err != nil {
		return "", fmt.Errorf("ION API service-account token request failed: %w", err)
	}

	m.token = token
	return token.AccessToken, nil
}
