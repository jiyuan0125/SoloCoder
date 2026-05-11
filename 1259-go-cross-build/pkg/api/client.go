package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("server error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}

	return nil
}

func (c *Client) Build(req *BuildRequest) (*BuildResponse, error) {
	var result BuildResponse
	if err := c.doRequest("POST", "/api/build", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GenerateScript(req *ScriptRequest) (*ScriptResponse, error) {
	var result ScriptResponse
	if err := c.doRequest("POST", "/api/script", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ValidatePlatforms(req *ValidatePlatformRequest) (*ValidatePlatformResponse, error) {
	var result ValidatePlatformResponse
	if err := c.doRequest("POST", "/api/validate", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
