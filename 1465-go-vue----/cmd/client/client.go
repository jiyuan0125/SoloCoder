package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"envmonitor/internal/common"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, path string, body interface{}, queryParams url.Values) (*common.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	fullURL := c.baseURL + path
	if len(queryParams) > 0 {
		fullURL += "?" + queryParams.Encode()
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp common.Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &apiResp, nil
}

func (c *Client) RegisterGasOutlet(id, location string) (*common.Response, error) {
	req := common.RegisterGasOutletRequest{
		ID:       id,
		Location: location,
	}
	return c.doRequest(http.MethodPost, "/api/v1/gas/outlets", req, nil)
}

func (c *Client) SubmitGasReport(outletID string, measurements map[string]float64, reportedAt *time.Time) (*common.Response, error) {
	req := common.GasReportRequest{
		OutletID:     outletID,
		Measurements: measurements,
	}
	if reportedAt != nil {
		req.ReportedAt = *reportedAt
	}
	return c.doRequest(http.MethodPost, "/api/v1/gas/reports", req, nil)
}

func (c *Client) QueryGasReports(outletID string, from, to *time.Time) (*common.Response, error) {
	params := url.Values{}
	if outletID != "" {
		params.Set("outlet_id", outletID)
	}
	if from != nil {
		params.Set("from", from.Format(time.RFC3339))
	}
	if to != nil {
		params.Set("to", to.Format(time.RFC3339))
	}
	return c.doRequest(http.MethodGet, "/api/v1/gas/reports", nil, params)
}

func (c *Client) RegisterWastewaterOutlet(id string) (*common.Response, error) {
	req := common.RegisterWastewaterOutletRequest{ID: id}
	return c.doRequest(http.MethodPost, "/api/v1/wastewater/outlets", req, nil)
}

func (c *Client) SubmitWastewaterReport(outletID string, measurements map[string]float64, reportedAt *time.Time) (*common.Response, error) {
	req := common.WastewaterReportRequest{
		OutletID:     outletID,
		Measurements: measurements,
	}
	if reportedAt != nil {
		req.ReportedAt = *reportedAt
	}
	return c.doRequest(http.MethodPost, "/api/v1/wastewater/reports", req, nil)
}

func (c *Client) QueryWastewaterReports(outletID string, from, to *time.Time) (*common.Response, error) {
	params := url.Values{}
	if outletID != "" {
		params.Set("outlet_id", outletID)
	}
	if from != nil {
		params.Set("from", from.Format(time.RFC3339))
	}
	if to != nil {
		params.Set("to", to.Format(time.RFC3339))
	}
	return c.doRequest(http.MethodGet, "/api/v1/wastewater/reports", nil, params)
}

func (c *Client) GetDailyReport(outletID string, date *time.Time) (*common.Response, error) {
	params := url.Values{}
	params.Set("outlet_id", outletID)
	if date != nil {
		params.Set("date", date.Format("2006-01-02"))
	}
	return c.doRequest(http.MethodGet, "/api/v1/wastewater/daily", nil, params)
}

func (c *Client) AddSolidWasteRecord(req *common.AddSolidWasteRequest) (*common.Response, error) {
	return c.doRequest(http.MethodPost, "/api/v1/solid-waste", req, nil)
}

func (c *Client) QuerySolidWasteRecords(category, status *int) (*common.Response, error) {
	params := url.Values{}
	if category != nil {
		params.Set("category", fmt.Sprintf("%d", *category))
	}
	if status != nil {
		params.Set("status", fmt.Sprintf("%d", *status))
	}
	return c.doRequest(http.MethodGet, "/api/v1/solid-waste", nil, params)
}

func (c *Client) QueryAlarms(entityType, level *int, entityID string, resolved *bool) (*common.Response, error) {
	params := url.Values{}
	if entityType != nil {
		params.Set("entity_type", fmt.Sprintf("%d", *entityType))
	}
	if entityID != "" {
		params.Set("entity_id", entityID)
	}
	if level != nil {
		params.Set("level", fmt.Sprintf("%d", *level))
	}
	if resolved != nil {
		params.Set("resolved", fmt.Sprintf("%t", *resolved))
	}
	return c.doRequest(http.MethodGet, "/api/v1/alarms", nil, params)
}

func (c *Client) ResolveAlarm(id string) (*common.Response, error) {
	req := common.ResolveAlarmRequest{ID: id}
	return c.doRequest(http.MethodPost, "/api/v1/alarms/resolve", req, nil)
}
