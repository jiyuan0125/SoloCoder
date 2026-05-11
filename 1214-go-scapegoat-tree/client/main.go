package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/scapegoat/common"
)

func getServerURL() string {
	var serverURL string
	flag.StringVar(&serverURL, "server", "", "server URL")
	flag.Parse()

	if serverURL == "" {
		serverURL = os.Getenv("SCAPEGOAT_SERVER")
	}

	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	return serverURL
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  insert <key> <value>  - Insert or update a key-value pair")
	fmt.Println("  search <key>          - Search for a key")
	fmt.Println("  remove <key>          - Remove a key")
	fmt.Println("  size                  - Get the number of elements")
	fmt.Println("\nOptions:")
	fmt.Println("  -server <url>         - Server URL (default: http://localhost:8080)")
	fmt.Println("  Environment variable: SCAPEGOAT_SERVER")
}

func main() {
	serverURL := getServerURL()
	
	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	switch cmd {
	case "insert":
		if len(args) != 3 {
			fmt.Println("Usage: insert <key> <value>")
			os.Exit(1)
		}
		handleInsert(serverURL, args[1], args[2])
	case "search":
		if len(args) != 2 {
			fmt.Println("Usage: search <key>")
			os.Exit(1)
		}
		handleSearch(serverURL, args[1])
	case "remove":
		if len(args) != 2 {
			fmt.Println("Usage: remove <key>")
			os.Exit(1)
		}
		handleDelete(serverURL, args[1])
	case "size":
		handleSize(serverURL)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func handleInsert(serverURL, keyStr, valueStr string) {
	key, err := strconv.ParseInt(keyStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid key: %s\n", keyStr)
		os.Exit(1)
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid value: %s\n", valueStr)
		os.Exit(1)
	}

	reqBody := common.InsertRequest{Key: key, Value: value}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/insert", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result map[string]bool
	json.NewDecoder(resp.Body).Decode(&result)

	if result["success"] {
		fmt.Printf("Inserted: %d -> %d\n", key, value)
	} else {
		fmt.Println("Insert failed")
		os.Exit(1)
	}
}

func handleSearch(serverURL, keyStr string) {
	key, err := strconv.ParseInt(keyStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid key: %s\n", keyStr)
		os.Exit(1)
	}

	query := url.Values{}
	query.Set("key", strconv.FormatInt(key, 10))

	resp, err := http.Get(serverURL + "/search?" + query.Encode())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result common.SearchResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Success {
		fmt.Printf("Found: %d -> %d\n", key, result.Value)
	} else {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}
}

func handleDelete(serverURL, keyStr string) {
	key, err := strconv.ParseInt(keyStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid key: %s\n", keyStr)
		os.Exit(1)
	}

	query := url.Values{}
	query.Set("key", strconv.FormatInt(key, 10))

	req, err := http.NewRequest(http.MethodDelete, serverURL+"/delete?"+query.Encode(), nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result common.DeleteResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Success {
		fmt.Printf("Removed: %d\n", key)
	} else {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}
}

func handleSize(serverURL string) {
	resp, err := http.Get(serverURL + "/size")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result common.SizeResponse
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Printf("Size: %d\n", result.Size)
}
