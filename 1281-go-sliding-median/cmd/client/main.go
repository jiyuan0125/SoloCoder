package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"sliding-median/internal/models"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) Push(value float64) error {
	reqBody := models.PushRequest{Value: value}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := c.client.Post(c.baseURL+"/push", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("push failed: %s", string(body))
	}

	return nil
}

func (c *Client) Median() (float64, bool, error) {
	resp, err := c.client.Get(c.baseURL + "/median")
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()

	var result models.MedianResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, false, err
	}

	if !result.Success {
		return 0, false, nil
	}

	return *result.Median, true, nil
}

func (c *Client) SetWindowSize(size int) error {
	reqBody := models.SetWindowSizeRequest{Size: size}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := c.client.Post(c.baseURL+"/window-size", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set window size failed: %s", string(body))
	}

	return nil
}

func (c *Client) Reset() error {
	resp, err := c.client.Post(c.baseURL+"/reset", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("reset failed: %s", string(body))
	}

	return nil
}

func (c *Client) Status() (*models.StatusResponse, error) {
	resp, err := c.client.Get(c.baseURL + "/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func printUsage() {
	fmt.Println("Sliding Median Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  push <value>          Push a single value")
	fmt.Println("  push-stream           Continuously push random values (0-100)")
	fmt.Println("  median                Get current median")
	fmt.Println("  watch-median          Watch median continuously")
	fmt.Println("  set-window <size>     Set window size")
	fmt.Println("  reset                 Reset all data")
	fmt.Println("  status                Show current status")
	fmt.Println()
	fmt.Println("Global Options:")
	fmt.Println("  --server <url>        Server URL (default: http://localhost:8508)")
	fmt.Println("  --interval <seconds>  Interval for stream/watch commands (default: 1)")
}

func main() {
	serverURL := flag.String("server", "http://localhost:8508", "Server URL")
	interval := flag.Int("interval", 1, "Interval in seconds for stream/watch commands")
	flag.Usage = printUsage
	flag.Parse()

	if flag.NArg() < 1 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(*serverURL)
	command := flag.Arg(0)

	switch command {
	case "push":
		if flag.NArg() < 2 {
			fmt.Println("Usage: client push <value>")
			os.Exit(1)
		}
		value, err := strconv.ParseFloat(flag.Arg(1), 64)
		if err != nil {
			fmt.Printf("Invalid value: %v\n", err)
			os.Exit(1)
		}
		if err := client.Push(value); err != nil {
			fmt.Printf("Push failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Pushed: %.2f\n", value)

	case "push-stream":
		fmt.Println("Pushing random values (0-100) every second. Press Ctrl+C to stop.")
		rand.Seed(time.Now().UnixNano())
		for {
			value := rand.Float64() * 100
			if err := client.Push(value); err != nil {
				fmt.Printf("Push failed: %v\n", err)
			} else {
				fmt.Printf("[%s] Pushed: %.2f\n", time.Now().Format("15:04:05"), value)
			}
			time.Sleep(time.Duration(*interval) * time.Second)
		}

	case "median":
		median, ok, err := client.Median()
		if err != nil {
			fmt.Printf("Get median failed: %v\n", err)
			os.Exit(1)
		}
		if !ok {
			fmt.Println("No data available")
		} else {
			fmt.Printf("Median: %.4f\n", median)
		}

	case "watch-median":
		fmt.Println("Watching median every second. Press Ctrl+C to stop.")
		for {
			median, ok, err := client.Median()
			if err != nil {
				fmt.Printf("[%s] Error: %v\n", time.Now().Format("15:04:05"), err)
			} else if !ok {
				fmt.Printf("[%s] No data available\n", time.Now().Format("15:04:05"))
			} else {
				fmt.Printf("[%s] Median: %.4f\n", time.Now().Format("15:04:05"), median)
			}
			time.Sleep(time.Duration(*interval) * time.Second)
		}

	case "set-window":
		if flag.NArg() < 2 {
			fmt.Println("Usage: client set-window <size>")
			os.Exit(1)
		}
		size, err := strconv.Atoi(flag.Arg(1))
		if err != nil || size <= 0 {
			fmt.Println("Window size must be a positive integer")
			os.Exit(1)
		}
		if err := client.SetWindowSize(size); err != nil {
			fmt.Printf("Set window size failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Window size set to: %d\n", size)

	case "reset":
		if err := client.Reset(); err != nil {
			fmt.Printf("Reset failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Data reset successfully")

	case "status":
		status, err := client.Status()
		if err != nil {
			fmt.Printf("Get status failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Status:\n")
		fmt.Printf("  Window Size: %d\n", status.WindowSize)
		fmt.Printf("  Data Count:  %d\n", status.DataCount)
		if status.Median != nil {
			fmt.Printf("  Median:      %.4f\n", *status.Median)
		} else {
			fmt.Printf("  Median:      (no data)\n")
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}
