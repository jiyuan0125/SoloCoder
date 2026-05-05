package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bankcard/common"
)

const (
	defaultServerURL = "http://localhost:8080"
	timeout          = 10 * time.Second
)

type Client struct {
	serverURL string
	httpClient *http.Client
}

func NewClient(serverURL string) *Client {
	if serverURL == "" {
		serverURL = defaultServerURL
	}
	return &Client{
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) ParseCard(cardNumber string) (*common.ParseResponse, error) {
	reqBody := common.ParseRequest{
		CardNumber: cardNumber,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.serverURL+"/parse",
		"application/json",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var result common.ParseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
