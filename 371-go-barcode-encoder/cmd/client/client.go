package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"barcode-encoder/pkg/protocol"
)

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: serverURL,
	}
}

func (c *Client) Encode(input string) (*protocol.EncodeResponse, error) {
	req := protocol.EncodeRequest{
		Input: input,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/encode", c.serverURL),
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var response protocol.EncodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}
