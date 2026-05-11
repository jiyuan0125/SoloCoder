package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"firemanagement/pkg/api"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = os.Getenv("SERVER_URL")
	}
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{},
	}
}

func (c *Client) do(method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return err
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp api.Response
	if len(respData) > 0 {
		if err := json.Unmarshal(respData, &apiResp); err != nil {
			return fmt.Errorf("response parse error: %v (raw: %s)", err, string(respData))
		}
	}

	if !apiResp.Success {
		if apiResp.Error != nil {
			return errors.New(apiResp.Error.Message)
		}
		return errors.New(http.StatusText(resp.StatusCode))
	}

	if result != nil && apiResp.Data != nil {
		dataBytes, err := json.Marshal(apiResp.Data)
		if err != nil {
			return err
		}
		return json.Unmarshal(dataBytes, result)
	}

	return nil
}

func (c *Client) CreateDevice(req api.CreateDeviceRequest) (*api.IDResponse, error) {
	var result api.IDResponse
	err := c.do(http.MethodPost, "/api/devices", req, &result)
	return &result, err
}

func (c *Client) GetDevice(id string) (*api.DeviceResponse, error) {
	var result api.DeviceResponse
	err := c.do(http.MethodGet, "/api/devices/"+id, nil, &result)
	return &result, err
}

func (c *Client) ListDevices(req api.ListDevicesRequest) (*api.DevicesResponse, error) {
	var result api.DevicesResponse
	err := c.do(http.MethodGet, "/api/devices", req, &result)
	return &result, err
}

func (c *Client) UpdateDevice(id string, req api.UpdateDeviceRequest) error {
	return c.do(http.MethodPut, "/api/devices/"+id, req, nil)
}

func (c *Client) CreateInspectionPoint(req api.CreateInspectionPointRequest) (*api.IDResponse, error) {
	var result api.IDResponse
	err := c.do(http.MethodPost, "/api/inspection/points", req, &result)
	return &result, err
}

func (c *Client) ListInspectionPoints() (*api.InspectionPointsResponse, error) {
	var result api.InspectionPointsResponse
	err := c.do(http.MethodGet, "/api/inspection/points", nil, &result)
	return &result, err
}

func (c *Client) CreateInspectionRoute(req api.CreateInspectionRouteRequest) (*api.IDResponse, error) {
	var result api.IDResponse
	err := c.do(http.MethodPost, "/api/inspection/routes", req, &result)
	return &result, err
}

func (c *Client) ListInspectionRoutes() (*api.InspectionRoutesResponse, error) {
	var result api.InspectionRoutesResponse
	err := c.do(http.MethodGet, "/api/inspection/routes", nil, &result)
	return &result, err
}

func (c *Client) CreateInspectionPlan(req api.CreateInspectionPlanRequest) (*api.IDResponse, error) {
	var result api.IDResponse
	err := c.do(http.MethodPost, "/api/inspection/plans", req, &result)
	return &result, err
}

func (c *Client) ListInspectionPlans() (*api.InspectionPlansResponse, error) {
	var result api.InspectionPlansResponse
	err := c.do(http.MethodGet, "/api/inspection/plans", nil, &result)
	return &result, err
}

func (c *Client) ListInspectionTasks() (*api.InspectionTasksResponse, error) {
	var result api.InspectionTasksResponse
	err := c.do(http.MethodGet, "/api/inspection/tasks", nil, &result)
	return &result, err
}

func (c *Client) GetInspectionTask(id string) (*api.InspectionTaskResponse, error) {
	var result api.InspectionTaskResponse
	err := c.do(http.MethodGet, "/api/inspection/tasks/"+id, nil, &result)
	return &result, err
}

func (c *Client) CheckTaskPoint(req api.CheckTaskPointRequest) error {
	return c.do(http.MethodPost, "/api/inspection/tasks/check", req, nil)
}

func (c *Client) GetTaskSummary(id string) (*api.TaskSummary, error) {
	var result api.TaskSummary
	err := c.do(http.MethodGet, "/api/inspection/tasks/"+id+"/summary", nil, &result)
	return &result, err
}

func (c *Client) CreateDrillPlan(req api.CreateDrillPlanRequest) (*api.IDResponse, error) {
	var result api.IDResponse
	err := c.do(http.MethodPost, "/api/drills/plans", req, &result)
	return &result, err
}

func (c *Client) ListDrillPlans() (*api.DrillPlansResponse, error) {
	var result api.DrillPlansResponse
	err := c.do(http.MethodGet, "/api/drills/plans", nil, &result)
	return &result, err
}

func (c *Client) CompleteDrill(planID string, req api.CompleteDrillRequest) (*api.IDResponse, error) {
	var result api.IDResponse
	err := c.do(http.MethodPost, "/api/drills/plans/"+planID+"/complete", req, &result)
	return &result, err
}

func (c *Client) ListDrillRecords() (*api.DrillRecordsResponse, error) {
	var result api.DrillRecordsResponse
	err := c.do(http.MethodGet, "/api/drills/records", nil, &result)
	return &result, err
}

func (c *Client) ListReminders(req api.ListRemindersRequest) (*api.RemindersResponse, error) {
	var result api.RemindersResponse
	err := c.do(http.MethodGet, "/api/reminders", req, &result)
	return &result, err
}

func (c *Client) MarkReminderRead(id string, read bool) error {
	return c.do(http.MethodPut, "/api/reminders/"+id+"/read", api.MarkReminderReadRequest{Read: read}, nil)
}
