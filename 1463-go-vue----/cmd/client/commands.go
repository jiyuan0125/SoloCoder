package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"safetymanager/internal/api"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

func (c *Client) doRequest(method, path string, body interface{}) (*api.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}

	if reader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp api.Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return &apiResp, nil
}

func (c *Client) CreateZone(name string, parentID *string) error {
	req := api.CreateZoneRequest{
		Name:     name,
		ParentID: parentID,
	}
	resp, err := c.doRequest("POST", "/zones", &req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *Client) ListZones() error {
	resp, err := c.doRequest("GET", "/zones", nil)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *Client) CreatePlan(name, zoneID string, inspectors []string, items []string, freq api.Frequency) error {
	req := api.CreateInspectionPlanRequest{
		Name:         name,
		ZoneID:       zoneID,
		InspectorIDs: inspectors,
		Items:        items,
		Frequency:    freq,
	}
	resp, err := c.doRequest("POST", "/plans", &req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *Client) ListPlans() error {
	resp, err := c.doRequest("GET", "/plans", nil)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *Client) GenerateTasks() error {
	resp, err := c.doRequest("POST", "/tasks/generate", nil)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *Client) ListTasks() error {
	resp, err := c.doRequest("GET", "/tasks", nil)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *Client) SubmitInspection(taskID string, items []api.SubmitInspectionItem) error {
	req := api.SubmitInspectionRequest{
		TaskID: taskID,
		Items:  items,
	}
	resp, err := c.doRequest("POST", "/inspections", &req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *Client) ListHazards() error {
	resp, err := c.doRequest("GET", "/hazards", nil)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *Client) SubmitRemediation(hazardID, note string) error {
	req := api.SubmitRemediationRequest{
		HazardID: hazardID,
		Note:     note,
	}
	resp, err := c.doRequest("POST", "/hazards/remediate", &req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *Client) ReviewRemediation(hazardID string, approved bool) error {
	req := api.ReviewRemediationRequest{
		HazardID: hazardID,
		Approved: approved,
	}
	resp, err := c.doRequest("POST", "/hazards/review", &req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *Client) RequestLevelChange(hazardID string, level api.HazardLevel, reason string) error {
	req := api.RequestLevelChangeRequest{
		HazardID:      hazardID,
		ProposedLevel: level,
		Reason:        reason,
	}
	resp, err := c.doRequest("POST", "/hazards/level-change/request", &req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *Client) ReviewLevelChange(requestID string, approved bool) error {
	req := api.ReviewLevelChangeRequest{
		RequestID: requestID,
		Approved:  approved,
	}
	resp, err := c.doRequest("POST", "/hazards/level-change/review", &req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *Client) CheckEscalations() error {
	resp, err := c.doRequest("POST", "/hazards/escalations/check", nil)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *Client) ListAuditLogs() error {
	resp, err := c.doRequest("GET", "/audit-logs", nil)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *Client) ExportCriticalLogs() error {
	resp, err := c.doRequest("GET", "/audit-logs/critical", nil)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	data, _ := json.MarshalIndent(resp.Data, "", "  ")
	fmt.Println(string(data))
	return nil
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: safety-cli [command] [options]

Commands:
  zone create <name> [parent-id]
  zone list
  plan create <name> <zone-id> <inspectors> <items> <frequency>
  plan list
  task generate
  task list
  inspection submit <task-id> <items-json>
  hazard list
  hazard remediate <hazard-id> <note>
  hazard review <hazard-id> <approved>
  hazard level-change request <hazard-id> <level> <reason>
  hazard level-change review <request-id> <approved>
  hazard escalations check
  audit list
  audit export-critical

Environment:
  SAFETY_MANAGER_URL  Server URL (default: http://localhost:8080)
`)
	os.Exit(1)
}
