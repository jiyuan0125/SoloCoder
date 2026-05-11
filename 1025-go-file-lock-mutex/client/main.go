package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/solocoder/file-lock-mutex/api"
)

type Client struct {
	serverURL string
	httpClient *http.Client
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL:  serverURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) doPOST(path string, req, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpResp, err := c.httpClient.Post(
		c.serverURL+path,
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	if httpResp.StatusCode >= 400 {
		var errResp api.ErrorResponse
		_ = json.Unmarshal(respBody, &errResp)
		if errResp.Error != "" {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d", httpResp.StatusCode)
	}

	if resp != nil {
		return json.Unmarshal(respBody, resp)
	}

	return nil
}

func (c *Client) Lock(filePath, mode string, timeout time.Duration) error {
	req := api.LockRequest{
		FilePath: filePath,
		Mode:     mode,
		Timeout:  timeout,
	}

	var resp api.LockResponse
	if err := c.doPOST("/lock", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("lock failed: %s", resp.Error)
	}

	fmt.Printf("Successfully acquired %s lock on %s\n", mode, filePath)
	return nil
}

func (c *Client) Unlock(filePath string) error {
	req := api.UnlockRequest{
		FilePath: filePath,
	}

	var resp api.UnlockResponse
	if err := c.doPOST("/unlock", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("unlock failed: %s", resp.Error)
	}

	fmt.Printf("Successfully released lock on %s\n", filePath)
	return nil
}

func (c *Client) Status(filePath string) error {
	req := api.StatusRequest{
		FilePath: filePath,
	}

	var resp api.StatusResponse
	if err := c.doPOST("/status", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("status check failed: %s", resp.Error)
	}

	fmt.Printf("Status for %s:\n", filePath)
	fmt.Printf("  Lock file path: %s\n", resp.LockFilePath)
	fmt.Printf("  Is held: %v\n", resp.IsHeld)
	if resp.IsHeld {
		fmt.Printf("  Mode: %s\n", resp.Mode)
	}

	return nil
}

func main() {
	serverURL := flag.String("server", "http://localhost:8400", "server URL")
	timeout := flag.Duration("timeout", 30*time.Second, "lock acquisition timeout")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(*serverURL)

	cmd := args[0]
	switch cmd {
	case "lock":
		if len(args) < 3 {
			fmt.Println("Usage: client lock <file_path> <mode>")
			fmt.Println("  mode: exclusive|shared")
			os.Exit(1)
		}
		if err := client.Lock(args[1], args[2], *timeout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "unlock":
		if len(args) < 2 {
			fmt.Println("Usage: client unlock <file_path>")
			os.Exit(1)
		}
		if err := client.Unlock(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "status":
		if len(args) < 2 {
			fmt.Println("Usage: client status <file_path>")
			os.Exit(1)
		}
		if err := client.Status(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("File Lock Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client [flags] <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  lock <file_path> <mode>   Acquire a lock")
	fmt.Println("    mode: exclusive|shared")
	fmt.Println("  unlock <file_path>        Release a lock")
	fmt.Println("  status <file_path>        Check lock status")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -server string    Server URL (default http://localhost:8400)")
	fmt.Println("  -timeout duration Lock acquisition timeout (default 30s)")
}
