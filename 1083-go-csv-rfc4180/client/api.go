package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"csv-rfc4180/common"
)

type APIClient struct {
	BaseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{BaseURL: baseURL}
}

func (c *APIClient) Parse(csvText string, withHeader bool) (*common.ParseResponse, error) {
	req := common.ParseRequest{
		CSV:        csvText,
		WithHeader: withHeader,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.BaseURL+"/parse", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.ParseResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("invalid response: %s", string(respBody))
	}

	return &result, nil
}

func (c *APIClient) Serialize(data [][]string) (*common.SerializeResponse, error) {
	req := common.SerializeRequest{
		Data: data,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.BaseURL+"/serialize", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.SerializeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("invalid response: %s", string(respBody))
	}

	return &result, nil
}

func (c *APIClient) Validate(csvText string, withHeader bool) (*common.ValidateResponse, error) {
	req := common.ValidateRequest{
		CSV:        csvText,
		WithHeader: withHeader,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.BaseURL+"/validate", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.ValidateResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("invalid response: %s", string(respBody))
	}

	return &result, nil
}
