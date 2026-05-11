package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	
	"hashtable/pkg/api"
)

const defaultServerURL = "http://localhost:8201"

func getServerURL() string {
	url := os.Getenv("SERVER_URL")
	if url != "" {
		return url
	}
	return defaultServerURL
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client create <method> <capacity> - Create a hash table")
	fmt.Println("  client put <key> <value>          - Put a key-value pair")
	fmt.Println("  client get <key>                  - Get value by key")
	fmt.Println("  client remove <key>               - Remove a key")
	fmt.Println("  client stats                      - Get hash table stats")
	fmt.Println("")
	fmt.Println("Methods: linear, quadratic, double")
}

func doPost(url string, reqBody interface{}, respBody interface{}) error {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	return json.Unmarshal(body, respBody)
}

func doGet(url string, respBody interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	return json.Unmarshal(body, respBody)
}

func handleCreate(args []string) {
	if len(args) < 2 {
		printUsage()
		return
	}
	
	method := args[0]
	capacity, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Error: capacity must be an integer")
		return
	}
	
	var resp api.CreateResponse
	url := getServerURL() + "/create"
	req := api.CreateRequest{Method: method, Capacity: capacity}
	
	if err := doPost(url, req, &resp); err != nil {
		fmt.Println("Error:", err)
		return
	}
	
	if resp.Success {
		fmt.Println(resp.Message)
	} else {
		fmt.Println("Error:", resp.Message)
	}
}

func handlePut(args []string) {
	if len(args) < 2 {
		printUsage()
		return
	}
	
	key := args[0]
	value := args[1]
	
	var resp api.PutResponse
	url := getServerURL() + "/put"
	req := api.PutRequest{Key: key, Value: value}
	
	if err := doPost(url, req, &resp); err != nil {
		fmt.Println("Error:", err)
		return
	}
	
	if resp.Success {
		fmt.Println("OK")
	} else {
		fmt.Println("Error:", resp.Message)
	}
}

func handleGet(args []string) {
	if len(args) < 1 {
		printUsage()
		return
	}
	
	key := args[0]
	
	var resp api.GetResponse
	url := getServerURL() + "/get"
	req := api.GetRequest{Key: key}
	
	if err := doPost(url, req, &resp); err != nil {
		fmt.Println("Error:", err)
		return
	}
	
	if resp.Success {
		fmt.Println(resp.Value)
	} else {
		fmt.Println("Error:", resp.Message)
	}
}

func handleRemove(args []string) {
	if len(args) < 1 {
		printUsage()
		return
	}
	
	key := args[0]
	
	var resp api.RemoveResponse
	url := getServerURL() + "/remove"
	req := api.RemoveRequest{Key: key}
	
	if err := doPost(url, req, &resp); err != nil {
		fmt.Println("Error:", err)
		return
	}
	
	if resp.Success {
		fmt.Println("OK")
	} else {
		fmt.Println("Error:", resp.Message)
	}
}

func handleStats() {
	var resp api.StatsResponse
	url := getServerURL() + "/stats"
	
	if err := doGet(url, &resp); err != nil {
		fmt.Println("Error:", err)
		return
	}
	
	if resp.Success {
		fmt.Printf("Size:        %d\n", resp.Size)
		fmt.Printf("Tombstones:  %d\n", resp.Tombstones)
		fmt.Printf("Capacity:    %d\n", resp.Capacity)
		fmt.Printf("Max Probe:   %d\n", resp.MaxProbeLen)
	} else {
		fmt.Println("Error:", resp.Message)
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	
	command := os.Args[1]
	args := os.Args[2:]
	
	switch command {
	case "create":
		handleCreate(args)
	case "put":
		handlePut(args)
	case "get":
		handleGet(args)
	case "remove":
		handleRemove(args)
	case "stats":
		handleStats()
	default:
		printUsage()
		os.Exit(1)
	}
}
