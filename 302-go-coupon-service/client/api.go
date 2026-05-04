package main

import (
	"bytes"
	"coupon-service/pkg/api"
	"encoding/json"
	"fmt"
	"net/http"
)

type APIClient struct {
	baseURL string
	client  *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *APIClient) CreateBatch(req *api.CreateBatchRequest) (*api.CreateBatchResponse, error) {
	url := fmt.Sprintf("%s/api/batches", c.baseURL)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.CreateBatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *APIClient) UpdateBatch(req *api.UpdateBatchRequest) (*api.UpdateBatchResponse, error) {
	url := fmt.Sprintf("%s/api/batches/%s", c.baseURL, req.BatchID)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.UpdateBatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *APIClient) ClaimCoupon(req *api.ClaimCouponRequest) (*api.ClaimCouponResponse, error) {
	url := fmt.Sprintf("%s/api/claims", c.baseURL)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.ClaimCouponResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *APIClient) RedeemCoupon(req *api.RedeemCouponRequest) (*api.RedeemCouponResponse, error) {
	url := fmt.Sprintf("%s/api/redemptions", c.baseURL)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.RedeemCouponResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *APIClient) GetBatchStats(batchID string) (*api.GetBatchStatsResponse, error) {
	url := fmt.Sprintf("%s/api/batches/%s/stats", c.baseURL, batchID)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.GetBatchStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *APIClient) GetUserCoupons(userID string) (*api.GetUserCouponsResponse, error) {
	url := fmt.Sprintf("%s/api/users/%s/coupons", c.baseURL, userID)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.GetUserCouponsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *APIClient) ReturnCoupon(req *api.ReturnCouponRequest) (*api.ReturnCouponResponse, error) {
	url := fmt.Sprintf("%s/api/returns", c.baseURL)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.ReturnCouponResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
