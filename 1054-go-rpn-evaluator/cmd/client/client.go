package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"rpn-evaluator/pkg/api"
)

type RPCClient struct {
	baseURL string
}

func NewRPCClient(baseURL string) *RPCClient {
	return &RPCClient{baseURL: baseURL}
}

func (c *RPCClient) post(endpoint string, req interface{}, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpResp, err := http.Post(c.baseURL+endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	data, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, resp)
}

func (c *RPCClient) Convert(expr string) (*api.ConvertResponse, error) {
	var resp api.ConvertResponse
	err := c.post("/api/convert", api.ConvertRequest{Expression: expr}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *RPCClient) Evaluate(expr string) (*api.EvaluateResponse, error) {
	var resp api.EvaluateResponse
	err := c.post("/api/evaluate", api.EvaluateRequest{Expression: expr}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *RPCClient) SetVariable(name string, value float64) (*api.SetVariableResponse, error) {
	var resp api.SetVariableResponse
	err := c.post("/api/variable/set", api.SetVariableRequest{Name: name, Value: value}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *RPCClient) GetVariable(name string) (*api.GetVariableResponse, error) {
	var resp api.GetVariableResponse
	err := c.post("/api/variable/get", api.GetVariableRequest{Name: name}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *RPCClient) ListVariables() (*api.ListVariablesResponse, error) {
	httpResp, err := http.Get(c.baseURL + "/api/variables")
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	data, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	var resp api.ListVariablesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *RPCClient) ProcessExpression(expr string) error {
	resp, err := c.Evaluate(expr)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	fmt.Printf("Expression: %s\n", expr)
	fmt.Printf("  RPN:    %s\n", resp.RPN)
	fmt.Printf("  Result: %s\n", resp.ResultStr)
	return nil
}
