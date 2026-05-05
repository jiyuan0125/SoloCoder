// Package main implements a command-line client for the rate limiter service.
// It communicates with the HTTP server to perform rate limiting operations.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"rate-shaper/protocol"
)

var serverURL = "http://localhost:8080"

func printUsage() {
	fmt.Println(`Rate Limiter Client

Usage:
  client allow <key> [count]     - Check if request is allowed
  client stats <key>              - Get rate limiter statistics
  client config                    - Show current configuration
  client config set [options]     - Update configuration
  client whitelist                 - List whitelisted keys
  client whitelist add <key>       - Add key to whitelist
  client whitelist remove <key>    - Remove key from whitelist
  client reset [key]               - Reset rate limiter (omit key to reset all)

Options for 'config set':
  -mode <mode>       - Rate limiting mode (fixed_window, sliding_window, smooth_window, token_bucket, leaky_bucket)
  -limit <num>       - Maximum requests per window
  -window <ms>       - Window duration in milliseconds
  -bucket-count <n>  - Number of buckets for smooth window mode
  -burst <num>       - Maximum burst for token bucket mode
  -rate <float>      - Rate per second for token/leaky bucket
  -capacity <num>    - Queue capacity for leaky bucket mode

Examples:
  client allow user123
  client allow user123 5
  client stats user123
  client config
  client config set -mode token_bucket -limit 100 -rate 10.0 -burst 50
  client whitelist add admin_user
  client reset user123`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "allow":
		handleAllow(os.Args[2:])
	case "stats":
		handleStats(os.Args[2:])
	case "config":
		handleConfig(os.Args[2:])
	case "whitelist":
		handleWhitelist(os.Args[2:])
	case "reset":
		handleReset(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleAllow(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: key is required")
		fmt.Println("Usage: client allow <key> [count]")
		os.Exit(1)
	}

	key := args[0]
	count := int64(1)

	if len(args) >= 2 {
		var err error
		count, err = strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			fmt.Printf("Error: invalid count: %v\n", err)
			os.Exit(1)
		}
	}

	req := protocol.AllowRequest{
		Key:   key,
		Count: count,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/allow", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var result protocol.AllowResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Printf("Error: failed to parse response: %v\n", err)
		fmt.Printf("Response: %s\n", string(respBody))
		os.Exit(1)
	}

	fmt.Printf("Request for key '%s' (count=%d):\n", key, count)
	fmt.Printf("  Status:   ")
	if result.Allowed {
		fmt.Println("\033[32mALLOWED\033[0m")
	} else {
		fmt.Println("\033[31mLIMITED\033[0m")
	}
	fmt.Printf("  Remaining: %d\n", result.Remaining)
	if result.WaitTimeMs > 0 {
		fmt.Printf("  Wait time: %v\n", time.Duration(result.WaitTimeMs)*time.Millisecond)
	}
	if result.ResetAt != "" {
		fmt.Printf("  Reset at:  %s\n", result.ResetAt)
	}
}

func handleStats(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: key is required")
		fmt.Println("Usage: client stats <key>")
		os.Exit(1)
	}

	key := args[0]

	resp, err := http.Get(serverURL + "/stats?key=" + key)
	if err != nil {
		fmt.Printf("Error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var result protocol.StatsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Printf("Error: failed to parse response: %v\n", err)
		fmt.Printf("Response: %s\n", string(respBody))
		os.Exit(1)
	}

	fmt.Printf("Statistics for key '%s':\n", key)
	fmt.Printf("  Mode:       %s\n", result.Mode)
	fmt.Printf("  Used:       %d\n", result.Used)
	fmt.Printf("  Limit:      %d\n", result.Limit)
	fmt.Printf("  Remaining:  %d\n", result.Remaining)
	if result.ResetAt != "" {
		fmt.Printf("  Reset at:   %s\n", result.ResetAt)
	}
}

func handleConfig(args []string) {
	if len(args) == 0 {
		resp, err := http.Get(serverURL + "/config")
		if err != nil {
			fmt.Printf("Error: failed to connect to server: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		var config map[string]interface{}
		if err := json.Unmarshal(respBody, &config); err != nil {
			fmt.Printf("Error: failed to parse response: %v\n", err)
			fmt.Printf("Response: %s\n", string(respBody))
			os.Exit(1)
		}

		fmt.Println("Current Configuration:")
		for k, v := range config {
			fmt.Printf("  %s: %v\n", k, v)
		}
		return
	}

	if args[0] == "set" {
		flags := flag.NewFlagSet("config set", flag.ExitOnError)
		mode := flags.String("mode", "", "Rate limiting mode")
		limit := flags.Int64("limit", 0, "Maximum requests per window")
		window := flags.Int64("window", 0, "Window duration in milliseconds")
		bucketCount := flags.Int("bucket-count", 0, "Number of buckets")
		burst := flags.Int64("burst", 0, "Maximum burst")
		rate := flags.Float64("rate", 0.0, "Rate per second")
		capacity := flags.Int64("capacity", 0, "Queue capacity")

		flags.Parse(args[1:])

		req := protocol.ConfigRequest{
			Mode:        *mode,
			Limit:       *limit,
			WindowMs:    *window,
			BucketCount: *bucketCount,
			Burst:       *burst,
			Rate:        *rate,
			Capacity:    *capacity,
		}

		body, err := json.Marshal(req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		httpReq, err := http.NewRequest("PUT", serverURL+"/config", bytes.NewReader(body))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			fmt.Printf("Error: failed to connect to server: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		var result protocol.ConfigResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			fmt.Printf("Error: failed to parse response: %v\n", err)
			fmt.Printf("Response: %s\n", string(respBody))
			os.Exit(1)
		}

		if result.Success {
			fmt.Println("\033[32mConfiguration updated successfully\033[0m")
		} else {
			fmt.Printf("\033[31mError: %s\033[0m\n", result.Error)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Unknown config command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func handleWhitelist(args []string) {
	if len(args) == 0 {
		resp, err := http.Get(serverURL + "/whitelist")
		if err != nil {
			fmt.Printf("Error: failed to connect to server: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		var result protocol.WhitelistResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			fmt.Printf("Error: failed to parse response: %v\n", err)
			fmt.Printf("Response: %s\n", string(respBody))
			os.Exit(1)
		}

		fmt.Println("Whitelisted keys:")
		if len(result.Keys) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, k := range result.Keys {
				fmt.Printf("  - %s\n", k)
			}
		}
		return
	}

	switch args[0] {
	case "add":
		if len(args) < 2 {
			fmt.Println("Error: key is required")
			os.Exit(1)
		}
		key := args[1]
		req := protocol.WhitelistRequest{Key: key}
		body, _ := json.Marshal(req)
		resp, err := http.Post(serverURL+"/whitelist", "application/json", bytes.NewReader(body))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		var result protocol.WhitelistResponse
		json.Unmarshal(respBody, &result)
		if result.Success {
			fmt.Printf("\033[32mKey '%s' added to whitelist\033[0m\n", key)
		} else {
			fmt.Printf("\033[31mError: %s\033[0m\n", result.Error)
		}

	case "remove":
		if len(args) < 2 {
			fmt.Println("Error: key is required")
			os.Exit(1)
		}
		key := args[1]
		req := protocol.WhitelistRequest{Key: key}
		body, _ := json.Marshal(req)
		httpReq, _ := http.NewRequest("DELETE", serverURL+"/whitelist", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		var result protocol.WhitelistResponse
		json.Unmarshal(respBody, &result)
		if result.Success {
			fmt.Printf("\033[32mKey '%s' removed from whitelist\033[0m\n", key)
		} else {
			fmt.Printf("\033[31mError: %s\033[0m\n", result.Error)
		}

	default:
		fmt.Printf("Unknown whitelist command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func handleReset(args []string) {
	key := ""
	if len(args) > 0 {
		key = args[0]
	}

	req := protocol.ResetRequest{Key: key}
	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/reset", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result protocol.ResetResponse
	json.Unmarshal(respBody, &result)

	if result.Success {
		if key == "" {
			fmt.Println("\033[32mAll rate limiters reset\033[0m")
		} else {
			fmt.Printf("\033[32mRate limiter for key '%s' reset\033[0m\n", key)
		}
	} else {
		fmt.Printf("\033[31mError: %s\033[0m\n", result.Error)
	}
}
