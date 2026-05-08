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

	"github.com/example/sliding-window-limiter/common"
)

const defaultServer = "http://localhost:8080"

type Client struct {
	serverURL string
	httpCli   *http.Client
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: serverURL,
		httpCli:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) doRequest(method, path string, body interface{}, result interface{}) (int, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.serverURL+path, bodyReader)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return resp.StatusCode, err
		}
	}

	return resp.StatusCode, nil
}

func (c *Client) TestEndpoint(endpoint string) (*common.RateLimitResponse, int, error) {
	var result common.RateLimitResponse
	status, err := c.doRequest("POST", "/api/test", common.TestRequest{Endpoint: endpoint}, &result)
	if err != nil {
		return nil, status, err
	}
	return &result, status, nil
}

func (c *Client) AddRule(endpoint string, rule common.LimitRuleDTO) error {
	req := common.AddRuleRequest{
		Endpoint: endpoint,
		Rule:     rule,
	}
	_, err := c.doRequest("POST", "/admin/rules", req, nil)
	return err
}

func (c *Client) RemoveRule(endpoint, ruleName string) error {
	req := common.RemoveRuleRequest{
		Endpoint: endpoint,
		RuleName: ruleName,
	}
	_, err := c.doRequest("DELETE", "/admin/rules", req, nil)
	return err
}

func (c *Client) GetStatus() (*common.AdminStatusResponse, error) {
	var result common.AdminStatusResponse
	_, err := c.doRequest("GET", "/admin/status", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Reset(endpoint string) error {
	path := "/admin/reset"
	if endpoint != "" {
		path += "?endpoint=" + endpoint
	}
	_, err := c.doRequest("POST", path, nil, nil)
	return err
}

func cmdTest(args []string) {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	endpoint := fs.String("endpoint", "/api/user", "Endpoint to test")
	count := fs.Int("count", 1, "Number of requests to send")
	concurrency := fs.Int("concurrency", 1, "Number of concurrent workers")
	interval := fs.Duration("interval", 0, "Interval between requests (e.g., 100ms)")
	server := fs.String("server", defaultServer, "Server URL")
	fs.Parse(args)

	client := NewClient(*server)

	fmt.Printf("Testing endpoint: %s\n", *endpoint)
	fmt.Printf("Sending %d requests with concurrency %d\n\n", *count, *concurrency)

	var successCount int32
	var failCount int32
	var retryAfterTotal int64
	var mu sync.Mutex

	sem := make(chan struct{}, *concurrency)
	var wg sync.WaitGroup

	startTime := time.Now()

	for i := 0; i < *count; i++ {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()

			if *interval > 0 {
				time.Sleep(time.Duration(idx) * (*interval) / time.Duration(*concurrency))
			}

			result, status, err := client.TestEndpoint(*endpoint)
			if err != nil {
				fmt.Printf("[%d] Error: %v\n", idx+1, err)
				atomic.AddInt32(&failCount, 1)
				return
			}

			if status == http.StatusOK {
				atomic.AddInt32(&successCount, 1)
				fmt.Printf("[%d] 200 OK - allowed (remaining: %d/%d)\n", idx+1, result.Remaining, result.Limit)
			} else if status == http.StatusTooManyRequests {
				atomic.AddInt32(&failCount, 1)
				mu.Lock()
				retryAfterTotal += result.RetryAfter
				mu.Unlock()
				fmt.Printf("[%d] 429 Too Many Requests - blocked by %s (retry after: %dms)\n",
					idx+1, result.RuleName, result.RetryAfter)
			} else {
				atomic.AddInt32(&failCount, 1)
				fmt.Printf("[%d] %d - %s\n", idx+1, status, result.Message)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total requests: %d\n", *count)
	fmt.Printf("Success: %d\n", successCount)
	fmt.Printf("Blocked: %d\n", failCount)
	fmt.Printf("Elapsed time: %v\n", elapsed)
	if failCount > 0 {
		avgRetry := retryAfterTotal / int64(failCount)
		fmt.Printf("Average Retry-After: %dms\n", avgRetry)
	}
}

func cmdAddRule(args []string) {
	fs := flag.NewFlagSet("add-rule", flag.ExitOnError)
	endpoint := fs.String("endpoint", "/api/user", "Endpoint name")
	name := fs.String("name", "", "Rule name (required)")
	limit := fs.Int("limit", 100, "Max requests per window")
	windowMs := fs.Int64("window-ms", 60000, "Window size in milliseconds")
	gridMs := fs.Int64("grid-ms", 1000, "Grid size in milliseconds (for grid mode)")
	mode := fs.String("mode", "grid", "Limiter mode: grid or precise")
	priority := fs.Int("priority", 0, "Rule priority (lower = higher priority)")
	server := fs.String("server", defaultServer, "Server URL")
	fs.Parse(args)

	if *name == "" {
		fmt.Println("Error: -name is required")
		os.Exit(1)
	}

	client := NewClient(*server)

	rule := common.LimitRuleDTO{
		Name:     *name,
		Limit:    *limit,
		WindowMs: *windowMs,
		GridMs:   *gridMs,
		Mode:     *mode,
		Priority: *priority,
	}

	if err := client.AddRule(*endpoint, rule); err != nil {
		fmt.Printf("Error adding rule: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Rule '%s' added to endpoint '%s'\n", *name, *endpoint)
	fmt.Printf("  Mode: %s\n", *mode)
	fmt.Printf("  Limit: %d requests per %dms\n", *limit, *windowMs)
}

func cmdRemoveRule(args []string) {
	fs := flag.NewFlagSet("remove-rule", flag.ExitOnError)
	endpoint := fs.String("endpoint", "/api/user", "Endpoint name")
	name := fs.String("name", "", "Rule name to remove (required)")
	server := fs.String("server", defaultServer, "Server URL")
	fs.Parse(args)

	if *name == "" {
		fmt.Println("Error: -name is required")
		os.Exit(1)
	}

	client := NewClient(*server)

	if err := client.RemoveRule(*endpoint, *name); err != nil {
		fmt.Printf("Error removing rule: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Rule '%s' removed from endpoint '%s'\n", *name, *endpoint)
}

func cmdStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	server := fs.String("server", defaultServer, "Server URL")
	fs.Parse(args)

	client := NewClient(*server)

	status, err := client.GetStatus()
	if err != nil {
		fmt.Printf("Error getting status: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Status at %v\n\n", status.Timestamp.Format(time.RFC3339))

	if len(status.Endpoints) == 0 {
		fmt.Println("No endpoints configured.")
		return
	}

	for endpoint, data := range status.Endpoints {
		fmt.Printf("Endpoint: %s\n", endpoint)
		fmt.Printf("  Rules:\n")
		for _, rule := range data.Rules {
			fmt.Printf("    - %s: %d req/%dms (mode: %s, priority: %d)\n",
				rule.Name, rule.Limit, rule.WindowMs, rule.Mode, rule.Priority)
		}
		fmt.Printf("  Stats:\n")
		for ruleName, statMap := range data.Stats {
			fmt.Printf("    - %s: current=%v, remaining=%v\n",
				ruleName, statMap["current"], statMap["remaining"])
		}
		fmt.Println()
	}
}

func cmdReset(args []string) {
	fs := flag.NewFlagSet("reset", flag.ExitOnError)
	endpoint := fs.String("endpoint", "", "Endpoint to reset (empty = all)")
	server := fs.String("server", defaultServer, "Server URL")
	fs.Parse(args)

	client := NewClient(*server)

	if err := client.Reset(*endpoint); err != nil {
		fmt.Printf("Error resetting: %v\n", err)
		os.Exit(1)
	}

	if *endpoint == "" {
		fmt.Println("All endpoints reset.")
	} else {
		fmt.Printf("Endpoint '%s' reset.\n", *endpoint)
	}
}

func printUsage() {
	fmt.Println("Usage: client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  test        Send test requests to an endpoint")
	fmt.Println("  add-rule    Add a rate limiting rule")
	fmt.Println("  remove-rule Remove a rate limiting rule")
	fmt.Println("  status      Show current status of all endpoints")
	fmt.Println("  reset       Reset rate limit counters")
	fmt.Println()
	fmt.Println("Use 'client <command> -h' for command-specific options")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "test":
		cmdTest(os.Args[2:])
	case "add-rule":
		cmdAddRule(os.Args[2:])
	case "remove-rule":
		cmdRemoveRule(os.Args[2:])
	case "status":
		cmdStatus(os.Args[2:])
	case "reset":
		cmdReset(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}
