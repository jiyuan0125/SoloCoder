package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"backoff/common"
)

type RetryClient struct {
	serverAddr string
	httpClient *http.Client
}

func NewRetryClient(serverAddr string) *RetryClient {
	return &RetryClient{
		serverAddr: serverAddr,
		httpClient: &http.Client{},
	}
}

func (c *RetryClient) Execute(req common.ExecuteRequest) (*common.ExecuteResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.serverAddr + "/execute"
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil {
			return nil, fmt.Errorf("server error [%s]: %s", errResp.Code, errResp.Message)
		}
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var execResp common.ExecuteResponse
	if err := json.Unmarshal(respBody, &execResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &execResp, nil
}

func (c *RetryClient) Health() (bool, error) {
	url := c.serverAddr + "/health"
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return false, fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
