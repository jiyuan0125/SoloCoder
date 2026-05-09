package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/solocoder/priority-threadpool/common"
)

const (
	defaultServerURL = "http://localhost:8080"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := flag.String("server", defaultServerURL, "Server URL")
	flag.Parse()

	cmd := os.Args[1]

	switch cmd {
	case "submit":
		handleSubmit(*serverURL)
	case "status":
		handleStatus(*serverURL)
	case "shutdown":
		handleShutdown(*serverURL)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client submit <priority> <task>")
	fmt.Println("  client status")
	fmt.Println("  client shutdown")
	fmt.Println("\nPriority levels:")
	fmt.Println("  0 - Low")
	fmt.Println("  1 - Medium")
	fmt.Println("  2 - High")
	fmt.Println("\nOptions:")
	fmt.Println("  -server string")
	fmt.Println("        Server URL (default \"http://localhost:8080\")")
}

func handleSubmit(serverURL string) {
	if len(os.Args) < 4 {
		fmt.Println("Usage: client submit <priority> <task>")
		os.Exit(1)
	}

	priorityStr := os.Args[2]
	task := os.Args[3]

	priority, err := strconv.Atoi(priorityStr)
	if err != nil {
		fmt.Printf("Invalid priority: %s\n", priorityStr)
		os.Exit(1)
	}

	if priority < 0 || priority > 2 {
		fmt.Println("Priority must be 0 (Low), 1 (Medium), or 2 (High)")
		os.Exit(1)
	}

	req := common.SubmitRequest{
		Priority: common.Priority(priority),
		Task:     task,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/submit", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Failed to send request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		os.Exit(1)
	}

	var submitResp common.SubmitResponse
	if err := json.Unmarshal(respBody, &submitResp); err != nil {
		fmt.Printf("Failed to unmarshal response: %v\n", err)
		fmt.Printf("Response body: %s\n", string(respBody))
		os.Exit(1)
	}

	if submitResp.Success {
		fmt.Println(submitResp.Message)
	} else {
		fmt.Printf("Error: %s\n", submitResp.Error)
		os.Exit(1)
	}
}

func handleStatus(serverURL string) {
	resp, err := http.Get(serverURL + "/status")
	if err != nil {
		fmt.Printf("Failed to send request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		os.Exit(1)
	}

	var statusResp common.StatusResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		fmt.Printf("Failed to unmarshal response: %v\n", err)
		fmt.Printf("Response body: %s\n", string(respBody))
		os.Exit(1)
	}

	if statusResp.Success {
		fmt.Println("Thread Pool Status:")
		fmt.Printf("  Worker Count: %d\n", statusResp.WorkerCount)
		fmt.Printf("  Idle Worker Count: %d\n", statusResp.IdleWorkerCount)
		fmt.Printf("  Queue - High Priority: %d\n", statusResp.QueueHighPriority)
		fmt.Printf("  Queue - Medium Priority: %d\n", statusResp.QueueMediumPriority)
		fmt.Printf("  Queue - Low Priority: %d\n", statusResp.QueueLowPriority)
		fmt.Printf("  Starved Tasks Count: %d\n", statusResp.StarvedTasksCount)
	} else {
		fmt.Println("Failed to get status")
		os.Exit(1)
	}
}

func handleShutdown(serverURL string) {
	resp, err := http.Post(serverURL+"/shutdown", "application/json", nil)
	if err != nil {
		fmt.Printf("Failed to send request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		os.Exit(1)
	}

	var shutdownResp common.ShutdownResponse
	if err := json.Unmarshal(respBody, &shutdownResp); err != nil {
		fmt.Printf("Failed to unmarshal response: %v\n", err)
		fmt.Printf("Response body: %s\n", string(respBody))
		os.Exit(1)
	}

	if shutdownResp.Success {
		fmt.Println(shutdownResp.Message)
	} else {
		fmt.Printf("Error: %s\n", shutdownResp.Error)
		os.Exit(1)
	}
}
