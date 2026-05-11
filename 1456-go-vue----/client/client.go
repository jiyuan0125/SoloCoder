package main

import (
	"archivesystem/common"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

func (c *Client) postJSON(endpoint string, reqBody interface{}, respData interface{}) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.baseURL+endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var apiResp common.Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("响应解析失败: %s", string(respBody))
	}

	if !apiResp.Success {
		return fmt.Errorf(apiResp.Message)
	}

	if respData != nil && apiResp.Data != nil {
		dataBody, _ := json.Marshal(apiResp.Data)
		json.Unmarshal(dataBody, respData)
	}
	return nil
}

func (c *Client) downloadCSV(endpoint string, reqBody interface{}, filename string) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.baseURL+endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func (c *Client) downloadCSVGet(endpoint string, filename string) error {
	resp, err := http.Get(c.baseURL + endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func (c *Client) getJSON(endpoint string, respData interface{}) error {
	resp, err := http.Get(c.baseURL + endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var apiResp common.Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("响应解析失败: %s", string(respBody))
	}

	if !apiResp.Success {
		return fmt.Errorf(apiResp.Message)
	}

	if respData != nil && apiResp.Data != nil {
		dataBody, _ := json.Marshal(apiResp.Data)
		json.Unmarshal(dataBody, respData)
	}
	return nil
}
