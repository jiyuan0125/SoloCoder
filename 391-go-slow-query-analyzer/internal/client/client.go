package client

import (
	"fmt"
	"os"

	"slowquery/internal/parser"
	"slowquery/protocol"
)

type Client struct {
	serverAddr string
}

func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
	}
}

func (c *Client) StartAnalysis(logFilePath string, formatStr string, minExecTimeMs float64) error {
	format, err := parser.ParseFormatString(formatStr)
	if err != nil {
		return fmt.Errorf("invalid format: %v", err)
	}

	conn, err := protocol.Dial(c.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	req := &protocol.Request{
		Command:       protocol.CommandTypeStart,
		LogFilePath:   logFilePath,
		LogFormat:     format,
		MinExecTimeMs: minExecTimeMs,
	}

	if err := conn.SendRequest(req); err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	resp, err := conn.ReceiveResponse()
	if err != nil {
		return fmt.Errorf("failed to receive response: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Error)
	}

	fmt.Fprintf(os.Stderr, "Analysis started: %s\n", resp.Status)
	return nil
}

func (c *Client) StopAnalysis() error {
	conn, err := protocol.Dial(c.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	req := &protocol.Request{
		Command: protocol.CommandTypeStop,
	}

	if err := conn.SendRequest(req); err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	resp, err := conn.ReceiveResponse()
	if err != nil {
		return fmt.Errorf("failed to receive response: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Error)
	}

	fmt.Fprintf(os.Stderr, "Analysis stopped: %s\n", resp.Status)
	return nil
}

func (c *Client) GetStatus() (string, error) {
	conn, err := protocol.Dial(c.serverAddr)
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	req := &protocol.Request{
		Command: protocol.CommandTypeStatus,
	}

	if err := conn.SendRequest(req); err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}

	resp, err := conn.ReceiveResponse()
	if err != nil {
		return "", fmt.Errorf("failed to receive response: %v", err)
	}

	if !resp.Success {
		return "", fmt.Errorf("server error: %s", resp.Error)
	}

	return resp.Status, nil
}

func (c *Client) GetStats(topN int, sortBy protocol.SortType) (*protocol.Statistics, error) {
	conn, err := protocol.Dial(c.serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	req := &protocol.Request{
		Command: protocol.CommandTypeGetStats,
		TopN:    topN,
		SortBy:  sortBy,
	}

	if err := conn.SendRequest(req); err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	resp, err := conn.ReceiveResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to receive response: %v", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("server error: %s", resp.Error)
	}

	return resp.Stats, nil
}
