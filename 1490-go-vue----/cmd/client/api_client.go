package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"piperepair/api"
)

type APIClient struct {
	baseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{baseURL: baseURL}
}

func (c *APIClient) doRequest(method, endpoint string, body interface{}) (*api.Response, error) {
	var reader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, reader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp api.Response
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v, 原始响应: %s", err, string(bodyBytes))
	}

	return &apiResp, nil
}

func (c *APIClient) CreateOrder(req *api.CreateRepairOrderRequest) (*api.Response, error) {
	return c.doRequest(http.MethodPost, "/api/orders/create", req)
}

func (c *APIClient) GetOrder(orderID string) (*api.Response, error) {
	endpoint := fmt.Sprintf("/api/orders/get?order_id=%s", orderID)
	return c.doRequest(http.MethodGet, endpoint, nil)
}

func (c *APIClient) ListOrders() (*api.Response, error) {
	return c.doRequest(http.MethodGet, "/api/orders/list", nil)
}

func (c *APIClient) StartProcessing(orderID, masterID string) (*api.Response, error) {
	endpoint := fmt.Sprintf("/api/orders/start?order_id=%s&master_id=%s", orderID, masterID)
	return c.doRequest(http.MethodPost, endpoint, nil)
}

func (c *APIClient) SubmitRepairRecord(req *api.SubmitRepairRecordRequest) (*api.Response, error) {
	return c.doRequest(http.MethodPost, "/api/orders/submit", req)
}

func (c *APIClient) AcceptOrder(req *api.AcceptOrderRequest) (*api.Response, error) {
	return c.doRequest(http.MethodPost, "/api/orders/accept", req)
}

func (c *APIClient) GetMasterInfo(masterID string) (*api.Response, error) {
	endpoint := fmt.Sprintf("/api/masters/info?master_id=%s", masterID)
	return c.doRequest(http.MethodGet, endpoint, nil)
}
