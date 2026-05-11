package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"coldchain/common"
)

type APIClient struct {
	baseURL string
	http    *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		http:    &http.Client{},
	}
}

func (c *APIClient) CreateTask(deviceID, cargoType, startWarehouse, endWarehouse string) (*common.CreateTaskResponse, error) {
	reqBody := common.CreateTaskRequest{
		DeviceID:       deviceID,
		CargoType:      cargoType,
		StartWarehouse: startWarehouse,
		EndWarehouse:   endWarehouse,
	}

	var resp common.CreateTaskResponse
	if err := c.doRequest(http.MethodPost, "/api/tasks", reqBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) ListTasks() (*common.ListTasksResponse, error) {
	var resp common.ListTasksResponse
	if err := c.doRequest(http.MethodGet, "/api/tasks", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) GetTask(taskID string) (*common.TaskResponse, error) {
	var resp common.TaskResponse
	if err := c.doRequest(http.MethodGet, "/api/tasks/"+taskID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) EndTask(taskID string) error {
	var resp common.SuccessResponse
	return c.doRequest(http.MethodPost, "/api/tasks/"+taskID+"/end", nil, &resp)
}

func (c *APIClient) SubmitSensorData(data *common.SubmitSensorDataRequest) (*common.SubmitSensorDataResponse, error) {
	var resp common.SubmitSensorDataResponse
	if err := c.doRequest(http.MethodPost, "/api/sensor-data", data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) GetSensorData(taskID string) (*common.ListSensorDataResponse, error) {
	var resp common.ListSensorDataResponse
	if err := c.doRequest(http.MethodGet, "/api/tasks/"+taskID+"/data", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) GetAlerts(taskID string) (*common.ListAlertsResponse, error) {
	var resp common.ListAlertsResponse
	if err := c.doRequest(http.MethodGet, "/api/tasks/"+taskID+"/alerts", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) GetStats(taskID string) (*common.TaskStatsResponse, error) {
	var resp common.TaskStatsResponse
	if err := c.doRequest(http.MethodGet, "/api/tasks/"+taskID+"/stats", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) GetReport(taskID string) (*common.TaskReportResponse, error) {
	var resp common.TaskReportResponse
	if err := c.doRequest(http.MethodGet, "/api/tasks/"+taskID+"/report", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) ExportCSV(taskID, exportType, outputPath string) error {
	url := c.baseURL + "/api/tasks/" + taskID + "/export/" + exportType
	resp, err := c.http.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(body))
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (c *APIClient) doRequest(method, path string, body interface{}, result interface{}) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("API error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
