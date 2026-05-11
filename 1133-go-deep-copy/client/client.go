package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"deepcopy/shared"
)

const serverURL = "http://localhost:8080"

type Client struct {
	baseURL string
}

func NewClient() *Client {
	return &Client{baseURL: serverURL}
}

func (c *Client) RegisterPrototype(name string, data map[string]interface{}) error {
	reqBody := &shared.RegisterPrototypeRequest{
		Name: name,
		Data: data,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.baseURL+"/api/prototypes/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return readError(resp)
	}

	return nil
}

func (c *Client) ListPrototypes() ([]string, error) {
	resp, err := http.Get(c.baseURL + "/api/prototypes")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	body, _ := io.ReadAll(resp.Body)
	var result shared.ListPrototypesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Prototypes, nil
}

func (c *Client) ClonePrototype(name string) (map[string]interface{}, error) {
	reqBody := &shared.CloneRequest{Name: name}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.baseURL+"/api/prototypes/clone", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result shared.CloneResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) DeepCopy(data map[string]interface{}) (map[string]interface{}, error) {
	reqBody := &shared.DeepCopyRequest{Data: data}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.baseURL+"/api/deepcopy", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result shared.DeepCopyResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) Compare(a, b map[string]interface{}, mode string) (bool, error) {
	reqBody := &shared.CompareRequest{
		A:    a,
		B:    b,
		Mode: mode,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return false, err
	}

	resp, err := http.Post(c.baseURL+"/api/compare", "application/json", bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, readError(resp)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result shared.CompareResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return false, err
	}

	return result.Equal, nil
}

func readError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var errResp shared.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("http status %d: %s", resp.StatusCode, string(body))
	}
	return fmt.Errorf(errResp.Error)
}
