package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"config-hotload/pkg/api"
)

type client struct {
	serverURL string
}

func newClient(serverURL string) *client {
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "http://" + serverURL
	}
	return &client{serverURL: serverURL}
}

func (c *client) getConfig() (*api.GetConfigResponse, error) {
	resp, err := http.Get(c.serverURL + "/api/config")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.GetConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) getField(path string) (*api.GetFieldResponse, error) {
	resp, err := http.Get(c.serverURL + "/api/config/field?path=" + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.GetFieldResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) reload() (*api.ReloadResponse, error) {
	resp, err := http.Post(c.serverURL+"/api/config/reload", "application/json", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.ReloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) getHistory() (*api.HistoryResponse, error) {
	resp, err := http.Get(c.serverURL + "/api/config/history")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.HistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) watch() error {
	resp, err := http.Get(c.serverURL + "/api/config/watch")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var event api.WatchEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			for _, change := range event.Changes {
				printChange(change)
			}
		}
	}
	return nil
}

func printChange(change api.ChangeInfo) {
	var oldStr, newStr string
	if change.Old == nil {
		oldStr = "<nil>"
	} else {
		oldStr = fmt.Sprintf("%v", change.Old)
	}
	if change.New == nil {
		newStr = "<nil>"
	} else {
		newStr = fmt.Sprintf("%v", change.New)
	}

	switch change.Type {
	case "add":
		fmt.Printf("%s: <nil> -> %v\n", change.Path, change.New)
	case "delete":
		fmt.Printf("%s: %v -> <nil>\n", change.Path, change.Old)
	default:
		fmt.Printf("%s: %s -> %s\n", change.Path, oldStr, newStr)
	}
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case map[string]interface{}, []interface{}:
		data, _ := json.MarshalIndent(val, "", "  ")
		return string(data)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func main() {
	serverFlag := flag.String("server", "localhost:8080", "config server address")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	client := newClient(*serverFlag)
	command := args[0]

	switch command {
	case "get":
		resp, err := client.getConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
			os.Exit(1)
		}
		fmt.Println(formatValue(resp.Config))

	case "get-field":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: field path required")
			os.Exit(1)
		}
		resp, err := client.getField(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
			os.Exit(1)
		}
		if !resp.Exists {
			fmt.Fprintf(os.Stderr, "Field '%s' does not exist\n", args[1])
			os.Exit(1)
		}
		fmt.Println(formatValue(resp.Value))

	case "reload":
		resp, err := client.reload()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
			os.Exit(1)
		}
		if len(resp.Changes) == 0 {
			fmt.Println("No changes detected")
		} else {
			for _, change := range resp.Changes {
				printChange(change)
			}
		}

	case "history":
		resp, err := client.getHistory()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
			os.Exit(1)
		}
		if len(resp.History) == 0 {
			fmt.Println("No history records")
		} else {
			for _, record := range resp.History {
				fmt.Printf("=== %s ===\n", record.Timestamp)
				for _, change := range record.Changes {
					printChange(change)
				}
				fmt.Println()
			}
		}

	case "watch":
		fmt.Println("Watching for config changes... (Ctrl+C to exit)")
		if err := client.watch(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage: config-client [options] <command> [args]

Options:
  -server string
        config server address (default "localhost:8080")

Commands:
  get                    Get full config
  get-field <path>       Get specific field by path (e.g., database.host)
  reload                 Trigger reload and show changes
  history                Show change history
  watch                  Watch for config changes in real-time

Examples:
  config-client get
  config-client get-field database.pool.max_connections
  config-client reload
  config-client history
  config-client watch`)
}
