// Package main provides the config client CLI tool.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"config-loader/common"
)

type Client struct {
	serverURL string
	client    *http.Client
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: strings.TrimSuffix(serverURL, "/"),
		client:    &http.Client{},
	}
}

func (c *Client) Health() (*common.HealthResponse, error) {
	resp, err := c.client.Get(c.serverURL + common.EndpointHealth)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var healthResp common.HealthResponse
	if err := json.Unmarshal(body, &healthResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &healthResp, nil
}

func (c *Client) GetConfig(path string) (*common.ConfigResponse, error) {
	url := c.serverURL + common.EndpointConfig
	if path != "" {
		url = c.serverURL + common.EndpointConfigByPath + path
	}

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var configResp common.ConfigResponse
	if err := json.Unmarshal(body, &configResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &configResp, nil
}

func (c *Client) SetConfig(path string, value interface{}) (*common.ConfigResponse, error) {
	req := common.ConfigRequest{
		Operation: "set",
		Path:      path,
		Value:     value,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(
		c.serverURL+common.EndpointConfig,
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var configResp common.ConfigResponse
	if err := json.Unmarshal(respBody, &configResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &configResp, nil
}

func (c *Client) Reload() (*common.ConfigResponse, error) {
	resp, err := c.client.Post(
		c.serverURL+common.EndpointReload,
		"application/json",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var configResp common.ConfigResponse
	if err := json.Unmarshal(body, &configResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &configResp, nil
}

func (c *Client) WatchStatus() (*common.ConfigResponse, error) {
	resp, err := c.client.Get(c.serverURL + common.EndpointWatch)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var configResp common.ConfigResponse
	if err := json.Unmarshal(body, &configResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &configResp, nil
}

func printHelp() {
	fmt.Println(`Config Client - Command line tool for config server

Usage:
  config-client [command] [options]

Commands:
  health        Check server health status
  get           Get configuration (all or by path)
  set           Set a configuration value
  reload        Reload configuration from files
  watch         Show watch status
  help          Show this help message

Options:
  -server       Server URL (default: http://localhost:8080)

Examples:
  config-client health
  config-client get
  config-client get database.host
  config-client set app.debug true
  config-client set server.port 8080
  config-client reload
  config-client watch`)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	serverURL := flag.String("server", "http://localhost:8080", "Config server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printHelp()
		os.Exit(1)
	}

	command := args[0]
	client := NewClient(*serverURL)

	switch command {
	case "health":
		handleHealth(client)

	case "get":
		path := ""
		if len(args) > 1 {
			path = args[1]
		}
		handleGet(client, path)

	case "set":
		if len(args) < 3 {
			fmt.Println("Usage: config-client set <path> <value>")
			os.Exit(1)
		}
		path := args[1]
		valueStr := args[2]
		handleSet(client, path, valueStr)

	case "reload":
		handleReload(client)

	case "watch":
		handleWatch(client)

	case "help", "--help", "-h":
		printHelp()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func handleHealth(client *Client) {
	health, err := client.Health()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server Health:")
	fmt.Printf("  Status:       %s\n", health.Status)
	fmt.Printf("  Uptime:       %s\n", health.Uptime)
	fmt.Printf("  Version:      %s\n", health.Version)
	fmt.Printf("  Config Items: %d\n", health.ConfigCount)
}

func handleGet(client *Client, path string) {
	resp, err := client.GetConfig(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		os.Exit(1)
	}

	// Pretty print the JSON
	output, err := json.MarshalIndent(resp.Value, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to format output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}

func handleSet(client *Client, path string, valueStr string) {
	var value interface{}

	valueStr = strings.TrimSpace(valueStr)

	if valueStr == "true" || valueStr == "TRUE" || valueStr == "True" {
		value = true
	} else if valueStr == "false" || valueStr == "FALSE" || valueStr == "False" {
		value = false
	} else if strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"") {
		value = strings.Trim(valueStr, "\"")
	} else if strings.HasPrefix(valueStr, "'") && strings.HasSuffix(valueStr, "'") {
		value = strings.Trim(valueStr, "'")
	} else {
		if i, err := json.Number(valueStr).Int64(); err == nil {
			value = i
		} else if f, err := json.Number(valueStr).Float64(); err == nil {
			value = f
		} else {
			value = valueStr
		}
	}

	resp, err := client.SetConfig(path, value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Printf("Successfully set %s = %v\n", path, value)
}

func handleReload(client *Client) {
	resp, err := client.Reload()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Println("Configuration reloaded successfully")

	output, err := json.MarshalIndent(resp.Value, "", "  ")
	if err == nil {
		fmt.Println(string(output))
	}
}

func handleWatch(client *Client) {
	resp, err := client.WatchStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Println("Watch Status:")
	output, err := json.MarshalIndent(resp.Value, "", "  ")
	if err == nil {
		fmt.Println(string(output))
	}
}
