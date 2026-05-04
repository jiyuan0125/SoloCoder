package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"password-strength/pkg/api"
)

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: serverURL,
	}
}

func (c *Client) Evaluate(password string) (*api.EvaluateResponse, error) {
	req := api.EvaluateRequest{
		Password: password,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/evaluate", c.serverURL)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	var result api.EvaluateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

func PrintResponse(resp *api.EvaluateResponse) {
	if !resp.Success {
		fmt.Printf("错误: %s\n", resp.Error)
		return
	}

	fmt.Printf("密码强度等级: %s\n", resp.Level)
	fmt.Printf("分数: %d\n", resp.Score)

	if len(resp.Suggestions) > 0 {
		fmt.Println("\n改进建议:")
		for i, s := range resp.Suggestions {
			fmt.Printf("  %d. %s\n", i+1, s)
		}
	}
}
