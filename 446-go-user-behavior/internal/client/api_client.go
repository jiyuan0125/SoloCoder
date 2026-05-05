package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"userbehavior/internal/shared"
)

type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *APIClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr shared.APIError
		if err := json.Unmarshal(respBody, &apiErr); err == nil {
			return nil, fmt.Errorf("API error [%d]: %s", apiErr.Code, apiErr.Message)
		}
		return nil, fmt.Errorf("API error [%d]: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *APIClient) Health() (bool, error) {
	_, err := c.doRequest(http.MethodGet, "/health", nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *APIClient) Track(req *shared.TrackRequest) (*shared.TrackResponse, error) {
	respBody, err := c.doRequest(http.MethodPost, "/track", req)
	if err != nil {
		return nil, err
	}

	var resp shared.TrackResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

func (c *APIClient) FunnelAnalysis(steps []shared.FunnelStep) (*shared.FunnelResult, error) {
	req := shared.FunnelRequest{Steps: steps}
	respBody, err := c.doRequest(http.MethodPost, "/analytics/funnel", req)
	if err != nil {
		return nil, err
	}

	var resp shared.FunnelResult
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

func (c *APIClient) RetentionAnalysis(startDate, endDate time.Time) (*shared.RetentionResult, error) {
	req := shared.RetentionRequest{
		StartDate: startDate,
		EndDate:   endDate,
	}
	respBody, err := c.doRequest(http.MethodPost, "/analytics/retention", req)
	if err != nil {
		return nil, err
	}

	var resp shared.RetentionResult
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

func (c *APIClient) PathAnalysis(fromPage, toPage string) (*shared.PathResult, error) {
	req := shared.PathAnalysisRequest{
		FromPage: fromPage,
		ToPage:   toPage,
	}
	respBody, err := c.doRequest(http.MethodPost, "/analytics/paths", req)
	if err != nil {
		return nil, err
	}

	var resp shared.PathResult
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

func (c *APIClient) GetUserProfile(userID string) (*shared.UserProfile, error) {
	respBody, err := c.doRequest(http.MethodGet, "/user/profile?user_id="+userID, nil)
	if err != nil {
		return nil, err
	}

	var resp shared.UserProfile
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

func (c *APIClient) GetRealtimeStats() (*shared.RealtimeStats, error) {
	respBody, err := c.doRequest(http.MethodGet, "/realtime/stats", nil)
	if err != nil {
		return nil, err
	}

	var resp shared.RealtimeStats
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}
