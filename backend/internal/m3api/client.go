package m3api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles M3 REST API calls (MI programs)
type Client struct {
	baseURL    string
	httpClient *http.Client
	getToken   func() (string, error)
}

// NewClient creates a new M3 API client
func NewClient(baseURL string, getToken func() (string, error)) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		getToken: getToken,
	}
}

// M3Response represents a generic M3 API response
type M3Response struct {
	Results []M3TransactionResult `json:"results"`
}

// M3TransactionResult represents a single transaction result
type M3TransactionResult struct {
	Transaction string                   `json:"transaction"`
	Records     []map[string]interface{} `json:"records"`
}

// Execute calls an M3 API transaction
func (c *Client) Execute(ctx context.Context, program, transaction string, params map[string]string) (*M3Response, error) {
	// Build URL: /M3/m3api-rest/v2/execute/{program}/{transaction}
	url := fmt.Sprintf("%sM3/m3api-rest/v2/execute/%s/%s", c.baseURL, program, transaction)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters with defaults from M3 Shop Floor app
	q := req.URL.Query()
	q.Add("dateformat", "YMD8")
	q.Add("excludeempty", "false")
	q.Add("righttrim", "true")
	q.Add("metadata", "false")
	q.Add("returnSystemFields", "false")

	// Add custom parameters (will override defaults if same key)
	if len(params) > 0 {
		for key, value := range params {
			q.Set(key, value)
		}
	}
	req.URL.RawQuery = q.Encode()

	// Add authentication
	token, err := c.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("M3 API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Debug: Log raw response
	fmt.Printf("DEBUG M3 API Response for %s/%s:\n%s\n", program, transaction, string(body))

	// Parse response
	var m3Resp M3Response
	if err := json.Unmarshal(body, &m3Resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	recordCount := 0
	if len(m3Resp.Results) > 0 {
		recordCount = len(m3Resp.Results[0].Records)
	}
	fmt.Printf("DEBUG M3 API Parsed: Found %d records\n", recordCount)

	return &m3Resp, nil
}

// GetSingleRecord executes a transaction expecting a single record
func (c *Client) GetSingleRecord(ctx context.Context, program, transaction string, params map[string]string) (map[string]interface{}, error) {
	resp, err := c.Execute(ctx, program, transaction, params)
	if err != nil {
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no results returned")
	}

	records := resp.Results[0].Records
	if len(records) == 0 {
		return nil, fmt.Errorf("no records returned")
	}

	return records[0], nil
}

// GetMultipleRecords executes a transaction expecting multiple records
func (c *Client) GetMultipleRecords(ctx context.Context, program, transaction string, params map[string]string) ([]map[string]interface{}, error) {
	resp, err := c.Execute(ctx, program, transaction, params)
	if err != nil {
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no results returned")
	}

	return resp.Results[0].Records, nil
}
