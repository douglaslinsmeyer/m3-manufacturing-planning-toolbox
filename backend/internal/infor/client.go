package infor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pinggolf/m3-planning-tools/internal/m3api"
)

// Client is an HTTP client for Infor User Management API
type Client struct {
	baseURL    string
	httpClient *http.Client
	getToken   func() (string, error)
}

// NewClient creates a new Infor API client
func NewClient(baseURL string, getToken func() (string, error)) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		getToken:   getToken,
	}
}

// GetUserProfile fetches the current user's profile from Infor User Management API
func (c *Client) GetUserProfile(ctx context.Context) (*UserProfile, error) {
	// Get fresh token
	token, err := c.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// Build request
	url := c.baseURL + "ifsservice/usermgt/v2/users/me"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user profile request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var response UserProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Response.UserList) == 0 {
		return nil, fmt.Errorf("no user found in response")
	}

	return &response.Response.UserList[0], nil
}

// GetM3UserInfo fetches M3-specific user defaults from CRS650MI/GetUserInfo
func GetM3UserInfo(ctx context.Context, m3Client *m3api.Client) (*M3UserInfo, error) {
	// Call CRS650MI/GetUserInfo (no parameters needed)
	result, err := m3Client.Execute(ctx, "CRS650MI", "GetUserInfo", map[string]string{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch M3 user info: %w", err)
	}

	// Parse response
	if len(result.Results) == 0 {
		return nil, fmt.Errorf("no results in M3 user info response")
	}

	if len(result.Results[0].Records) == 0 {
		return nil, fmt.Errorf("no records in M3 user info response")
	}

	record := result.Results[0].Records[0]

	// Extract fields
	m3Info := &M3UserInfo{
		UserID:           getString(record, "ZZUSID"),
		FullName:         getString(record, "USFN"),
		DefaultCompany:   getString(record, "ZDCONO"),
		DefaultDivision:  getString(record, "ZDDIVI"),
		DefaultFacility:  getString(record, "ZDFACI"),
		DefaultWarehouse: getString(record, "ZZWHLO"),
		LanguageCode:     getString(record, "ZDLANC"),
		DateFormat:       getString(record, "ZDDTFM"),
		DateSeparator:    getString(record, "DSEP"),
		TimeSeparator:    getString(record, "TSEP"),
		TimeZone:         getString(record, "TIZO"),
	}

	return m3Info, nil
}

// Helper to safely extract string from map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
