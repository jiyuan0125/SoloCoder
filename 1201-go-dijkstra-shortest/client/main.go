package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/example/shortestpath/common"
)

const defaultServerURL = "http://localhost:8080"

func printUsage() {
	fmt.Println("Shortest Path Command Line Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client import <graph-file.json> [--server=<url>]")
	fmt.Println("  client path <start> <end> [--server=<url>]")
	fmt.Println("  client allpaths <start> [--server=<url>]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --server=<url>   Server URL (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client import graph.json")
	fmt.Println("  client path A D")
	fmt.Println("  client allpaths A")
}

func getServerURL(args []string) (string, []string) {
	for i, arg := range args {
		if strings.HasPrefix(arg, "--server=") {
			url := strings.TrimPrefix(arg, "--server=")
			remaining := append(args[:i], args[i+1:]...)
			return url, remaining
		}
	}
	return defaultServerURL, args
}

func postJSON(url string, data interface{}, result interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func handleImport(args []string, serverURL string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing graph file path")
	}

	filePath := args[0]
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var graphData common.GraphData
	if err := json.Unmarshal(data, &graphData); err != nil {
		return fmt.Errorf("invalid JSON file: %w", err)
	}

	var result common.ImportGraphResponse
	if err := postJSON(serverURL+"/graph/import", graphData, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("import failed: %s", result.Message)
	}

	fmt.Printf("✓ %s\n", result.Message)
	return nil
}

func handlePath(args []string, serverURL string) error {
	if len(args) < 2 {
		return fmt.Errorf("missing start and end nodes")
	}

	start := args[0]
	end := args[1]

	req := common.ShortestPathRequest{Start: start, End: end}
	var result common.ShortestPathResponse
	if err := postJSON(serverURL+"/path/shortest", req, &result); err != nil {
		return err
	}

	if !result.Success {
		fmt.Printf("✗ %s\n", result.Message)
		return nil
	}

	fmt.Printf("✓ Shortest path found:\n")
	fmt.Printf("  Total time: %.2f minutes\n", result.TotalTime)
	fmt.Printf("  Edge count: %d\n", result.EdgeCount)
	fmt.Printf("  Path: %s\n", strings.Join(result.Path, " → "))
	return nil
}

func handleAllPaths(args []string, serverURL string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing start node")
	}

	start := args[0]

	req := common.AllShortestPathsRequest{Start: start}
	var result common.AllShortestPathsResponse
	if err := postJSON(serverURL+"/path/all", req, &result); err != nil {
		return err
	}

	if !result.Success {
		fmt.Printf("✗ %s\n", result.Message)
		return nil
	}

	fmt.Printf("✓ Shortest distances from %s:\n", start)
	for node, dist := range result.Distances {
		fmt.Printf("  %s → %.2f minutes\n", node, dist)
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL, remainingArgs := getServerURL(os.Args[1:])

	if len(remainingArgs) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := remainingArgs[0]
	args := remainingArgs[1:]

	var err error
	switch command {
	case "import":
		err = handleImport(args, serverURL)
	case "path":
		err = handlePath(args, serverURL)
	case "allpaths":
		err = handleAllPaths(args, serverURL)
	case "--help", "-h", "help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
