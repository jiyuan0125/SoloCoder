package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func (c *APIClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode >= 400 {
		return data, fmt.Errorf("server returned error status: %d", resp.StatusCode)
	}

	return data, nil
}

func (c *APIClient) Post(path string, body interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPost, path, body)
}

func (c *APIClient) Get(path string) ([]byte, error) {
	return c.doRequest(http.MethodGet, path, nil)
}

func (c *APIClient) PostWithQuery(path string, body interface{}, query url.Values) ([]byte, error) {
	fullPath := path
	if len(query) > 0 {
		fullPath = path + "?" + query.Encode()
	}
	return c.doRequest(http.MethodPost, fullPath, body)
}
