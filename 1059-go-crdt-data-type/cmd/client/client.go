package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/example/crdt/pkg/api"
)

type Client struct {
	BaseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) CreateInstance(crdtType api.CRDTType) (*api.CreateInstanceResponse, error) {
	req := api.CreateInstanceRequest{Type: crdtType}
	var resp api.CreateInstanceResponse
	if err := c.doRequest(http.MethodPost, "/instances", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetInstance(id string) (*api.GetInstanceResponse, error) {
	var resp api.GetInstanceResponse
	if err := c.doRequest(http.MethodGet, "/instances/"+id, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ListInstances() ([]api.GetInstanceResponse, error) {
	var resp []api.GetInstanceResponse
	if err := c.doRequest(http.MethodGet, "/instances", nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) ExecuteOperation(id string, req api.OperationRequest) (*api.OperationResponse, error) {
	var resp api.OperationResponse
	if err := c.doRequest(http.MethodPost, "/instances/"+id+"/operations", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) MergeInstances(targetID, sourceID string) (*api.MergeResponse, error) {
	req := api.MergeRequest{TargetID: targetID, SourceID: sourceID}
	var resp api.MergeResponse
	if err := c.doRequest(http.MethodPost, "/merge", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) HealthCheck() error {
	var resp map[string]string
	if err := c.doRequest(http.MethodGet, "/health", nil, &resp); err != nil {
		return err
	}
	return nil
}
