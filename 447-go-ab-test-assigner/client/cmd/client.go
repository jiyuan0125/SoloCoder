package cmd

import (
	"abtest/api"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	BaseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		if json.Unmarshal(data, &errResp) == nil {
			if msg, ok := errResp["message"].(string); ok {
				return nil, fmt.Errorf("%s", msg)
			}
		}
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

func (c *Client) CreateExperiment(name string, expType api.ExperimentType, controlPct, treatmentPct int) (*api.Experiment, error) {
	req := api.CreateExperimentRequest{
		Name:               name,
		Type:               expType,
		ControlPercentage:  controlPct,
		TreatmentPercentage: treatmentPct,
	}

	data, err := c.doRequest("POST", "/experiments/create", req)
	if err != nil {
		return nil, err
	}

	var resp api.CreateExperimentResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp.Experiment, nil
}

func (c *Client) StartExperiment(experimentID string) (*api.Experiment, error) {
	data, err := c.doRequest("POST", "/experiments/start/"+url.PathEscape(experimentID), nil)
	if err != nil {
		return nil, err
	}

	var resp api.StartExperimentResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp.Experiment, nil
}

func (c *Client) EndExperiment(experimentID string) (*api.Experiment, error) {
	data, err := c.doRequest("POST", "/experiments/end/"+url.PathEscape(experimentID), nil)
	if err != nil {
		return nil, err
	}

	var resp api.EndExperimentResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp.Experiment, nil
}

func (c *Client) ArchiveExperiment(experimentID string) (*api.Experiment, error) {
	data, err := c.doRequest("POST", "/experiments/archive/"+url.PathEscape(experimentID), nil)
	if err != nil {
		return nil, err
	}

	var resp api.ArchiveExperimentResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp.Experiment, nil
}

func (c *Client) UpdateTraffic(experimentID string, controlPct, treatmentPct int) (*api.Experiment, error) {
	req := api.UpdateTrafficRequest{
		ExperimentID:       experimentID,
		ControlPercentage:  controlPct,
		TreatmentPercentage: treatmentPct,
	}

	data, err := c.doRequest("POST", "/experiments/traffic", req)
	if err != nil {
		return nil, err
	}

	var resp api.UpdateTrafficResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp.Experiment, nil
}

func (c *Client) AssignUser(experimentID, userID string) (api.GroupType, error) {
	req := api.AssignUserRequest{
		ExperimentID: experimentID,
		UserID:       userID,
	}

	data, err := c.doRequest("POST", "/assign", req)
	if err != nil {
		return "", err
	}

	var resp api.AssignUserResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}

	return resp.Group, nil
}

func (c *Client) RecordMetrics(experimentID, userID string, converted bool, stayDuration float64) error {
	req := api.RecordMetricsRequest{
		ExperimentID: experimentID,
		UserID:       userID,
		Converted:    converted,
		StayDuration: stayDuration,
	}

	_, err := c.doRequest("POST", "/metrics", req)
	return err
}

func (c *Client) GetStats(experimentID string) (*api.ExperimentStatsResponse, error) {
	data, err := c.doRequest("GET", "/experiments/stats/"+url.PathEscape(experimentID), nil)
	if err != nil {
		return nil, err
	}

	var resp api.ExperimentStatsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) GetUserAssignments(userID string) ([]api.UserAssignment, error) {
	data, err := c.doRequest("GET", "/users/assignments/"+url.PathEscape(userID), nil)
	if err != nil {
		return nil, err
	}

	var resp api.GetUserAssignmentsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Assignments, nil
}

func (c *Client) ListExperiments(status *api.ExperimentStatus) ([]api.Experiment, error) {
	path := "/experiments"
	if status != nil {
		path += "?status=" + url.QueryEscape(string(*status))
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp api.ListExperimentsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Experiments, nil
}

func (c *Client) GetExperiment(experimentID string) (*api.Experiment, error) {
	data, err := c.doRequest("GET", "/experiments/"+url.PathEscape(experimentID), nil)
	if err != nil {
		return nil, err
	}

	var resp api.GetExperimentResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp.Experiment, nil
}

func (c *Client) HealthCheck() (bool, error) {
	data, err := c.doRequest("GET", "/health", nil)
	if err != nil {
		return false, err
	}

	var resp map[string]string
	if err := json.Unmarshal(data, &resp); err != nil {
		return false, err
	}

	return resp["status"] == "ok", nil
}
