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

	"leftistheap/api"
)

var serverURL = "http://localhost:8080"
var currentHeapID string

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	if envURL := os.Getenv("HEAP_SERVER_URL"); envURL != "" {
		serverURL = envURL
	}

	command := os.Args[1]
	args := os.Args[2:]

	if len(args) > 0 && strings.HasPrefix(args[len(args)-1], "--heap=") {
		heapArg := args[len(args)-1]
		currentHeapID = strings.TrimPrefix(heapArg, "--heap=")
		args = args[:len(args)-1]
	}

	switch command {
	case "create":
		handleCreate(args)
	case "insert":
		handleInsert(args)
	case "top":
		handleTop()
	case "pop":
		handlePop()
	case "merge":
		handleMerge(args)
	case "set-heap":
		handleSetHeap(args)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func handleCreate(args []string) {
	heapType := "min"
	if len(args) > 0 {
		if args[0] == "min" || args[0] == "max" {
			heapType = args[0]
		}
	}

	req := api.CreateHeapRequest{
		HeapType: heapType,
	}
	reqBody, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/create", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var createResp api.CreateHeapResponse
	json.Unmarshal(body, &createResp)

	if createResp.Error != "" {
		fmt.Printf("Error: %s\n", createResp.Error)
		return
	}

	currentHeapID = createResp.HeapID
	fmt.Printf("Created heap: %s\n", currentHeapID)
}

func handleInsert(args []string) {
	if currentHeapID == "" {
		fmt.Println("Error: No heap selected. Use 'create' or 'set-heap' first.")
		return
	}

	if len(args) < 1 {
		fmt.Println("Usage: insert <value>")
		return
	}

	value, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Printf("Invalid value: %v\n", err)
		return
	}

	req := api.InsertRequest{
		HeapID: currentHeapID,
		Value:  value,
	}
	reqBody, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/insert", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var insertResp api.InsertResponse
	json.Unmarshal(body, &insertResp)

	if insertResp.Error != "" {
		fmt.Printf("Error: %s\n", insertResp.Error)
		return
	}

	fmt.Printf("Inserted value: %d\n", value)
}

func handleTop() {
	if currentHeapID == "" {
		fmt.Println("Error: No heap selected. Use 'create' or 'set-heap' first.")
		return
	}

	req := api.TopRequest{
		HeapID: currentHeapID,
	}
	reqBody, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/top", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var topResp api.TopResponse
	json.Unmarshal(body, &topResp)

	if topResp.Error != "" {
		fmt.Printf("Error: %s\n", topResp.Error)
		return
	}

	if topResp.Value != nil {
		fmt.Printf("Top value: %d\n", *topResp.Value)
	}
}

func handlePop() {
	if currentHeapID == "" {
		fmt.Println("Error: No heap selected. Use 'create' or 'set-heap' first.")
		return
	}

	req := api.PopRequest{
		HeapID: currentHeapID,
	}
	reqBody, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/pop", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var popResp api.PopResponse
	json.Unmarshal(body, &popResp)

	if popResp.Error != "" {
		fmt.Printf("Error: %s\n", popResp.Error)
		return
	}

	if popResp.Value != nil {
		fmt.Printf("Popped value: %d\n", *popResp.Value)
	}
}

func handleMerge(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: merge <heap_id>")
		return
	}

	heapID2 := args[0]

	if currentHeapID == "" {
		fmt.Println("Error: No heap selected. Use 'create' or 'set-heap' first.")
		return
	}

	req := api.MergeRequest{
		HeapID1: currentHeapID,
		HeapID2: heapID2,
	}
	reqBody, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/merge", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var mergeResp api.MergeResponse
	json.Unmarshal(body, &mergeResp)

	if mergeResp.Error != "" {
		fmt.Printf("Error: %s\n", mergeResp.Error)
		return
	}

	fmt.Printf("Merged heaps. New heap ID: %s\n", mergeResp.NewHeapID)
}

func handleSetHeap(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: set-heap <heap_id>")
		return
	}

	currentHeapID = args[0]
	fmt.Printf("Current heap set to: %s\n", currentHeapID)
}

func printUsage() {
	fmt.Println("Leftist Heap Client")
	fmt.Println("Usage:")
	fmt.Println("  client create [min|max]  - Create a new heap")
	fmt.Println("  client insert <value>    - Insert a value into the current heap")
	fmt.Println("  client top               - Get the top value of the current heap")
	fmt.Println("  client pop               - Pop the top value from the current heap")
	fmt.Println("  client merge <heap_id>   - Merge another heap into the current heap")
	fmt.Println("  client set-heap <id>     - Set the current heap ID")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --heap=<heap_id>         - Specify heap ID for a single command")
	fmt.Println("")
	fmt.Println("Environment variables:")
	fmt.Println("  HEAP_SERVER_URL          - Server URL (default: http://localhost:8080)")
}
