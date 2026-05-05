package main

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"env-switcher/protocol"
)

type Client struct {
	address string
}

func NewClient(address string) *Client {
	return &Client{address: address}
}

func (c *Client) SendRequest(req *protocol.Request) (*protocol.Response, error) {
	conn, err := net.DialTimeout("tcp", c.address, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()
	
	reqData, err := protocol.EncodeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}
	
	_, err = conn.Write(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	
	reader := bufio.NewReader(conn)
	respData, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	resp, err := protocol.DecodeResponse(respData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return resp, nil
}
