package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"generic-conn-pool/pkg/api"
)

const defaultServerAddr = "http://localhost:8080"

func printUsage() {
	fmt.Println("Usage: client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create   Create a new connection pool")
	fmt.Println("  acquire  Acquire a connection from pool")
	fmt.Println("  release  Release a connection back to pool")
	fmt.Println("  stats    Get pool statistics")
	fmt.Println()
	fmt.Println("Global Options:")
	fmt.Println("  -server <addr>   Server address (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("Create Options:")
	fmt.Println("  -target <addr>     Target address (required)")
	fmt.Println("  -min <int>         Minimum idle connections")
	fmt.Println("  -max <int>         Maximum active connections")
	fmt.Println("  -idle-timeout <seconds>   Idle timeout in seconds")
	fmt.Println("  -lifetime <seconds>       Max connection lifetime in seconds")
	fmt.Println("  -acquire-timeout <ms>     Acquire timeout in milliseconds")
	fmt.Println("  -health-check <ms>        Health check timeout in milliseconds")
	fmt.Println()
	fmt.Println("Acquire Options:")
	fmt.Println("  -pool-id <id>      Pool ID (required)")
	fmt.Println()
	fmt.Println("Release Options:")
	fmt.Println("  -conn-id <id>      Connection ID (required)")
	fmt.Println()
	fmt.Println("Stats Options:")
	fmt.Println("  -pool-id <id>      Pool ID (required)")
}

type client struct {
	serverAddr string
	httpClient *http.Client
}

func newClient(serverAddr string) *client {
	return &client{
		serverAddr: serverAddr,
		httpClient: &http.Client{},
	}
}

func (c *client) doRequest(method, path string, body interface{}, result interface{}) error {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
	}

	req, err := http.NewRequest(method, c.serverAddr+path, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(respBody))
		}
	}

	return nil
}

func cmdCreate(c *client, args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	target := fs.String("target", "", "Target address")
	min := fs.Int("min", 0, "Minimum idle connections")
	max := fs.Int("max", 0, "Maximum active connections")
	idleTimeout := fs.Int("idle-timeout", 0, "Idle timeout in seconds")
	lifetime := fs.Int("lifetime", 0, "Max lifetime in seconds")
	acquireTimeout := fs.Int("acquire-timeout", 0, "Acquire timeout in ms")
	healthCheck := fs.Int("health-check", 0, "Health check timeout in ms")

	fs.Parse(args)

	if *target == "" {
		fmt.Println("Error: -target is required")
		os.Exit(1)
	}

	req := api.CreatePoolRequest{
		TargetAddress:    *target,
		MinIdle:          *min,
		MaxActive:        *max,
		IdleTimeoutSec:   *idleTimeout,
		MaxLifetimeSec:   *lifetime,
		AcquireTimeoutMs: *acquireTimeout,
		HealthCheckMs:    *healthCheck,
	}

	var resp api.CreatePoolResponse
	if err := c.doRequest(http.MethodPost, "/pool/create", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Printf("Pool created successfully\n")
	fmt.Printf("Pool ID: %s\n", resp.Message)
}

func cmdAcquire(c *client, args []string) {
	fs := flag.NewFlagSet("acquire", flag.ExitOnError)
	poolID := fs.String("pool-id", "", "Pool ID")

	fs.Parse(args)

	if *poolID == "" {
		fmt.Println("Error: -pool-id is required")
		os.Exit(1)
	}

	var resp api.AcquireConnectionResponse
	if err := c.doRequest(http.MethodGet, "/pool/acquire?pool_id="+*poolID, nil, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Printf("Connection acquired successfully\n")
	fmt.Printf("Conn ID: %s\n", resp.ConnID)
}

func cmdRelease(c *client, args []string) {
	fs := flag.NewFlagSet("release", flag.ExitOnError)
	connID := fs.String("conn-id", "", "Connection ID")

	fs.Parse(args)

	if *connID == "" {
		fmt.Println("Error: -conn-id is required")
		os.Exit(1)
	}

	req := api.ReleaseConnectionRequest{
		ConnID: *connID,
	}

	var resp api.ReleaseConnectionResponse
	if err := c.doRequest(http.MethodPost, "/pool/release", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Println("Connection released successfully")
}

func cmdStats(c *client, args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	poolID := fs.String("pool-id", "", "Pool ID")

	fs.Parse(args)

	if *poolID == "" {
		fmt.Println("Error: -pool-id is required")
		os.Exit(1)
	}

	var resp api.PoolStatsResponse
	if err := c.doRequest(http.MethodGet, "/pool/stats?pool_id="+*poolID, nil, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Println("Pool Statistics:")
	fmt.Printf("  Min Idle:       %d\n", resp.MinIdle)
	fmt.Printf("  Max Active:     %d\n", resp.MaxActive)
	fmt.Printf("  Total Conns:    %d\n", resp.TotalConns)
	fmt.Printf("  Idle Conns:     %d\n", resp.IdleConns)
	fmt.Printf("  Active Conns:   %d\n", resp.ActiveConns)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverAddr := defaultServerAddr
	cmdStart := 1

	if len(os.Args) >= 3 && os.Args[1] == "-server" {
		serverAddr = os.Args[2]
		cmdStart = 3
	}

	if cmdStart >= len(os.Args) {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[cmdStart]
	args := os.Args[cmdStart+1:]

	for i, arg := range args {
		if arg == "-server" && i+1 < len(args) {
			serverAddr = args[i+1]
			newArgs := make([]string, 0, len(args)-2)
			newArgs = append(newArgs, args[:i]...)
			newArgs = append(newArgs, args[i+2:]...)
			args = newArgs
			break
		}
	}

	c := newClient(serverAddr)

	switch command {
	case "create":
		cmdCreate(c, args)
	case "acquire":
		cmdAcquire(c, args)
	case "release":
		cmdRelease(c, args)
	case "stats":
		cmdStats(c, args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}
