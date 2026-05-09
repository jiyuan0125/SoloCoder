package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"plugin-system/pkg/common"
)

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Plugin server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	switch cmd {
	case "list":
		listPlugins()
	case "run":
		if len(args) < 3 {
			fmt.Println("Usage: run <plugin-name> <input-file>")
			os.Exit(1)
		}
		runPlugin(args[1], args[2])
	case "scan":
		scanPlugins()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: plugin-client [flags] <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  list                  List all loaded plugins")
	fmt.Println("  run <name> <file>     Execute a plugin with input from file")
	fmt.Println("  scan                  Trigger server to rescan plugin directory")
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

func listPlugins() {
	resp, err := http.Get(serverURL + "/api/plugins/list")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error (%d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var listResp common.ListResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		fmt.Printf("Raw response: %s\n", string(body))
		os.Exit(1)
	}

	if len(listResp.Plugins) == 0 {
		fmt.Println("No plugins loaded.")
		return
	}

	fmt.Println("Loaded plugins:")
	for _, p := range listResp.Plugins {
		fmt.Printf("  - %s (API: %s)\n", p.Name, p.APIVersion)
		if len(p.Capabilities) > 0 {
			fmt.Printf("    Capabilities: %v\n", p.Capabilities)
		}
	}
}

func runPlugin(name, inputFile string) {
	input, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read input file: %v\n", err)
		os.Exit(1)
	}

	req := common.ExecuteRequest{
		Name:  name,
		Input: input,
	}
	reqBody, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/api/plugins/execute", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error (%d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var execResp common.ExecuteResponse
	if err := json.Unmarshal(body, &execResp); err != nil {
		fmt.Printf("Result: %s\n", string(body))
		return
	}

	if execResp.Error != "" {
		fmt.Fprintf(os.Stderr, "Plugin error: %s\n", execResp.Error)
		os.Exit(1)
	}

	fmt.Printf("Plugin output:\n%s\n", string(execResp.Output))
}

func scanPlugins() {
	resp, err := http.Post(serverURL+"/api/plugins/scan", "application/json", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error (%d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var scanResp common.ScanResponse
	json.Unmarshal(body, &scanResp)
	fmt.Println(scanResp.Message)
}
