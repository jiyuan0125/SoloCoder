package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"usedcar/api"
)

type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

func (c *APIClient) CreateVehicle(req api.CreateVehicleRequest) (*api.Vehicle, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/vehicles", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.CreateVehicleResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return &result.Vehicle, nil
}

func (c *APIClient) ListVehicles(status api.VehicleStatus) ([]api.Vehicle, error) {
	u := c.BaseURL + "/vehicles"
	if status != "" {
		u += "?status=" + url.QueryEscape(string(status))
	}

	resp, err := c.HTTPClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.ListVehiclesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return result.Vehicles, nil
}

func (c *APIClient) GetVehicle(vehicleID string) (*api.Vehicle, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/vehicles/" + vehicleID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.GetVehicleResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return &result.Vehicle, nil
}

func (c *APIClient) EvaluateVehicle(req api.EvaluateVehicleRequest) (*api.Evaluation, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/evaluations", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.EvaluateVehicleResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return &result.Evaluation, nil
}

func (c *APIClient) ListEvaluations(vehicleID string) ([]api.Evaluation, error) {
	u := c.BaseURL + "/evaluations?vehicle_id=" + url.QueryEscape(vehicleID)

	resp, err := c.HTTPClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.ListEvaluationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return result.Evaluations, nil
}

func (c *APIClient) PayDeposit(req api.PayDepositRequest) (*api.Deposit, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/deposits", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.PayDepositResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return &result.Deposit, nil
}

func (c *APIClient) PayFull(req api.PayFullRequest) (*api.TransferRecord, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/transfers", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.PayFullResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return &result.TransferRecord, nil
}

func fenToYuan(fen int64) string {
	yuan := float64(fen) / 100.0
	return fmt.Sprintf("%.2f", yuan)
}

func formatMileage(m float64) string {
	return strconv.FormatFloat(m, 'f', 1, 64)
}
