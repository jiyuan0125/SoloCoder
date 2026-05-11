package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/example/unionfind-rollback/common"
)

func main() {
	serverURLFlag := flag.String("server", "", "Server URL (e.g., http://localhost:8080)")
	inputFlag := flag.String("input", "", "JSON input (optional, if not provided, reads from stdin)")
	flag.Parse()

	serverURL := *serverURLFlag
	if serverURL == "" {
		serverURL = os.Getenv("UF_SERVER")
	}
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	var input string
	if *inputFlag != "" {
		input = *inputFlag
	} else {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
		input = string(data)
	}

	var req common.Request
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON input: %v\n", err)
		os.Exit(1)
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/actions", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending request to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	var response common.Response
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		fmt.Fprintf(os.Stderr, "Response body: %s\n", string(body))
		os.Exit(1)
	}

	for i, result := range response.Results {
		status := "SUCCESS"
		if !result.Success {
			status = "ERROR"
		}
		fmt.Printf("[%d] %s - %s (Step %d)\n", i, result.Type, status, result.Step)
		if result.Message != "" {
			fmt.Printf("  Message: %s\n", result.Message)
		}
		if result.Data != nil {
			jsonData, _ := json.MarshalIndent(result.Data, "  ", "  ")
			fmt.Printf("  Data: %s\n", string(jsonData))
		}
		fmt.Println()
	}

	if !response.Success {
		os.Exit(1)
	}
}
