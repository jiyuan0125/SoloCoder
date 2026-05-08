package main

import (
	"astar-pathfinding/common"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		http:    &http.Client{},
	}
}

func (c *Client) do(method, path string, body interface{}, resp interface{}) error {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	httpResp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		type errorResp struct {
			Message string `json:"message"`
		}
		var e errorResp
		json.Unmarshal(respBody, &e)
		if e.Message != "" {
			return fmt.Errorf("server error (%d): %s", httpResp.StatusCode, e.Message)
		}
		return fmt.Errorf("server error (%d): %s", httpResp.StatusCode, string(respBody))
	}

	if resp != nil {
		if err := json.Unmarshal(respBody, resp); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) SetMap(req *common.SetMapRequest) error {
	return c.do(http.MethodPost, "/map/set", req, nil)
}

func (c *Client) RandomMap(req *common.RandomMapRequest) (*common.GetMapResponse, error) {
	type resp struct {
		Success bool                  `json:"success"`
		Map     common.GetMapResponse `json:"map"`
	}
	var r resp
	if err := c.do(http.MethodPost, "/map/random", req, &r); err != nil {
		return nil, err
	}
	return &r.Map, nil
}

func (c *Client) GetMap() (*common.GetMapResponse, error) {
	var resp common.GetMapResponse
	if err := c.do(http.MethodGet, "/map", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Pathfind(req *common.RunPathfindingRequest) (*common.PathfindingResponse, error) {
	var resp common.PathfindingResponse
	if err := c.do(http.MethodPost, "/pathfind", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetResult() (*common.PathfindingResponse, error) {
	var resp common.PathfindingResponse
	if err := c.do(http.MethodGet, "/result", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
