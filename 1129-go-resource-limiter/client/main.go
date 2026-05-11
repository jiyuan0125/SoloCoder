package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"resource-limiter/common"
)

type Client struct {
	baseURL string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return json.Unmarshal(respBody, result)
	}

	return nil
}

func (c *Client) SendRequest(key string) (bool, error) {
	req := common.RequestRequest{Key: key}
	var result common.RequestResponse
	
	err := c.doRequest("POST", "/request", req, &result)
	if err != nil {
		return false, err
	}
	
	return result.Success, nil
}

func (c *Client) GetConfig() (*common.ConfigResponse, error) {
	var result common.ConfigResponse
	err := c.doRequest("GET", "/config", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateConfig(globalLimit *int, keyLimit *int, key string) error {
	req := common.ConfigUpdateRequest{
		GlobalLimit: globalLimit,
		KeyLimit:    keyLimit,
		Key:         key,
	}
	return c.doRequest("POST", "/config", req, nil)
}

func (c *Client) GetMode() (common.LimitMode, error) {
	var result map[string]common.LimitMode
	err := c.doRequest("GET", "/mode", nil, &result)
	if err != nil {
		return "", err
	}
	return result["mode"], nil
}

func (c *Client) SetMode(mode common.LimitMode) error {
	req := common.ModeUpdateRequest{Mode: mode}
	return c.doRequest("POST", "/mode", req, nil)
}

func (c *Client) GetStats() (*common.StatsResponse, error) {
	var result common.StatsResponse
	err := c.doRequest("GET", "/stats", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	flag.Parse()

	client := NewClient(*serverURL)

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]
	remaining := args[1:]

	switch command {
	case "send":
		handleSend(client, remaining)
	case "config":
		handleConfig(client, remaining)
	case "mode":
		handleMode(client, remaining)
	case "stats":
		handleStats(client)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Resource Limiter Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client [options] <command> [arguments]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -server <url>    Server URL (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  send <key> [-n <count>] [-c <concurrency>]")
	fmt.Println("      Send requests for a specific key")
	fmt.Println()
	fmt.Println("  config [-global <limit>] [-key-limit <limit>] [-key <specific-key>]")
	fmt.Println("      View or update configuration")
	fmt.Println()
	fmt.Println("  mode [wait|fail]")
	fmt.Println("      View or switch limit mode (queue or fail fast)")
	fmt.Println()
	fmt.Println("  stats")
	fmt.Println("      View current statistics")
}

func handleSend(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client send <key> [-n <count>] [-c <concurrency>]")
		os.Exit(1)
	}

	key := args[0]
	
	sendFlagSet := flag.NewFlagSet("send", flag.ExitOnError)
	count := sendFlagSet.Int("n", 1, "Number of requests to send")
	concurrency := sendFlagSet.Int("c", 1, "Number of concurrent requests")
	sendFlagSet.Parse(args[1:])

	if *count < 1 {
		*count = 1
	}
	if *concurrency < 1 {
		*concurrency = 1
	}

	fmt.Printf("Sending %d requests for key '%s' with concurrency %d...\n", *count, key, *concurrency)

	var successCount atomic.Int64
	var failCount atomic.Int64

	var wg sync.WaitGroup
	sem := make(chan struct{}, *concurrency)

	startTime := time.Now()

	for i := 0; i < *count; i++ {
		wg.Add(1)
		sem <- struct{}{}
		
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			
			success, err := client.SendRequest(key)
			if err != nil {
				fmt.Printf("Request error: %v\n", err)
				failCount.Add(1)
				return
			}
			
			if success {
				successCount.Add(1)
			} else {
				failCount.Add(1)
			}
		}()
	}

	wg.Wait()
	duration := time.Since(startTime)

	fmt.Printf("\nResults:\n")
	fmt.Printf("  Total:   %d\n", *count)
	fmt.Printf("  Success: %d\n", successCount.Load())
	fmt.Printf("  Failed:  %d\n", failCount.Load())
	fmt.Printf("  Duration: %v\n", duration.Round(time.Millisecond))
}

func handleConfig(client *Client, args []string) {
	configFlagSet := flag.NewFlagSet("config", flag.ExitOnError)
	globalLimit := configFlagSet.Int("global", -1, "Set global limit")
	keyLimit := configFlagSet.Int("key-limit", -1, "Set key limit")
	key := configFlagSet.String("key", "", "Specific key to update limit for")
	configFlagSet.Parse(args)

	updateRequested := *globalLimit >= 0 || *keyLimit >= 0

	if updateRequested {
		var globalPtr *int
		var keyPtr *int

		if *globalLimit >= 0 {
			globalPtr = globalLimit
		}
		if *keyLimit >= 0 {
			keyPtr = keyLimit
		}

		if err := client.UpdateConfig(globalPtr, keyPtr, *key); err != nil {
			fmt.Printf("Error updating config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Configuration updated successfully")
	}

	config, err := client.GetConfig()
	if err != nil {
		fmt.Printf("Error getting config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nCurrent Configuration:")
	fmt.Printf("  Global Limit: %d\n", config.GlobalLimit)
	fmt.Printf("  Default Key Limit: %d\n", config.KeyLimit)
	fmt.Printf("  Mode: %s\n", config.Mode)
}

func handleMode(client *Client, args []string) {
	if len(args) > 0 {
		mode := common.LimitMode(args[0])
		if mode != common.ModeWait && mode != common.ModeFail {
			fmt.Println("Invalid mode. Use 'wait' or 'fail'")
			os.Exit(1)
		}

		if err := client.SetMode(mode); err != nil {
			fmt.Printf("Error setting mode: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Mode set to: %s\n", mode)
	}

	currentMode, err := client.GetMode()
	if err != nil {
		fmt.Printf("Error getting mode: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Current Mode: %s\n", currentMode)
}

func handleStats(client *Client) {
	stats, err := client.GetStats()
	if err != nil {
		fmt.Printf("Error getting stats: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Global Statistics:")
	fmt.Printf("  Mode:          %s\n", stats.Mode)
	fmt.Printf("  Capacity:      %d\n", stats.GlobalCapacity)
	fmt.Printf("  In Use:        %d\n", stats.GlobalInUse)
	fmt.Printf("  Available:     %d\n", stats.GlobalAvailable)
	fmt.Printf("  Queued:        %d\n", stats.GlobalQueued)
	fmt.Printf("  Total Rejected: %d\n", stats.GlobalRejected)

	if len(stats.KeyStats) > 0 {
		fmt.Println("\nPer-Key Statistics:")
		fmt.Printf("  %-15s %-10s %-10s %-10s %-10s\n", "Key", "In Use", "Capacity", "Available", "Rejected")
		fmt.Println("  ----------------------------------------------------------------")
		for key, ks := range stats.KeyStats {
			fmt.Printf("  %-15s %-10d %-10d %-10d %-10d\n", 
				truncate(key, 15), ks.InUse, ks.Capacity, ks.Available, ks.Rejected)
		}
	} else {
		fmt.Println("\nNo active keys")
	}
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}
