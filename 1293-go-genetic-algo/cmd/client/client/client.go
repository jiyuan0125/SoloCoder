package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"genetic-algo/pkg/api"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) SubmitTask(req *api.OptimizeRequest) (*api.TaskResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/v1/optimize", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return nil, decodeError(resp)
	}

	var result api.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetTask(taskID string) (*api.TaskResult, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/v1/tasks/%s", c.baseURL, taskID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, decodeError(resp)
	}

	var result api.TaskResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetStatistics(taskID string) ([]api.Statistics, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/v1/tasks/%s/statistics", c.baseURL, taskID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, decodeError(resp)
	}

	var result []api.Statistics
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) ListFunctions() ([]api.FunctionInfo, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v1/functions")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, decodeError(resp)
	}

	var result api.ListFunctionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Functions, nil
}

func decodeError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var errResp api.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("server error: %s", string(body))
	}
	if errResp.Message != "" {
		return fmt.Errorf("%s: %s", errResp.Error, errResp.Message)
	}
	return fmt.Errorf("%s", errResp.Error)
}
