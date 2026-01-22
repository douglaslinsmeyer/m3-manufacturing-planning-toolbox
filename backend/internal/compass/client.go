package compass

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles interactions with Compass Data Fabric API
type Client struct {
	baseURL    string
	httpClient *http.Client
	getToken   func() (string, error) // Function to get current access token
}

// NewClient creates a new Compass client
func NewClient(baseURL string, getToken func() (string, error)) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		getToken: getToken,
	}
}

// SubmitQueryRequest represents a query submission request
type SubmitQueryRequest struct {
	Query   string `json:"query"`
	Records string `json:"records,omitempty"` // "0" for all records, or specific count
}

// SubmitQueryResponse represents the response from query submission
type SubmitQueryResponse struct {
	JobID  string `json:"jobId"`
	Status string `json:"status"`
}

// QueryStatusResponse represents the status of a query
type QueryStatusResponse struct {
	JobID        string `json:"jobId"`
	Status       string `json:"status"`
	RecordCount  int    `json:"recordCount,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// QueryResult represents the raw result from Compass
type QueryResult struct {
	Records []map[string]interface{} `json:"records"`
	Columns []string                 `json:"columns"`
}

// SubmitQuery submits a SQL query to Compass Data Fabric
func (c *Client) SubmitQuery(ctx context.Context, query string, maxRecords int) (*SubmitQueryResponse, error) {
	// Build URL with query parameters (baseURL already ends with /)
	url := fmt.Sprintf("%sjobs/", c.baseURL)
	if maxRecords > 0 {
		url = fmt.Sprintf("%s?records=%d", url, maxRecords)
	} else {
		url = fmt.Sprintf("%s?records=0", url) // 0 means all records
	}

	// Log the query being submitted
	fmt.Printf("=== COMPASS QUERY SUBMIT ===\n")
	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Query:\n%s\n", query)
	fmt.Printf("===========================\n")

	// Create HTTP request with SQL query as plain text body
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(query))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication and headers
	token, err := c.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit query: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log full response
	fmt.Printf("=== COMPASS SUBMIT RESPONSE ===\n")
	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Response Body:\n%s\n", string(respBody))
	fmt.Printf("===============================\n")

	// Check status code - Compass returns 202 Accepted for async queries
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("query submission failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response - Compass returns queryId in different formats
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract queryId (could be "queryId" or "jobId")
	queryID := ""
	if id, ok := result["queryId"].(string); ok {
		queryID = id
	} else if id, ok := result["jobId"].(string); ok {
		queryID = id
	}

	status := ""
	if s, ok := result["status"].(string); ok {
		status = s
	}

	if queryID == "" {
		return nil, fmt.Errorf("no queryId in response: %s", string(respBody))
	}

	submitResp := &SubmitQueryResponse{
		JobID:  queryID,
		Status: status,
	}

	return submitResp, nil
}

// GetQueryStatus checks the status of a submitted query
func (c *Client) GetQueryStatus(ctx context.Context, jobID string, timeout int) (*QueryStatusResponse, error) {
	// IMPORTANT: timeout parameter is ALWAYS required, even when 0
	url := fmt.Sprintf("%sjobs/%s/status/?timeout=%d", c.baseURL, jobID, timeout)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication and headers
	token, err := c.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get query status: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log full response
	fmt.Printf("=== COMPASS STATUS RESPONSE ===\n")
	fmt.Printf("JobID: %s\n", jobID)
	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Response Body:\n%s\n", string(respBody))
	fmt.Printf("===============================\n")

	// Check status code - Compass returns 200, 201, or 202 for status checks
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status check failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response - handle both formats
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	status := ""
	if s, ok := result["status"].(string); ok {
		status = s
	}

	queryID := ""
	if id, ok := result["queryId"].(string); ok {
		queryID = id
	}

	recordCount := 0
	if rc, ok := result["recordCount"].(float64); ok {
		recordCount = int(rc)
	}

	errorMsg := ""
	if err, ok := result["errorMessage"].(string); ok {
		errorMsg = err
	}

	statusResp := &QueryStatusResponse{
		JobID:        queryID,
		Status:       status,
		RecordCount:  recordCount,
		ErrorMessage: errorMsg,
	}

	return statusResp, nil
}

// GetQueryResult fetches the results of a completed query
func (c *Client) GetQueryResult(ctx context.Context, jobID string, offset, limit int) ([]byte, error) {
	url := fmt.Sprintf("%sjobs/%s/result/?offset=%d&limit=%d", c.baseURL, jobID, offset, limit)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication and headers
	token, err := c.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get query result: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("result fetch failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// WaitForQueryCompletion polls the query status until it completes or fails
func (c *Client) WaitForQueryCompletion(ctx context.Context, jobID string, pollInterval time.Duration) (*QueryStatusResponse, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			status, err := c.GetQueryStatus(ctx, jobID, 0)
			if err != nil {
				return nil, err
			}

			switch status.Status {
			case "completed", "COMPLETED", "finished", "FINISHED":
				return status, nil
			case "failed", "FAILED", "error", "ERROR":
				// Try to fetch error details from result
				errorDetails, _ := c.GetQueryResult(ctx, jobID, 0, 10)
				if errorDetails != nil {
					return nil, fmt.Errorf("query failed - details: %s", string(errorDetails))
				}
				return nil, fmt.Errorf("query failed: %s", status.ErrorMessage)
			case "running", "RUNNING", "pending", "PENDING":
				// Continue polling
				continue
			default:
				return nil, fmt.Errorf("unknown query status: %s", status.Status)
			}
		}
	}
}

// ExecuteQuery is a convenience method that submits a query, waits for completion, and returns results
func (c *Client) ExecuteQuery(ctx context.Context, query string, maxRecords int) ([]byte, error) {
	// Submit query
	submitResp, err := c.SubmitQuery(ctx, query, maxRecords)
	if err != nil {
		return nil, fmt.Errorf("failed to submit query: %w", err)
	}

	// Wait for completion
	statusResp, err := c.WaitForQueryCompletion(ctx, submitResp.JobID, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	// Fetch results (single page for now, can be enhanced for pagination)
	results, err := c.GetQueryResult(ctx, submitResp.JobID, 0, maxRecords)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch results: %w", err)
	}

	// Log completion
	fmt.Printf("Query completed successfully. JobID: %s, Records: %d\n", submitResp.JobID, statusResp.RecordCount)

	return results, nil
}

// CancelQuery cancels a running query
func (c *Client) CancelQuery(ctx context.Context, jobID string) error {
	url := fmt.Sprintf("%sjobs/%s/", c.baseURL, jobID)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication
	token, err := c.getToken()
	if err != nil {
		return fmt.Errorf("failed to get auth token: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to cancel query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cancel failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
