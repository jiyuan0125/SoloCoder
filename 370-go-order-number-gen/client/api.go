package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ordergen/protocol"
)

type Client struct {
	serverAddr string
	httpClient *http.Client
}

type OrderInfo struct {
	OrderNo   string
	Prefix    string
	Timestamp string
	SerialNum int
	RandomNum int
}

func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
		httpClient: &http.Client{},
	}
}

func (c *Client) Generate(prefix string) (string, error) {
	reqBody := protocol.GenerateRequest{
		Prefix: prefix,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.serverAddr+"/generate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result protocol.GenerateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("%s", result.Error)
	}

	return result.OrderNo, nil
}

func (c *Client) Parse(orderNo string) (*OrderInfo, error) {
	reqBody := protocol.ParseRequest{
		OrderNo: orderNo,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.serverAddr+"/parse",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result protocol.ParseResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("%s", result.Error)
	}

	return &OrderInfo{
		OrderNo:   orderNo,
		Prefix:    result.Prefix,
		Timestamp: result.Timestamp,
		SerialNum: result.SerialNum,
		RandomNum: result.RandomNum,
	}, nil
}

func (c *Client) GetPrefix() (string, error) {
	resp, err := c.httpClient.Get(c.serverAddr + "/prefix")
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result protocol.GetPrefixResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("%s", result.Error)
	}

	return result.Prefix, nil
}

func (c *Client) SetPrefix(prefix string) error {
	reqBody := protocol.SetPrefixRequest{
		Prefix: prefix,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, c.serverAddr+"/prefix", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result protocol.SetPrefixResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Error)
	}

	return nil
}
