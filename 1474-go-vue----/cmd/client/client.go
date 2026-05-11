package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"bikeshare/common"
)

type Client struct {
	config *Config
}

func NewClient(config *Config) *Client {
	return &Client{config: config}
}

func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, c.config.ServerURL+path, bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp common.Response
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("API error: %s", apiResp.Error)
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

func (c *Client) ListFences() ([]*common.FenceResponse, error) {
	var result []*common.FenceResponse
	err := c.doRequest("GET", "/fences", nil, &result)
	return result, err
}

func (c *Client) CreateFence(name, fenceType string, minLat, maxLat, minLng, maxLng float64, capacity int) (*common.FenceResponse, error) {
	req := common.CreateFenceRequest{
		Name:     name,
		Type:     fenceType,
		MinLat:   minLat,
		MaxLat:   maxLat,
		MinLng:   minLng,
		MaxLng:   maxLng,
		Capacity: capacity,
	}
	var result common.FenceResponse
	err := c.doRequest("POST", "/fences", req, &result)
	return &result, err
}

func (c *Client) GetFence(id string) (*common.FenceResponse, error) {
	var result common.FenceResponse
	err := c.doRequest("GET", "/fences/"+id, nil, &result)
	return &result, err
}

func (c *Client) DeleteFence(id string) (map[string]string, error) {
	var result map[string]string
	err := c.doRequest("DELETE", "/fences/"+id, nil, &result)
	return result, err
}

func (c *Client) ListBikes() ([]*common.BikeResponse, error) {
	var result []*common.BikeResponse
	err := c.doRequest("GET", "/bikes", nil, &result)
	return result, err
}

func (c *Client) AddBike(lat, lng float64) (*common.BikeResponse, error) {
	req := common.AddBikeRequest{Lat: lat, Lng: lng}
	var result common.BikeResponse
	err := c.doRequest("POST", "/bikes", req, &result)
	return &result, err
}

func (c *Client) GetBike(id string) (*common.BikeResponse, error) {
	var result common.BikeResponse
	err := c.doRequest("GET", "/bikes/"+id, nil, &result)
	return &result, err
}

func (c *Client) MoveBike(id string, lat, lng float64) (*common.BikeResponse, error) {
	req := common.UpdateBikeLocationRequest{BikeID: id, Lat: lat, Lng: lng}
	var result common.BikeResponse
	err := c.doRequest("PUT", "/bikes/"+id, req, &result)
	return &result, err
}

func (c *Client) ListRides(activeOnly bool) ([]*common.RideResponse, error) {
	path := "/rides"
	if activeOnly {
		path += "?active=true"
	}
	var result []*common.RideResponse
	err := c.doRequest("GET", path, nil, &result)
	return result, err
}

func (c *Client) StartRide(bikeID string, startLat, startLng float64) (*common.RideResponse, error) {
	req := common.StartRideRequest{
		BikeID:   bikeID,
		StartLat: startLat,
		StartLng: startLng,
	}
	var result common.RideResponse
	err := c.doRequest("POST", "/rides/start", req, &result)
	return &result, err
}

func (c *Client) EndRide(rideID string, endLat, endLng float64) (*common.RideResponse, error) {
	req := common.EndRideRequest{
		RideID: rideID,
		EndLat: endLat,
		EndLng: endLng,
	}
	var result common.RideResponse
	err := c.doRequest("POST", "/rides/end", req, &result)
	return &result, err
}

func (c *Client) AnalyzeCapacity() ([]*common.FenceCapacityResponse, error) {
	var result []*common.FenceCapacityResponse
	err := c.doRequest("GET", "/dispatch/analyze", nil, &result)
	return result, err
}

func (c *Client) ListDispatchTasks(pendingOnly bool) ([]*common.DispatchTaskResponse, error) {
	path := "/dispatch/tasks"
	if pendingOnly {
		path += "?pending=true"
	}
	var result []*common.DispatchTaskResponse
	err := c.doRequest("GET", path, nil, &result)
	return result, err
}

func (c *Client) GenerateDispatchTasks() ([]*common.DispatchTaskResponse, error) {
	var result []*common.DispatchTaskResponse
	err := c.doRequest("POST", "/dispatch/tasks", nil, &result)
	return result, err
}

func (c *Client) CompleteDispatchTask(taskID string) (*common.DispatchTaskResponse, error) {
	var result common.DispatchTaskResponse
	err := c.doRequest("POST", "/dispatch/tasks/"+taskID+"?action=complete", nil, &result)
	return &result, err
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		fmt.Printf("Warning: failed to parse float '%s', using 0\n", s)
		return 0
	}
	return v
}

func parseInt(s string) int {
	s = strings.TrimSpace(s)
	v, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("Warning: failed to parse int '%s', using 0\n", s)
		return 0
	}
	return v
}
