package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"taxisystem/common"
)

type APIClient struct {
	baseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{baseURL: baseURL}
}

func (c *APIClient) doRequest(method, path string, body interface{}) (*common.APIResponse, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var apiResp common.APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, 响应内容: %s", err, string(respBody))
	}

	return &apiResp, nil
}

func (c *APIClient) CreateVehicle(plate, driverName, driverPhone string, lat, lng float64) (*common.APIResponse, error) {
	req := common.CreateVehicleRequest{
		PlateNumber: plate,
		DriverName:  driverName,
		DriverPhone: driverPhone,
		Latitude:    lat,
		Longitude:   lng,
	}
	return c.doRequest("POST", "/api/v1/vehicles", req)
}

func (c *APIClient) ListVehicles() (*common.APIResponse, error) {
	return c.doRequest("GET", "/api/v1/vehicles", nil)
}

func (c *APIClient) GetVehicle(plate string) (*common.APIResponse, error) {
	return c.doRequest("GET", "/api/v1/vehicles/"+plate, nil)
}

func (c *APIClient) UpdateVehicleStatus(plate string, status common.VehicleStatus) (*common.APIResponse, error) {
	req := common.UpdateVehicleStatusRequest{Status: status}
	return c.doRequest("PUT", "/api/v1/vehicles/"+plate+"/status", req)
}

func (c *APIClient) UpdateVehicleLocation(plate string, lat, lng float64) (*common.APIResponse, error) {
	req := common.UpdateVehicleLocationRequest{
		Latitude:  lat,
		Longitude: lng,
	}
	return c.doRequest("PUT", "/api/v1/vehicles/"+plate+"/location", req)
}

func (c *APIClient) EstimateFare(pickup, dest common.Location) (*common.APIResponse, error) {
	req := common.EstimateFareRequest{
		PickupLocation: pickup,
		DestLocation:   dest,
	}
	return c.doRequest("POST", "/api/v1/orders/estimate", req)
}

func (c *APIClient) CreateOrder(passengerID, passengerPhone string, pickup, dest common.Location) (*common.APIResponse, error) {
	req := common.CreateOrderRequest{
		PassengerID:    passengerID,
		PassengerPhone: passengerPhone,
		PickupLocation: pickup,
		DestLocation:   dest,
	}
	return c.doRequest("POST", "/api/v1/orders", req)
}

func (c *APIClient) ListOrders() (*common.APIResponse, error) {
	return c.doRequest("GET", "/api/v1/orders", nil)
}

func (c *APIClient) GetOrder(orderID string) (*common.APIResponse, error) {
	return c.doRequest("GET", "/api/v1/orders/"+orderID, nil)
}

func (c *APIClient) AcceptOrder(orderID, plateNumber string) (*common.APIResponse, error) {
	req := common.AcceptOrderRequest{PlateNumber: plateNumber}
	return c.doRequest("POST", "/api/v1/orders/"+orderID+"/accept", req)
}

func (c *APIClient) StartTrip(orderID string, actualDistance float64, actualDuration int64, lowSpeedMinutes int) (*common.APIResponse, error) {
	req := common.StartTripRequest{
		ActualDistance:   actualDistance,
		ActualDuration:   actualDuration,
		LowSpeedMinutes:  lowSpeedMinutes,
	}
	return c.doRequest("POST", "/api/v1/orders/"+orderID+"/start", req)
}

func (c *APIClient) CompleteTrip(orderID string) (*common.APIResponse, error) {
	return c.doRequest("POST", "/api/v1/orders/"+orderID+"/complete", nil)
}

func (c *APIClient) SubmitRating(orderID string, stars int, content string) (*common.APIResponse, error) {
	req := common.SubmitRatingRequest{
		Stars:   stars,
		Content: content,
	}
	return c.doRequest("POST", "/api/v1/orders/"+orderID+"/rating", req)
}

func (c *APIClient) ListComplaints() (*common.APIResponse, error) {
	return c.doRequest("GET", "/api/v1/complaints", nil)
}

func (c *APIClient) HandleComplaint(complaintID string) (*common.APIResponse, error) {
	return c.doRequest("POST", "/api/v1/complaints/"+complaintID+"/handle", nil)
}

func (c *APIClient) GetMetrics() (*common.APIResponse, error) {
	return c.doRequest("GET", "/api/v1/metrics", nil)
}
