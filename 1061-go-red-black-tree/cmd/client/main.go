package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/solocoder/rbtree/pkg/api"
)

const baseURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "insert":
		if len(os.Args) != 3 {
			fmt.Println("Usage: rbtree-client insert <key>")
			os.Exit(1)
		}
		key, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid key:", os.Args[2])
			os.Exit(1)
		}
		handleInsert(key)

	case "delete":
		if len(os.Args) != 3 {
			fmt.Println("Usage: rbtree-client delete <key>")
			os.Exit(1)
		}
		key, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid key:", os.Args[2])
			os.Exit(1)
		}
		handleDelete(key)

	case "range":
		if len(os.Args) != 4 {
			fmt.Println("Usage: rbtree-client range <low> <high>")
			os.Exit(1)
		}
		low, err1 := strconv.Atoi(os.Args[2])
		high, err2 := strconv.Atoi(os.Args[3])
		if err1 != nil || err2 != nil {
			fmt.Println("Invalid bounds")
			os.Exit(1)
		}
		handleRange(low, high)

	case "kthlargest":
		if len(os.Args) != 3 {
			fmt.Println("Usage: rbtree-client kthlargest <k>")
			os.Exit(1)
		}
		k, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid k:", os.Args[2])
			os.Exit(1)
		}
		handleKthLargest(k)

	case "size":
		handleSize()

	default:
		fmt.Println("Unknown command:", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: rbtree-client <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  insert <key>      Insert a key into the tree")
	fmt.Println("  delete <key>      Delete a key from the tree")
	fmt.Println("  range <low> <high> Query keys in range [low, high]")
	fmt.Println("  kthlargest <k>    Find the k-th largest element")
	fmt.Println("  size              Get the size of the tree")
}

func handleInsert(key int) {
	req := api.InsertRequest{Key: key}
	body, _ := json.Marshal(req)

	resp, err := http.Post(baseURL+"/insert", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	var result api.InsertResponse
	json.Unmarshal(data, &result)

	if result.Success {
		fmt.Printf("Inserted key: %d\n", key)
	} else {
		fmt.Printf("Failed to insert: %s\n", result.Message)
	}
}

func handleDelete(key int) {
	req := api.DeleteRequest{Key: key}
	body, _ := json.Marshal(req)

	resp, err := http.Post(baseURL+"/delete", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	var result api.DeleteResponse
	json.Unmarshal(data, &result)

	if result.Success {
		fmt.Printf("Deleted key: %d\n", key)
	} else {
		fmt.Printf("Failed to delete: %s\n", result.Message)
	}
}

func handleRange(low, high int) {
	req := api.RangeQueryRequest{Low: low, High: high}
	body, _ := json.Marshal(req)

	resp, err := http.Post(baseURL+"/range", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	var result api.RangeQueryResponse
	json.Unmarshal(data, &result)

	if result.Success {
		fmt.Printf("Range query [%d, %d]: %v\n", low, high, result.Values)
	} else {
		fmt.Printf("Range query failed: %s\n", result.Message)
	}
}

func handleKthLargest(k int) {
	req := api.KthLargestRequest{K: k}
	body, _ := json.Marshal(req)

	resp, err := http.Post(baseURL+"/kthlargest", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	var result api.KthLargestResponse
	json.Unmarshal(data, &result)

	if result.Success {
		fmt.Printf("The %d-th largest element is: %d\n", k, result.Value)
	} else {
		fmt.Printf("Kth largest query failed: %s\n", result.Message)
	}
}

func handleSize() {
	resp, err := http.Post(baseURL+"/size", "application/json", bytes.NewBuffer([]byte{}))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	var result api.SizeResponse
	json.Unmarshal(data, &result)

	if result.Success {
		fmt.Printf("Tree size: %d\n", result.Size)
	} else {
		fmt.Printf("Size query failed: %s\n", result.Message)
	}
}
