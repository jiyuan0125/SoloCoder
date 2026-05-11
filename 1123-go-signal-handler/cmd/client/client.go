package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/example/graceful-shutdown/pkg/api"
)

type client struct {
	baseURL    string
	httpClient *http.Client
}

func newClient(baseURL string) *client {
	return &client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *client) Status() error {
	resp, err := c.httpClient.Get(c.baseURL + "/status")
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var status api.StatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	printStatus(&status)
	return nil
}

func (c *client) Graceful() error {
	resp, err := c.httpClient.Post(c.baseURL+"/graceful", "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var result api.ShutdownResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Printf("Success: %v\nMessage: %s\n", result.Success, result.Message)
	return nil
}

func (c *client) Force() error {
	resp, err := c.httpClient.Post(c.baseURL+"/force", "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var result api.ShutdownResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Printf("Success: %v\nMessage: %s\n", result.Success, result.Message)
	return nil
}

func (c *client) Register(name string, order int) error {
	req := api.RegisterRequest{
		Name:  name,
		Order: order,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/register",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var result api.RegisterResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Printf("Success: %v\nMessage: %s\n", result.Success, result.Message)
	return nil
}

func printStatus(status *api.StatusResponse) {
	fmt.Println("==========================================")
	fmt.Println("服务状态")
	fmt.Println("==========================================")
	fmt.Printf("状态: %s\n", status.State)
	fmt.Printf("消息: %s\n", status.Message)
	if status.StartTime != "" {
		fmt.Printf("开始时间: %s\n", status.StartTime)
	}
	fmt.Println()

	if len(status.Callbacks) == 0 {
		fmt.Println("没有注册的回调")
		return
	}

	fmt.Println("------------------------------------------")
	fmt.Println("清理回调列表")
	fmt.Println("------------------------------------------")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "顺序\t名称\t状态\t耗时\t错误")
	fmt.Fprintln(w, "----\t----\t----\t----\t----")

	for _, cb := range status.Callbacks {
		duration := "-"
		if cb.Duration != "" {
			duration = cb.Duration
		}
		err := "-"
		if cb.Error != "" {
			err = cb.Error
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			cb.Order, cb.Name, cb.Status, duration, err)
	}
	w.Flush()
}

func printUsage() {
	fmt.Println("优雅退出客户端工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client <command> [options]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  status              查看服务端状态和回调列表")
	fmt.Println("  graceful            触发优雅退出")
	fmt.Println("  force               触发强制退出")
	fmt.Println("  register <name> <order>  动态注册新的清理回调")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  --url <address>     服务端地址 (默认: http://localhost:8080)")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  client status")
	fmt.Println("  client graceful")
	fmt.Println("  client force")
	fmt.Println("  client register my-callback 50")
	fmt.Println("  client --url http://192.168.1.10:8080 status")
}
