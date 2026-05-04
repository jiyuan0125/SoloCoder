package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"luggage-tracking/common"
)

const (
	defaultServerURL = "http://localhost:8080"
	timeout          = 10 * time.Second
)

type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	if baseURL == "" {
		baseURL = defaultServerURL
	}
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *APIClient) Scan(req common.ScanRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/scan", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

func (c *APIClient) Backfill(req common.BackfillRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/backfill", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

func (c *APIClient) GetLuggage(tagNumber string) (*common.LuggageResponse, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/luggage/%s", c.baseURL, tagNumber))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var luggage common.LuggageResponse
	if err := json.NewDecoder(resp.Body).Decode(&luggage); err != nil {
		return nil, err
	}

	return &luggage, nil
}

func (c *APIClient) GetFlightLuggages(flightNumber string, date string) (*common.FlightLuggageResponse, error) {
	url := fmt.Sprintf("%s/api/flight/%s", c.baseURL, flightNumber)
	if date != "" {
		url += fmt.Sprintf("?date=%s", date)
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var response common.FlightLuggageResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *APIClient) CancelFlight(req common.FlightCancelRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/flight/cancel", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

func (c *APIClient) GetOperationLogs() ([]common.OperationLog, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/logs/operations")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var logs []common.OperationLog
	if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
		return nil, err
	}

	return logs, nil
}

func (c *APIClient) GetAuditLogs() ([]common.AuditLog, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/logs/audit")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var logs []common.AuditLog
	if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
		return nil, err
	}

	return logs, nil
}

func (c *APIClient) Health() (bool, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

func (c *APIClient) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp common.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil {
		return fmt.Errorf("API error [%d]: %s", errResp.Code, errResp.Message)
	}

	return fmt.Errorf("API error [%d]: %s", resp.StatusCode, string(body))
}
