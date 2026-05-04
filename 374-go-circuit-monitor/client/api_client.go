package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"circuit-monitor/common"
)

type APIClient struct {
	baseURL string
	client  *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *APIClient) ListCircuits() (*common.ListCircuitsResponse, error) {
	resp, err := c.client.Get(c.baseURL + "/api/circuits")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result common.ListCircuitsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *APIClient) GetCircuit(name string) (*common.GetCircuitResponse, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/api/circuits/%s", c.baseURL, name))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result common.GetCircuitResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *APIClient) CreateCircuit(name string, config *common.CircuitConfig) error {
	req := common.CreateCircuitRequest{
		Name:   name,
		Config: config,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(c.baseURL+"/api/circuits", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return c.parseError(resp)
	}

	return nil
}

func (c *APIClient) ResetCircuit(name string) error {
	resp, err := c.client.Post(fmt.Sprintf("%s/api/circuits/%s/reset", c.baseURL, name), "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

func (c *APIClient) ForceState(name string, state common.CircuitState) error {
	req := common.ForceStateRequest{
		State: state,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(fmt.Sprintf("%s/api/circuits/%s/force-state", c.baseURL, name), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

func (c *APIClient) GetConfig(name string) (*common.GetConfigResponse, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/api/circuits/%s/config", c.baseURL, name))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result common.GetConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *APIClient) SaveState(name string) error {
	resp, err := c.client.Post(fmt.Sprintf("%s/api/circuits/%s/save", c.baseURL, name), "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

func (c *APIClient) LoadState(name string) error {
	resp, err := c.client.Post(fmt.Sprintf("%s/api/circuits/%s/load", c.baseURL, name), "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

func (c *APIClient) HealthCheck() error {
	resp, err := c.client.Get(c.baseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

func (c *APIClient) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp common.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return fmt.Errorf("server error (status %d): %s", resp.StatusCode, errResp.Error)
	}

	return fmt.Errorf("server error (status %d): %s", resp.StatusCode, string(body))
}
