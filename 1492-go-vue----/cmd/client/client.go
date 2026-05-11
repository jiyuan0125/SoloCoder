package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Client struct {
	baseURL string
	httpCl  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpCl:  &http.Client{},
	}
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpCl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server error: %s - %s", resp.Status, string(respData))
	}

	return respData, nil
}

func prettyPrint(data []byte) error {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		fmt.Println(string(data))
		return err
	}
	fmt.Println(prettyJSON.String())
	return nil
}

func getBaseURL() string {
	if envURL := os.Getenv("SERVER_URL"); envURL != "" {
		return envURL
	}
	return "http://localhost:9012"
}
