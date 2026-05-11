package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"persistent-segtree/pkg/api"
)

func getServerURL() string {
	if url := os.Getenv("SERVER_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

func doRequest(method, url string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func buildCmd(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: build <v1> <v2> ...")
		os.Exit(1)
	}

	values := make([]int, len(args))
	for i, arg := range args {
		v, err := strconv.Atoi(arg)
		if err != nil {
			fmt.Printf("error: invalid value '%s'\n", arg)
			os.Exit(1)
		}
		values[i] = v
	}

	serverURL := getServerURL()
	respBytes, err := doRequest(http.MethodPost, serverURL+"/api/build", api.BuildRequest{Values: values})
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}

	var resp api.BuildResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("build success, version count: %d\n", resp.VersionCount)
	} else {
		fmt.Printf("build failed: %s\n", resp.Message)
		os.Exit(1)
	}
}

func queryCmd(args []string) {
	if len(args) != 3 {
		fmt.Println("usage: query <l> <r> <k>")
		os.Exit(1)
	}

	l, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("error: invalid l '%s'\n", args[0])
		os.Exit(1)
	}

	r, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("error: invalid r '%s'\n", args[1])
		os.Exit(1)
	}

	k, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Printf("error: invalid k '%s'\n", args[2])
		os.Exit(1)
	}

	serverURL := getServerURL()
	respBytes, err := doRequest(http.MethodPost, serverURL+"/api/query", api.QueryKthRequest{L: l, R: r, K: k})
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}

	var resp api.QueryKthResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("result: %d\n", resp.Value)
	} else {
		fmt.Printf("query failed: %s\n", resp.Message)
		os.Exit(1)
	}
}

func statusCmd() {
	serverURL := getServerURL()
	respBytes, err := doRequest(http.MethodGet, serverURL+"/api/status", nil)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}

	var resp api.StatusResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}

	if resp.Ready {
		fmt.Printf("ready, array length: %d\n", resp.ArrayLength)
	} else {
		fmt.Println("not ready")
	}
}

func printUsage() {
	fmt.Println("Persistent Segment Tree Client")
	fmt.Println()
	fmt.Println("usage:")
	fmt.Println("  client build <v1> <v2> ...  - build persistent segment tree")
	fmt.Println("  client query <l> <r> <k>     - query k-th smallest in [l, r]")
	fmt.Println("  client status                - check server status")
	fmt.Println()
	fmt.Println("environment:")
	fmt.Println("  SERVER_URL - server URL (default: http://localhost:8080)")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "build":
		buildCmd(os.Args[2:])
	case "query":
		queryCmd(os.Args[2:])
	case "status":
		statusCmd()
	default:
		printUsage()
		os.Exit(1)
	}
}
