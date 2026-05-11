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

	"github.com/example/lca-family-tree/common"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL (or use LCA_SERVER_URL environment variable)")
	inputFile := flag.String("input", "", "Path to input JSON file (or use '-' to read from stdin)")
	flag.Parse()

	if envURL := os.Getenv("LCA_SERVER_URL"); envURL != "" {
		serverURL = &envURL
	}

	var reqBytes []byte
	var err error

	if *inputFile == "-" || *inputFile == "" {
		reqBytes, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
	} else {
		reqBytes, err = ioutil.ReadFile(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
			os.Exit(1)
		}
	}

	if len(strings.TrimSpace(string(reqBytes))) == 0 {
		fmt.Fprintln(os.Stderr, "Error: empty input")
		os.Exit(1)
	}

	var req common.Request
	if err := json.Unmarshal(reqBytes, &req); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing input JSON: %v\n", err)
		os.Exit(1)
	}

	fullURL := strings.TrimRight(*serverURL, "/") + "/query"
	resp, err := http.Post(fullURL, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	var response common.Response
	if err := json.Unmarshal(respBytes, &response); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing server response: %v\n", err)
		fmt.Fprintf(os.Stderr, "Server response: %s\n", string(respBytes))
		os.Exit(1)
	}

	output, err := response.ToJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))

	hasError := response.Error != ""
	for _, result := range response.Results {
		if result.Error != "" {
			hasError = true
		}
	}

	if hasError {
		os.Exit(1)
	}
}
