package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"realtime-dashboard/common"
)

type APIClient struct {
	serverURL string
	client    *http.Client
}

func NewAPIClient(serverURL string) *APIClient {
	return &APIClient{
		serverURL: serverURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *APIClient) GetServerStatus(includeDetails bool) (*common.ServerStatus, error) {
	url := fmt.Sprintf("%s/api/status", c.serverURL)
	if includeDetails {
		url += "?details=true"
	}

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var status common.ServerStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (c *APIClient) CreateMetric(key string, metadata common.MetricMetadata) error {
	url := fmt.Sprintf("%s/api/metrics", c.serverURL)

	req := map[string]interface{}{
		"key":      key,
		"metadata": metadata,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}

func (c *APIClient) UpdateMetricValue(key string, value float64, timestamp *time.Time) error {
	url := fmt.Sprintf("%s/api/metrics/%s/value", c.serverURL, key)

	req := map[string]interface{}{
		"value": value,
	}
	if timestamp != nil {
		req["timestamp"] = *timestamp
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}

func (c *APIClient) GetMetric(key string) (*common.Metric, error) {
	url := fmt.Sprintf("%s/api/metrics/%s", c.serverURL, key)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var metric common.Metric
	if err := json.NewDecoder(resp.Body).Decode(&metric); err != nil {
		return nil, err
	}

	return &metric, nil
}

func (c *APIClient) GetAllMetrics() ([]*common.Metric, error) {
	url := fmt.Sprintf("%s/api/metrics", c.serverURL)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var metrics []*common.Metric
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (c *APIClient) SetMetricThreshold(key string, threshold *common.AlertThreshold) error {
	url := fmt.Sprintf("%s/api/metrics/%s/threshold", c.serverURL, key)

	body, err := json.Marshal(threshold)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}

func (c *APIClient) GetMetricHistory(key string, startTime, endTime *time.Time) ([]common.MetricDataPoint, error) {
	baseURL := fmt.Sprintf("%s/api/metrics/%s/history", c.serverURL, key)

	query := url.Values{}
	if startTime != nil {
		query.Add("start", startTime.Format(time.RFC3339))
	}
	if endTime != nil {
		query.Add("end", endTime.Format(time.RFC3339))
	}

	fullURL := baseURL
	if len(query) > 0 {
		fullURL = baseURL + "?" + query.Encode()
	}

	resp, err := c.client.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var history []common.MetricDataPoint
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, err
	}

	return history, nil
}

func (c *APIClient) GetMetricTrend(key string) ([]common.TrendDataPoint, error) {
	url := fmt.Sprintf("%s/api/metrics/%s/trend", c.serverURL, key)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var trend []common.TrendDataPoint
	if err := json.NewDecoder(resp.Body).Decode(&trend); err != nil {
		return nil, err
	}

	return trend, nil
}

func (c *APIClient) GetMetricComparison(key string) (*common.ComparisonData, error) {
	url := fmt.Sprintf("%s/api/metrics/%s/comparison", c.serverURL, key)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var comparison common.ComparisonData
	if err := json.NewDecoder(resp.Body).Decode(&comparison); err != nil {
		return nil, err
	}

	return &comparison, nil
}

func (c *APIClient) SaveDashboardLayout(clientID string, metrics []string, order []int) error {
	url := fmt.Sprintf("%s/api/dashboard/layout", c.serverURL)

	layout := common.DashboardLayout{
		ClientID: clientID,
		Metrics:  metrics,
		Order:    order,
	}

	body, err := json.Marshal(layout)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}

func (c *APIClient) GetDashboardLayout(clientID string) (*common.DashboardLayout, error) {
	url := fmt.Sprintf("%s/api/dashboard/layout?client_id=%s", c.serverURL, clientID)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var layout common.DashboardLayout
	if err := json.NewDecoder(resp.Body).Decode(&layout); err != nil {
		return nil, err
	}

	return &layout, nil
}

func (c *APIClient) GetDelayStats() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/delay/stats", c.serverURL)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}

	return stats, nil
}

func (c *APIClient) GetAlerts() ([]common.Alert, error) {
	url := fmt.Sprintf("%s/api/alerts", c.serverURL)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var alerts []common.Alert
	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		return nil, err
	}

	return alerts, nil
}

func (c *APIClient) CleanupOldData() error {
	url := fmt.Sprintf("%s/api/admin/cleanup", c.serverURL)

	resp, err := c.client.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}
