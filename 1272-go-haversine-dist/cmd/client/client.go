package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"geodist/common"
)

type APIClient struct {
	BaseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{BaseURL: baseURL}
}

func (c *APIClient) doRequest(method, path string, body interface{}, result interface{}) error {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return fmt.Errorf("api error: %s", errResp.Error)
		}
		return fmt.Errorf("api error: status %d", resp.StatusCode)
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *APIClient) Distance(pointA, pointB common.Coordinate, useEllipsoid bool) (float64, error) {
	req := common.DistanceRequest{
		PointA:       pointA,
		PointB:       pointB,
		UseEllipsoid: useEllipsoid,
	}

	var resp common.DistanceResponse
	if err := c.doRequest(http.MethodPost, "/distance", req, &resp); err != nil {
		return 0, err
	}
	return resp.DistanceKm, nil
}

func (c *APIClient) PolylineLength(points []common.Coordinate, useEllipsoid bool) (*common.PolylineLengthResponse, error) {
	req := common.PolylineRequest{
		Points:       points,
		UseEllipsoid: useEllipsoid,
	}

	var resp common.PolylineLengthResponse
	if err := c.doRequest(http.MethodPost, "/polyline/length", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) PointToPolyline(point common.Coordinate, polyline []common.Coordinate, useEllipsoid bool) (*common.PointToPolylineResponse, error) {
	req := common.PointToPolylineRequest{
		Point:        point,
		Polyline:     polyline,
		UseEllipsoid: useEllipsoid,
	}

	var resp common.PointToPolylineResponse
	if err := c.doRequest(http.MethodPost, "/polyline/closest", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) BatchDistances(center common.Coordinate, targets []common.Coordinate, useEllipsoid bool) (*common.BatchDistanceResponse, error) {
	req := common.BatchDistanceRequest{
		Center:       center,
		Targets:      targets,
		UseEllipsoid: useEllipsoid,
	}

	var resp common.BatchDistanceResponse
	if err := c.doRequest(http.MethodPost, "/batch", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) Circle(center common.Coordinate, radiusKm float64, numPoints int, useEllipsoid bool) (*common.CircleResponse, error) {
	req := common.CircleRequest{
		Center:       center,
		RadiusKm:     radiusKm,
		NumPoints:    numPoints,
		UseEllipsoid: useEllipsoid,
	}

	var resp common.CircleResponse
	if err := c.doRequest(http.MethodPost, "/circle", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) Health() error {
	type healthResp struct {
		Status string `json:"status"`
	}
	var resp healthResp
	return c.doRequest(http.MethodGet, "/health", nil, &resp)
}
