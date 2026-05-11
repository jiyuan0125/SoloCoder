package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"maxflow/pkg/api"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "server URL")
	inputFile := flag.String("input", "", "input JSON file path")
	source := flag.String("source", "", "source node ID")
	sink := flag.String("sink", "", "sink node ID")

	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintln(os.Stderr, "error: input file is required")
		flag.Usage()
		os.Exit(1)
	}

	data, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input file: %v\n", err)
		os.Exit(1)
	}

	var req api.FlowRequest
	if err := json.Unmarshal(data, &req); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	if *source != "" {
		req.Source = *source
	}

	if *sink != "" {
		req.Sink = *sink
	}

	if req.Source == "" {
		fmt.Fprintln(os.Stderr, "error: source is required (use --source or include in JSON)")
		os.Exit(1)
	}

	if req.Sink == "" {
		fmt.Fprintln(os.Stderr, "error: sink is required (use --sink or include in JSON)")
		os.Exit(1)
	}

	payload, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling request: %v\n", err)
		os.Exit(1)
	}

	url := *serverURL + "/maxflow"
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "server error (status %d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var result api.FlowResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Max Flow: %d\n\n", result.MaxFlow)
	fmt.Println("Flow Distribution:")
	for _, f := range result.Flows {
		fmt.Printf("  %s -> %s: %d\n", f.From, f.To, f.Flow)
	}
}
