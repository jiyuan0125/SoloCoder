package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"lockservice/pkg/common"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) post(path string, req interface{}, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + path
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer httpResp.Body.Close()

	data, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode >= 400 {
		var errResp common.ErrorResp
		if json.Unmarshal(data, &errResp) == nil && errResp.Message != "" {
			return fmt.Errorf(errResp.Message)
		}
		return fmt.Errorf("http error: %d", httpResp.StatusCode)
	}

	if err := json.Unmarshal(data, resp); err != nil {
		return fmt.Errorf("unmarshal response: %w, body=%s", err, string(data))
	}

	return nil
}

func (c *Client) get(path string, resp interface{}) error {
	url := c.baseURL + path
	httpResp, err := c.http.Get(url)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer httpResp.Body.Close()

	data, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode >= 400 {
		var errResp common.ErrorResp
		if json.Unmarshal(data, &errResp) == nil && errResp.Message != "" {
			return fmt.Errorf(errResp.Message)
		}
		return fmt.Errorf("http error: %d", httpResp.StatusCode)
	}

	if err := json.Unmarshal(data, resp); err != nil {
		return fmt.Errorf("unmarshal response: %w, body=%s", err, string(data))
	}

	return nil
}

func getServerURL() string {
	if url := os.Getenv("SERVER_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}
