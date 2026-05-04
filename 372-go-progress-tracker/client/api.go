package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"progress-tracker/common"
)

type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *APIClient) CreateProgress(total int64, description string) (*common.CreateProgressResponse, error) {
	req := common.CreateProgressRequest{
		Total:       total,
		Description: description,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/progress", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp common.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var result common.CreateProgressResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *APIClient) GetProgress(id string) (*common.ProgressStatus, error) {
	url := c.baseURL + "/progress"
	if id != "" {
		url += "/" + id
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var result common.ProgressStatus
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *APIClient) ListProgress() ([]common.ProgressStatus, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/progress")
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var result common.ListProgressResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Progresses, nil
}

func (c *APIClient) UpdateProgress(id string, delta int64) (*common.ProgressStatus, error) {
	req := common.UpdateProgressRequest{
		Delta: delta,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/progress/"+id+"/update", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var result common.ProgressStatus
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *APIClient) SetProgress(id string, current int64) (*common.ProgressStatus, error) {
	req := common.SetProgressRequest{
		Current: current,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/progress/"+id+"/set", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var result common.ProgressStatus
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *APIClient) CancelProgress(id string) (*common.ProgressStatus, error) {
	resp, err := c.httpClient.Post(c.baseURL+"/progress/"+id+"/cancel", "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var result common.ProgressStatus
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *APIClient) CloseProgress(id string) error {
	resp, err := c.httpClient.Post(c.baseURL+"/progress/"+id+"/close", "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("server error: %s", errResp.Error)
	}

	return nil
}

func (c *APIClient) AddSubTask(parentID string, total int64, weight float64) (*common.AddSubTaskResponse, error) {
	req := common.AddSubTaskRequest{
		ParentID: parentID,
		Total:    total,
		Weight:   weight,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/progress/subtask", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp common.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var result common.AddSubTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
