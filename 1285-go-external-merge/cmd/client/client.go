package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/example/externalsort/pkg/api"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

func (c *Client) AddData(data []int64) error {
	req := api.AddDataRequest{Data: data}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.baseURL+"/api/data", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.readError(resp)
	}
	return nil
}

func (c *Client) SetMemoryLimit(limit int) error {
	req := api.SetMemoryLimitRequest{Limit: limit}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.baseURL+"/api/config/memory", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.readError(resp)
	}
	return nil
}

func (c *Client) SetMergeWays(ways int) error {
	req := api.SetMergeWaysRequest{Ways: ways}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.baseURL+"/api/config/merge-ways", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.readError(resp)
	}
	return nil
}

func (c *Client) GetConfig() (*api.GetConfigResponse, error) {
	resp, err := http.Get(c.baseURL + "/api/config")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.readError(resp)
	}

	var result api.GetConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Sort() error {
	resp, err := http.Post(c.baseURL+"/api/sort", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.readError(resp)
	}
	return nil
}

func (c *Client) GetStats() (*api.StatsResponse, error) {
	resp, err := http.Get(c.baseURL + "/api/stats")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.readError(resp)
	}

	var result api.StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetResult() (*api.GetResultResponse, error) {
	resp, err := http.Get(c.baseURL + "/api/result")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.readError(resp)
	}

	var result api.GetResultResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Reset() error {
	resp, err := http.Post(c.baseURL+"/api/reset", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.readError(resp)
	}
	return nil
}

func (c *Client) readError(resp *http.Response) error {
	var errResp api.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &errResp)
	if errResp.Error != "" {
		return fmt.Errorf("server error: %s", errResp.Error)
	}
	return fmt.Errorf("server returned status %d", resp.StatusCode)
}
