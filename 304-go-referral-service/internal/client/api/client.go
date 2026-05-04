package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"referral-service/pkg/api"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

func (c *Client) doRequest(method, endpoint string, reqBody interface{}, respBody interface{}) error {
	var buf bytes.Buffer
	if reqBody != nil {
		if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
			return err
		}
	}

	url := c.baseURL + endpoint
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var errResp api.ErrorResponse
		json.Unmarshal(body, &errResp)
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error)
	}

	if respBody != nil {
		return json.Unmarshal(body, respBody)
	}

	return nil
}

func (c *Client) GenerateReferralCode(userID string) (string, error) {
	req := api.GenerateReferralCodeRequest{UserID: userID}
	var resp api.GenerateReferralCodeResponse

	if err := c.doRequest(http.MethodPost, api.EndpointGenerateReferralCode, req, &resp); err != nil {
		return "", err
	}

	return resp.ReferralCode, nil
}

func (c *Client) BindReferral(newUserID, referralCode string) (*api.BindReferralResponse, error) {
	req := api.BindReferralRequest{
		NewUserID:    newUserID,
		ReferralCode: referralCode,
	}
	var resp api.BindReferralResponse

	if err := c.doRequest(http.MethodPost, api.EndpointBindReferral, req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) CompleteFirstOrder(userID, orderID string) (*api.CompleteFirstOrderResponse, error) {
	req := api.CompleteFirstOrderRequest{
		UserID:  userID,
		OrderID: orderID,
	}
	var resp api.CompleteFirstOrderResponse

	if err := c.doRequest(http.MethodPost, api.EndpointCompleteFirstOrder, req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) RefundFirstOrder(userID, orderID string) (*api.RefundFirstOrderResponse, error) {
	req := api.RefundFirstOrderRequest{
		UserID:  userID,
		OrderID: orderID,
	}
	var resp api.RefundFirstOrderResponse

	if err := c.doRequest(http.MethodPost, api.EndpointRefundFirstOrder, req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) GetUserStats(userID string) (*api.GetUserStatsResponse, error) {
	req := api.GetUserStatsRequest{UserID: userID}
	var resp api.GetUserStatsResponse

	if err := c.doRequest(http.MethodPost, api.EndpointGetUserStats, req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) SetRewardPoints(points int64) (*api.SetRewardPointsResponse, error) {
	req := api.SetRewardPointsRequest{Points: points}
	var resp api.SetRewardPointsResponse

	if err := c.doRequest(http.MethodPost, api.EndpointSetRewardPoints, req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) GetRewardPoints() (int64, error) {
	var resp api.GetRewardPointsResponse

	if err := c.doRequest(http.MethodGet, api.EndpointGetRewardPoints, nil, &resp); err != nil {
		return 0, err
	}

	return resp.Points, nil
}

func (c *Client) GetStats() (*api.GetStatsResponse, error) {
	var resp api.GetStatsResponse

	if err := c.doRequest(http.MethodGet, api.EndpointGetStats, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
