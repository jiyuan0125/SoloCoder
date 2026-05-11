package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/user/tarjan-scc/pkg/api"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Tarjan SCC server URL")
	inputFile := flag.String("file", "", "Path to JSON input file with nodes and edges")
	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintln(os.Stderr, "Usage: client -file <input.json> [-server http://host:port]")
		flag.Usage()
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var req api.GraphRequest
	if err := json.Unmarshal(data, &req); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding request: %v\n", err)
		os.Exit(1)
	}

	url := strings.TrimRight(*serverURL, "/") + "/analyze"
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending request to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	var result api.GraphResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Println("Strongly Connected Components Analysis:")
	fmt.Println("=======================================")
	fmt.Println()

	hasCycles := false
	for i, scc := range result.SCCs {
		prefix := "  "
		marker := "OK"
		if scc.IsCycle {
			prefix = "⚠️ "
			marker = "CYCLE DETECTED"
			hasCycles = true
		}
		fmt.Printf("%sComponent %d: [%s] %s\n", prefix, i+1, marker, strings.Join(scc.Nodes, " <-> "))
	}

	fmt.Println()
	if hasCycles {
		fmt.Println("⚠️  Warning: Circular dependencies detected!")
		os.Exit(2)
	} else {
		fmt.Println("✓ No circular dependencies found.")
		os.Exit(0)
	}
}
