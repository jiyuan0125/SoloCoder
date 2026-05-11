package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pingtool/pkg/model"
)

type PingClient struct {
	serverURL string
	httpClient *http.Client
}

func NewPingClient(serverURL string) *PingClient {
	return &PingClient{
		serverURL:  serverURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *PingClient) StartPing(target string, count int, interval time.Duration, ttl int) (string, error) {
	req := model.PingRequest{
		Target:   target,
		Count:    count,
		Interval: interval,
		TTL:      ttl,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.serverURL+"/api/ping", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to start ping: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var pingResp model.PingResponse
	if err := json.NewDecoder(resp.Body).Decode(&pingResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return pingResp.TaskID, nil
}

func (c *PingClient) GetStatus(taskID string) (*model.TaskStatus, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/ping/%s", c.serverURL, taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var status model.TaskStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode status: %w", err)
	}

	return &status, nil
}

func (c *PingClient) StopTask(taskID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/ping/%s", c.serverURL, taskID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to stop task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
