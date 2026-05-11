package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"generic-checker/pkg/api"
)

const defaultServerURL = "http://localhost:8430"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	restArgs := os.Args[2:]

	switch cmd {
	case "check":
		runCheck(restArgs)
	case "health":
		runHealth(restArgs)
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage:
  client check [flags] <request.json>
  client health [flags]

Flags:`)
	flag.PrintDefaults()
}

func getServerURL() string {
	if envURL := os.Getenv("SERVER_URL"); envURL != "" {
		return envURL
	}
	return defaultServerURL
}

func runHealth(args []string) {
	fs := flag.NewFlagSet("health", flag.ExitOnError)
	serverURL := fs.String("server", getServerURL(), "server URL")
	fs.Parse(args)

	url := strings.TrimRight(*serverURL, "/") + "/health"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("health check failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func runCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	serverURL := fs.String("server", getServerURL(), "server URL")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Println("error: request file required")
		os.Exit(1)
	}

	reqFile := fs.Arg(0)
	reqData, err := os.ReadFile(reqFile)
	if err != nil {
		log.Fatalf("failed to read request file: %v", err)
	}

	var req api.CheckRequest
	if err := json.Unmarshal(reqData, &req); err != nil {
		log.Fatalf("failed to parse request: %v", err)
	}

	url := strings.TrimRight(*serverURL, "/") + "/check"
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqData))
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))
}
