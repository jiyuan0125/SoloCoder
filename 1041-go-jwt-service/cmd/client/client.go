package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"jwt-service/pkg/common"
)

type Client struct {
	BaseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

func (c *Client) doRequest(method, endpoint string, reqBody, respBody any) error {
	var bodyReader io.Reader
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.BaseURL+endpoint, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if err := json.Unmarshal(respData, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("API error [%s]: %s", errResp.Error, errResp.Message)
		}
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(respData))
	}

	if respBody != nil {
		if err := json.Unmarshal(respData, respBody); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) Issue(userID string, claims map[string]any) (*common.IssueTokenResponse, error) {
	req := common.IssueTokenRequest{
		UserID: userID,
		Claims: claims,
	}

	var resp common.IssueTokenResponse
	if err := c.doRequest(http.MethodPost, common.EndpointIssue, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Refresh(refreshToken string) (*common.RefreshTokenResponse, error) {
	req := common.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	var resp common.RefreshTokenResponse
	if err := c.doRequest(http.MethodPost, common.EndpointRefresh, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Validate(token string) (*common.ValidateTokenResponse, error) {
	req := common.ValidateTokenRequest{
		Token: token,
	}

	var resp common.ValidateTokenResponse
	if err := c.doRequest(http.MethodPost, common.EndpointValidate, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Revoke(tokenID string) (*common.RevokeTokenResponse, error) {
	req := common.RevokeTokenRequest{
		TokenID: tokenID,
	}

	var resp common.RevokeTokenResponse
	if err := c.doRequest(http.MethodPost, common.EndpointRevoke, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
