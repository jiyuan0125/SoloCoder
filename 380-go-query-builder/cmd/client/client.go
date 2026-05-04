package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"querybuilder/pkg/common"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
	}
}

func (c *Client) BuildQuery(req common.QueryRequest) (*common.QueryResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.baseURL+"/api/query", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var queryResp common.QueryResponse
	err = json.NewDecoder(resp.Body).Decode(&queryResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &queryResp, nil
}
