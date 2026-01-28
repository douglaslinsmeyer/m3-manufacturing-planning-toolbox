package m3api

import (
	"bytes"
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
			Timeout: 60 * time.Second, // Increased from 30s to 60s for bulk operations
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

// BulkRequestItem represents a single transaction in a bulk request
type BulkRequestItem struct {
	Program     string            `json:"-"` // Not serialized - used for grouping only
	Transaction string            `json:"transaction"`
	Record      map[string]string `json:"record,omitempty"` // Changed from "parameters" to "record"
}

// BulkRequest represents a batch of M3 API transactions for a SINGLE program
// Note: All transactions in a bulk request must be for the same program
// Field order matters: "transactions" must be LAST for streaming
type BulkRequest struct {
	Program             string                   `json:"program"`
	ExcludeEmptyValues  bool                     `json:"excludeEmptyValues"`
	RightTrim           bool                     `json:"rightTrim"`
	MaxReturnedRecords  int                      `json:"maxReturnedRecords"`
	Transactions        []BulkRequestTransaction `json:"transactions"` // Must be last attribute for streaming
}

// BulkRequestTransaction represents a single transaction within a bulk request
type BulkRequestTransaction struct {
	Transaction string            `json:"transaction"`
	Record      map[string]string `json:"record,omitempty"`
}

// BulkResultItem represents the result of a single transaction (from swagger: BulkResponseResult)
type BulkResultItem struct {
	Transaction  string                   `json:"transaction"`
	Parameters   map[string]string        `json:"parameters,omitempty"`
	ErrorMessage string                   `json:"errorMessage,omitempty"`
	ErrorCode    string                   `json:"errorCode,omitempty"`
	ErrorCfg     string                   `json:"errorCfg,omitempty"`
	ErrorField   string                   `json:"errorField,omitempty"`
	ErrorType    string                   `json:"errorType,omitempty"`
	NotProcessed bool                     `json:"notProcessed,omitempty"`
	Records      []map[string]interface{} `json:"records,omitempty"`
	PositionKey  string                   `json:"positionKey,omitempty"`
}

// BulkResponse represents the response from a bulk operation
type BulkResponse struct {
	Results                      []BulkResultItem `json:"results"`
	WasTerminated                bool             `json:"wasTerminated"`
	TerminationReason            string           `json:"terminationReason,omitempty"`
	NrOfNotProcessedTransactions int              `json:"nrOfNotProcessedTransactions"`
	NrOfFailedTransactions       int              `json:"nrOfFailedTransactions"`
	TerminationErrorType         string           `json:"terminationErrorType,omitempty"`
	BulkJobID                    string           `json:"bulkJobId,omitempty"`
	NrOfSuccessfullTransactions  int              `json:"nrOfSuccessfullTransactions"` // Note: Infor misspells "Successful"
}

// BulkErrorDetail contains detailed error information for failed transactions (legacy, kept for compatibility)
type BulkErrorDetail struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
	Type    string `json:"type,omitempty"`
}

// IsSuccess returns true if this transaction result was successful
func (r *BulkResultItem) IsSuccess() bool {
	return !r.NotProcessed && r.ErrorMessage == ""
}

// BulkOperationError is a structured error type for bulk operations
type BulkOperationError struct {
	TotalRequests  int
	SuccessCount   int
	FailureCount   int
	FailedItems    []BulkResultItem
	HTTPStatusCode int   // HTTP-level error
	NetworkError   error // Network-level error
}

func (e *BulkOperationError) Error() string {
	if e.NetworkError != nil {
		return fmt.Sprintf("bulk operation network error: %v", e.NetworkError)
	}
	if e.HTTPStatusCode != 0 && e.HTTPStatusCode != http.StatusOK {
		return fmt.Sprintf("bulk operation HTTP error: status %d", e.HTTPStatusCode)
	}
	return fmt.Sprintf("bulk operation partial failure: %d/%d succeeded",
		e.SuccessCount, e.TotalRequests)
}

func (e *BulkOperationError) IsPartialSuccess() bool {
	return e.SuccessCount > 0 && e.FailureCount > 0
}

// ExecuteBulk executes multiple M3 API transactions in a single request
// Returns BulkResponse with all results and BulkOperationError if any failures occurred
func (c *Client) ExecuteBulk(ctx context.Context, requests []BulkRequestItem) (*BulkResponse, error) {
	if len(requests) == 0 {
		return &BulkResponse{Results: []BulkResultItem{}}, nil
	}

	// Group requests by program (bulk calls are program-specific per swagger)
	programGroups := make(map[string][]BulkRequestItem)
	for _, req := range requests {
		programGroups[req.Program] = append(programGroups[req.Program], req)
	}

	fmt.Printf("DEBUG ExecuteBulk: Grouped %d requests into %d programs\n", len(requests), len(programGroups))

	// Execute one bulk call per program and combine results
	combinedResults := []BulkResultItem{}
	totalSuccess := 0
	totalFailure := 0
	var allFailedItems []BulkResultItem

	for program, programRequests := range programGroups {
		fmt.Printf("DEBUG ExecuteBulk: Calling %s with %d transactions\n", program, len(programRequests))

		bulkResp, err := c.ExecuteProgramBulk(ctx, program, programRequests)
		if err != nil {
			if bulkErr, ok := err.(*BulkOperationError); ok {
				totalSuccess += bulkErr.SuccessCount
				totalFailure += bulkErr.FailureCount
				allFailedItems = append(allFailedItems, bulkErr.FailedItems...)
			} else {
				// Complete network failure for this program
				totalFailure += len(programRequests)
				return nil, &BulkOperationError{
					TotalRequests: len(requests),
					FailureCount:  totalFailure,
					NetworkError:  fmt.Errorf("program %s failed: %w", program, err),
				}
			}
		} else {
			totalSuccess += len(programRequests)
		}

		if bulkResp != nil {
			combinedResults = append(combinedResults, bulkResp.Results...)
		}
	}

	// Build combined response
	combinedResp := &BulkResponse{
		Results:                     combinedResults,
		NrOfSuccessfullTransactions: totalSuccess,
		NrOfFailedTransactions:      totalFailure,
	}

	if totalFailure > 0 {
		bulkErr := &BulkOperationError{
			TotalRequests: len(requests),
			SuccessCount:  totalSuccess,
			FailureCount:  totalFailure,
			FailedItems:   allFailedItems,
		}
		return combinedResp, bulkErr
	}

	return combinedResp, nil
}

// ExecuteProgramBulk executes multiple transactions for a SINGLE program in one bulk request
func (c *Client) ExecuteProgramBulk(ctx context.Context, program string, requests []BulkRequestItem) (*BulkResponse, error) {
	if len(requests) == 0 {
		return &BulkResponse{Results: []BulkResultItem{}}, nil
	}

	// Build URL: /M3/m3api-rest/v2/execute
	url := fmt.Sprintf("%sM3/m3api-rest/v2/execute", c.baseURL)

	// Build transactions array
	transactions := make([]BulkRequestTransaction, len(requests))
	for i, req := range requests {
		transactions[i] = BulkRequestTransaction{
			Transaction: req.Transaction,
			Record:      req.Record,
		}
	}

	// Build request body with "transactions" as LAST attribute (required for streaming)
	bulkReq := BulkRequest{
		Program:            program,
		ExcludeEmptyValues: false,
		RightTrim:          true,
		MaxReturnedRecords: 0,
		Transactions:       transactions,
	}

	bodyBytes, err := json.Marshal(bulkReq)
	if err != nil {
		return nil, &BulkOperationError{
			TotalRequests: len(requests),
			NetworkError:  fmt.Errorf("failed to marshal bulk request: %w", err),
		}
	}

	fmt.Printf("DEBUG Bulk request body for %s: %s\n", program, string(bodyBytes))

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, &BulkOperationError{
			TotalRequests: len(requests),
			NetworkError:  fmt.Errorf("failed to create bulk request: %w", err),
		}
	}

	// Add query parameters (same defaults as single operations)
	q := httpReq.URL.Query()
	q.Add("dateformat", "YMD8")
	q.Add("excludeempty", "false")
	q.Add("righttrim", "true")
	q.Add("extendedresult", "true") // Include transaction parameters in response
	httpReq.URL.RawQuery = q.Encode()

	// Add authentication
	token, err := c.getToken()
	if err != nil {
		return nil, &BulkOperationError{
			TotalRequests: len(requests),
			NetworkError:  fmt.Errorf("failed to get auth token: %w", err),
		}
	}
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	httpReq.Header.Set("Content-Type", "application/json; charset=UTF-8")
	httpReq.Header.Set("Accept", "application/json; charset=UTF-8")

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &BulkOperationError{
			TotalRequests: len(requests),
			NetworkError:  fmt.Errorf("failed to execute bulk request: %w", err),
		}
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &BulkOperationError{
			TotalRequests: len(requests),
			NetworkError:  fmt.Errorf("failed to read bulk response: %w", err),
		}
	}

	// Handle non-200 HTTP status (complete failure)
	if resp.StatusCode != http.StatusOK {
		bulkErr := &BulkOperationError{
			TotalRequests:  len(requests),
			HTTPStatusCode: resp.StatusCode,
		}
		return nil, fmt.Errorf("bulk operation failed with HTTP %d: %w - body: %s", resp.StatusCode, bulkErr, string(body))
	}

	// Debug: Log raw response for format validation during initial testing
	fmt.Printf("DEBUG Bulk API Response for %s:\n%s\n", program, string(body))

	// Parse response
	var bulkResp BulkResponse
	if err := json.Unmarshal(body, &bulkResp); err != nil {
		return nil, &BulkOperationError{
			TotalRequests: len(requests),
			NetworkError:  fmt.Errorf("failed to parse bulk response: %w", err),
		}
	}

	// Collect failed items for error reporting
	var failedItems []BulkResultItem
	for _, result := range bulkResp.Results {
		if !result.IsSuccess() {
			failedItems = append(failedItems, result)
		}
	}

	// Use counts from response (swagger: nrOfSuccessfullTransactions, nrOfFailedTransactions)
	successCount := bulkResp.NrOfSuccessfullTransactions
	failureCount := bulkResp.NrOfFailedTransactions

	fmt.Printf("DEBUG Bulk API Results for %s: %d succeeded, %d failed, terminated=%v\n",
		program, successCount, failureCount, bulkResp.WasTerminated)

	// If any failures or termination, return BulkOperationError WITH the BulkResponse
	if failureCount > 0 || bulkResp.WasTerminated {
		bulkErr := &BulkOperationError{
			TotalRequests: len(requests),
			SuccessCount:  successCount,
			FailureCount:  failureCount,
			FailedItems:   failedItems,
		}
		return &bulkResp, bulkErr
	}

	return &bulkResp, nil
}
