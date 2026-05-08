package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/health-aggregator/pkg/common"
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

func (c *APIClient) doRequest(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(data))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *APIClient) GetHealth() (*common.ServerHealthResponse, error) {
	var result common.ServerHealthResponse
	err := c.doRequest(http.MethodGet, "/health", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) GetAggregateStatus() (*common.AggregateStatus, error) {
	var result common.AggregateStatus
	err := c.doRequest(http.MethodGet, "/api/status", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) ListServices() ([]*common.ServiceStatus, error) {
	var result []*common.ServiceStatus
	err := c.doRequest(http.MethodGet, "/api/services", nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *APIClient) AddService(cfg *common.ServiceConfig) (*common.AddServiceResponse, error) {
	req := common.AddServiceRequest{ServiceConfig: *cfg}
	var result common.AddServiceResponse
	err := c.doRequest(http.MethodPost, "/api/services", &req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) RemoveService(name string) (*common.RemoveServiceResponse, error) {
	req := common.RemoveServiceRequest{Name: name}
	var result common.RemoveServiceResponse
	err := c.doRequest(http.MethodDelete, "/api/services", &req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) GetServiceStatus(name string) (*common.ServiceStatus, error) {
	var result common.ServiceStatus
	err := c.doRequest(http.MethodGet, "/api/services/"+name, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) GetServiceHistory(name string) (*common.ProbeHistory, error) {
	var result common.ProbeHistory
	err := c.doRequest(http.MethodGet, "/api/services/"+name+"/history", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
