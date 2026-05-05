package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"go-log-cleaner/common"
)

type Client struct {
	host string
	port string
}

func NewClient(host, port string) *Client {
	return &Client{
		host: host,
		port: port,
	}
}

func (c *Client) SubmitTask(req *common.CleanTaskRequest) (*common.Response, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task request: %w", err)
	}

	return c.sendRequest(common.RequestTypeSubmitTask, payload)
}

func (c *Client) GetHistory() (*common.Response, error) {
	return c.sendRequest(common.RequestTypeGetHistory, nil)
}

func (c *Client) GetTask(taskID string) (*common.Response, error) {
	payload, err := json.Marshal(map[string]string{"task_id": taskID})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task id: %w", err)
	}

	return c.sendRequest(common.RequestTypeGetTask, payload)
}

func (c *Client) ListTasks() (*common.Response, error) {
	return c.sendRequest(common.RequestTypeListTasks, nil)
}

func (c *Client) HealthCheck() (*common.Response, error) {
	return c.sendRequest(common.RequestTypeHealthCheck, nil)
}

func (c *Client) sendRequest(reqType common.RequestType, payload json.RawMessage) (*common.Response, error) {
	addr := fmt.Sprintf("%s:%s", c.host, c.port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	req := common.Request{
		Type:    reqType,
		Payload: payload,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	reader := bufio.NewReader(conn)
	respData, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resp common.Response
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}
