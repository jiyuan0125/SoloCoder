package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"context-chain/pkg/api"
)

const defaultServerURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "request":
		handleRequest(args)
	case "cancel":
		handleCancel(args)
	case "chain":
		handleChain(args)
	case "precision":
		handlePrecision(args)
	case "leak":
		handleLeak(args)
	case "status":
		handleStatus(args)
	case "help":
		printUsage()
	default:
		fmt.Printf("unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Context Chain Client

Usage:
  client request [options]    Send a request to create a context
  client cancel <cancel_id>   Cancel a context by ID
  client chain                View the cancellation chain
  client precision [options]  Test timeout precision
  client leak <request_id>    Check for context leaks
  client status               Get server status

Options:
  For 'request':
    -timeout duration   Timeout for the context (e.g., 5s, 100ms)
    -parent string      Parent context ID (optional)
    -key string         Key for value (can be used multiple times with -val)
    -val string         Value for the key

  For 'precision':
    -timeout duration   Target timeout (default: 100ms)
    -iterations int     Number of test iterations (default: 10)
`)
}

func getServerURL() string {
	if url := os.Getenv("SERVER_URL"); url != "" {
		return url
	}
	return defaultServerURL
}

func httpRequest(method, path string, body interface{}) ([]byte, error) {
	url := getServerURL() + path

	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func handleRequest(args []string) {
	fs := flag.NewFlagSet("request", flag.ExitOnError)
	timeout := fs.Duration("timeout", 0, "timeout for the context")
	parentID := fs.String("parent", "", "parent context ID")
	key := fs.String("key", "", "key for value")
	val := fs.String("val", "", "value for the key")

	fs.Parse(args)

	values := []api.ValueItem{}
	if *key != "" {
		values = append(values, api.ValueItem{
			Key:   *key,
			Value: *val,
		})
	}

	req := api.RequestCreate{
		Timeout:  *timeout,
		Values:   values,
		ParentID: *parentID,
	}

	respBytes, err := httpRequest("POST", "/create", req)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	var pretty bytes.Buffer
	json.Indent(&pretty, respBytes, "", "  ")
	fmt.Println(string(pretty.Bytes()))
}

func handleCancel(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: client cancel <cancel_id>")
		os.Exit(1)
	}

	cancelID := args[0]

	req := api.RequestCancel{
		CancelID: cancelID,
	}

	respBytes, err := httpRequest("POST", "/cancel", req)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	var pretty bytes.Buffer
	json.Indent(&pretty, respBytes, "", "  ")
	fmt.Println(string(pretty.Bytes()))
}

func handleChain(args []string) {
	respBytes, err := httpRequest("GET", "/chain", nil)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	var pretty bytes.Buffer
	json.Indent(&pretty, respBytes, "", "  ")
	fmt.Println(string(pretty.Bytes()))
}

func handlePrecision(args []string) {
	fs := flag.NewFlagSet("precision", flag.ExitOnError)
	timeout := fs.Duration("timeout", 100*time.Millisecond, "target timeout")
	iterations := fs.Int("iterations", 10, "number of test iterations")

	fs.Parse(args)

	req := api.RequestPrecision{
		TargetTimeout: *timeout,
		Iterations:    *iterations,
	}

	respBytes, err := httpRequest("POST", "/precision", req)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	var pretty bytes.Buffer
	json.Indent(&pretty, respBytes, "", "  ")
	fmt.Println(string(pretty.Bytes()))
}

func handleLeak(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: client leak <request_id>")
		os.Exit(1)
	}

	requestID := args[0]

	req := api.RequestLeak{
		RequestID: requestID,
	}

	respBytes, err := httpRequest("POST", "/leak", req)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	var pretty bytes.Buffer
	json.Indent(&pretty, respBytes, "", "  ")
	fmt.Println(string(pretty.Bytes()))
}

func handleStatus(args []string) {
	respBytes, err := httpRequest("GET", "/status", nil)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	var pretty bytes.Buffer
	json.Indent(&pretty, respBytes, "", "  ")
	fmt.Println(string(pretty.Bytes()))
}
