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
	"time"

	"file-watch-debounce/common"
)

type Client struct {
	serverURL string
	httpClient *http.Client
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: strings.TrimRight(serverURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) AddWatch(path string, debounce, maxWait time.Duration, callbackURL string) (*common.AddWatchResponse, error) {
	req := common.AddWatchRequest{
		Path:        path,
		Debounce:    debounce,
		MaxWait:     maxWait,
		CallbackURL: callbackURL,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.serverURL+"/api/watch/add", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.AddWatchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) RemoveWatch(id string) (*common.RemoveWatchResponse, error) {
	req := common.RemoveWatchRequest{
		ID: id,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.serverURL+"/api/watch/remove", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.RemoveWatchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) ListWatches() (*common.ListWatchesResponse, error) {
	resp, err := c.httpClient.Get(c.serverURL + "/api/watches")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.ListWatchesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) ListEvents() (*common.ListEventsResponse, error) {
	resp, err := c.httpClient.Get(c.serverURL + "/api/events")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.ListEventsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func printUsage() {
	fmt.Println("Usage: client [command] [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  add <path>              Add a directory watch")
	fmt.Println("  remove <id>             Remove a watch by ID")
	fmt.Println("  list                    List all active watches")
	fmt.Println("  events                  List recent events")
	fmt.Println("  help                    Show this help message")
	fmt.Println()
	fmt.Println("Options for 'add' command:")
	fmt.Println("  --debounce duration     Debounce duration (default 500ms)")
	fmt.Println("  --max-wait duration     Max wait duration (default 5s)")
	fmt.Println("  --callback url          Callback URL for events")
	fmt.Println("  --server url            Server URL (default http://localhost:8080)")
	fmt.Println()
	fmt.Println("Options for other commands:")
	fmt.Println("  --server url            Server URL (default http://localhost:8080)")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	serverURL := "http://localhost:8080"

	switch command {
	case "add":
		addCmd := flag.NewFlagSet("add", flag.ExitOnError)
		debounce := addCmd.Duration("debounce", 500*time.Millisecond, "Debounce duration")
		maxWait := addCmd.Duration("max-wait", 5*time.Second, "Max wait duration")
		callbackURL := addCmd.String("callback", "", "Callback URL")
		serverFlag := addCmd.String("server", serverURL, "Server URL")
		
		if err := addCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}
		
		if addCmd.NArg() < 1 {
			fmt.Println("Error: path is required")
			os.Exit(1)
		}
		
		path := addCmd.Arg(0)
		client := NewClient(*serverFlag)
		
		resp, err := client.AddWatch(path, *debounce, *maxWait, *callbackURL)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		
		if resp.Success {
			fmt.Printf("Successfully added watch. ID: %s\n", resp.ID)
		} else {
			fmt.Printf("Error: %s\n", resp.Error)
			os.Exit(1)
		}

	case "remove":
		removeCmd := flag.NewFlagSet("remove", flag.ExitOnError)
		serverFlag := removeCmd.String("server", serverURL, "Server URL")
		
		if err := removeCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}
		
		if removeCmd.NArg() < 1 {
			fmt.Println("Error: watch ID is required")
			os.Exit(1)
		}
		
		id := removeCmd.Arg(0)
		client := NewClient(*serverFlag)
		
		resp, err := client.RemoveWatch(id)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		
		if resp.Success {
			fmt.Println("Successfully removed watch")
		} else {
			fmt.Printf("Error: %s\n", resp.Error)
			os.Exit(1)
		}

	case "list":
		listCmd := flag.NewFlagSet("list", flag.ExitOnError)
		serverFlag := listCmd.String("server", serverURL, "Server URL")
		
		if err := listCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}
		
		client := NewClient(*serverFlag)
		resp, err := client.ListWatches()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		
		if !resp.Success {
			fmt.Printf("Error: %s\n", resp.Error)
			os.Exit(1)
		}
		
		if len(resp.Watches) == 0 {
			fmt.Println("No active watches")
			return
		}
		
		fmt.Println("Active watches:")
		for _, w := range resp.Watches {
			fmt.Printf("  ID: %s\n", w.ID)
			fmt.Printf("    Path: %s\n", w.Path)
			fmt.Printf("    Debounce: %v\n", w.Debounce)
			fmt.Printf("    Max Wait: %v\n", w.MaxWait)
			fmt.Printf("    Active: %v\n", w.Active)
			fmt.Println()
		}

	case "events":
		eventsCmd := flag.NewFlagSet("events", flag.ExitOnError)
		serverFlag := eventsCmd.String("server", serverURL, "Server URL")
		
		if err := eventsCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}
		
		client := NewClient(*serverFlag)
		resp, err := client.ListEvents()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		
		if !resp.Success {
			fmt.Printf("Error: %s\n", resp.Error)
			os.Exit(1)
		}
		
		if len(resp.Events) == 0 {
			fmt.Println("No events recorded")
			return
		}
		
		fmt.Println("Recent events:")
		for i, e := range resp.Events {
			fmt.Printf("  %d. [%s] %s: %s\n", i+1, e.Timestamp.Format("15:04:05"), e.Type, e.Path)
			if e.OldPath != "" {
				fmt.Printf("     Old path: %s\n", e.OldPath)
			}
		}

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}
