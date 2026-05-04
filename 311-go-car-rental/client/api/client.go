package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"carrental/common"
)

const DefaultServerURL = "http://localhost:8080"

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	if serverURL == "" {
		serverURL = DefaultServerURL
	}
	return &Client{serverURL: serverURL}
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.serverURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		if err := json.Unmarshal(data, &errResp); err == nil && errResp["error"] != "" {
			return nil, fmt.Errorf("请求失败: %s", errResp["error"])
		}
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	return data, nil
}

func (c *Client) doGet(path string, params map[string]string) ([]byte, error) {
	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Add(k, v)
		}
		path = path + "?" + values.Encode()
	}

	return c.doRequest(http.MethodGet, path, nil)
}

func (c *Client) CreateCar(req *common.CreateCarRequest) (*common.Car, error) {
	data, err := c.doRequest(http.MethodPost, "/cars", req)
	if err != nil {
		return nil, err
	}

	var resp common.CreateCarResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf(resp.Error)
	}

	return resp.Car, nil
}

func (c *Client) ListCars() ([]*common.Car, error) {
	data, err := c.doGet("/cars", nil)
	if err != nil {
		return nil, err
	}

	var resp common.ListCarsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Cars, nil
}

func (c *Client) UpdateCarStatus(carID string, status common.CarStatus) (*common.Car, error) {
	req := &common.UpdateCarStatusRequest{
		CarID:  carID,
		Status: status,
	}

	data, err := c.doRequest(http.MethodPost, "/cars/status", req)
	if err != nil {
		return nil, err
	}

	var resp common.UpdateCarStatusResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf(resp.Error)
	}

	return resp.Car, nil
}

func (c *Client) CheckAvailability(carID string, pickupDate, returnDate string) (bool, error) {
	params := map[string]string{
		"car_id":       carID,
		"pickup_date":  pickupDate,
		"return_date":  returnDate,
	}

	data, err := c.doGet("/cars/availability", params)
	if err != nil {
		return false, err
	}

	var resp common.CheckAvailabilityResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return false, err
	}

	return resp.Available, nil
}

func (c *Client) CreateOrder(req *common.CreateOrderRequest) (*common.Order, error) {
	data, err := c.doRequest(http.MethodPost, "/orders", req)
	if err != nil {
		return nil, err
	}

	var resp common.CreateOrderResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf(resp.Error)
	}

	return resp.Order, nil
}

func (c *Client) GetOrder(orderID string) (*common.Order, error) {
	params := map[string]string{
		"order_id": orderID,
	}

	data, err := c.doGet("/orders", params)
	if err != nil {
		return nil, err
	}

	var resp common.GetOrderResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf(resp.Error)
	}

	return resp.Order, nil
}

func (c *Client) ListOrders(userPhone string) ([]*common.Order, error) {
	params := make(map[string]string)
	if userPhone != "" {
		params["user_phone"] = userPhone
	}

	data, err := c.doGet("/orders", params)
	if err != nil {
		return nil, err
	}

	var resp common.ListOrdersResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Orders, nil
}

func (c *Client) ReturnCar(req *common.ReturnCarRequest) (*common.Order, error) {
	data, err := c.doRequest(http.MethodPost, "/orders/return", req)
	if err != nil {
		return nil, err
	}

	var resp common.ReturnCarResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf(resp.Error)
	}

	return resp.Order, nil
}
