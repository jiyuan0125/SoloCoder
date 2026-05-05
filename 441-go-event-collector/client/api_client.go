package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"event-collector/common"
)

type APIClient struct {
	serverAddr string
	httpClient *http.Client
	maxRetries int
	retryDelay time.Duration
}

func NewAPIClient(serverAddr string) *APIClient {
	return &APIClient{
		serverAddr: serverAddr,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		maxRetries: common.MaxRetryAttempts,
		retryDelay: 1 * time.Second,
	}
}

func (c *APIClient) Track(event *common.Event) (*common.TrackResponse, error) {
	reqBody := common.TrackRequest{
		Event: *event,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/track", c.serverAddr)

	var resp common.TrackResponse
	err = c.doWithRetry(http.MethodPost, url, jsonBody, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) Query(req *common.QueryRequest) (*common.QueryResponse, error) {
	queryParams := url.Values{}
	if req.EventName != "" {
		queryParams.Set("event_name", req.EventName)
	}
	if req.UserID != "" {
		queryParams.Set("user_id", req.UserID)
	}
	if req.StartTime > 0 {
		queryParams.Set("start_time", strconv.FormatInt(req.StartTime, 10))
	}
	if req.EndTime > 0 {
		queryParams.Set("end_time", strconv.FormatInt(req.EndTime, 10))
	}
	if req.Page > 0 {
		queryParams.Set("page", strconv.Itoa(req.Page))
	}
	if req.PageSize > 0 {
		queryParams.Set("page_size", strconv.Itoa(req.PageSize))
	}

	url := fmt.Sprintf("%s/query?%s", c.serverAddr, queryParams.Encode())

	var resp common.QueryResponse
	err := c.doWithRetry(http.MethodGet, url, nil, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) Aggregate(req *common.AggregationRequest) (*common.AggregationResponse, error) {
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/aggregate", c.serverAddr)

	var resp common.AggregationResponse
	err = c.doWithRetry(http.MethodPost, url, jsonBody, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) GetReports() (interface{}, error) {
	url := fmt.Sprintf("%s/reports", c.serverAddr)

	var resp interface{}
	err := c.doWithRetry(http.MethodGet, url, nil, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) doWithRetry(method, url string, body []byte, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(c.retryDelay * time.Duration(attempt))
		}

		resp, err := c.do(method, url, body)
		if err != nil {
			lastErr = err
			continue
		}

		if result != nil {
			if err := json.Unmarshal(resp, result); err != nil {
				lastErr = fmt.Errorf("failed to unmarshal response: %w", err)
				continue
			}
		}

		return nil
	}

	return fmt.Errorf("failed after %d attempts: %w", c.maxRetries, lastErr)
}

func (c *APIClient) do(method, url string, body []byte) ([]byte, error) {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned error status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
