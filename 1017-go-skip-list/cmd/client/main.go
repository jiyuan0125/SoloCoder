package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"skip-list/common"
)

const baseURL = "http://localhost:8080/skiplist"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "put":
		putCommand(os.Args[2:])
	case "get":
		getCommand(os.Args[2:])
	case "delete":
		deleteCommand(os.Args[2:])
	case "range":
		rangeCommand(os.Args[2:])
	case "rank":
		rankCommand(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  client put <key> <value>")
	fmt.Println("  client get <key>")
	fmt.Println("  client delete <key>")
	fmt.Println("  client range <from> <to>")
	fmt.Println("  client rank <key>")
}

func putCommand(args []string) {
	if len(args) != 2 {
		fmt.Println("Usage: client put <key> <value>")
		os.Exit(1)
	}

	key := args[0]
	value := args[1]

	reqBody := common.PutRequest{
		Value: value,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	req, err := http.NewRequest("PUT", baseURL+"/"+key, bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("{\"status\":\"ok\",\"key\":\"%s\"}\n", key)
}

func getCommand(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: client get <key>")
		os.Exit(1)
	}

	key := args[0]

	resp, err := http.Get(baseURL + "/" + key)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result common.GetResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Found {
		fmt.Printf("Key: %s\n", result.Key)
		fmt.Printf("Value: %s\n", result.Value)
	} else {
		fmt.Println("Key not found")
	}
}

func deleteCommand(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: client delete <key>")
		os.Exit(1)
	}

	key := args[0]

	req, err := http.NewRequest("DELETE", baseURL+"/"+key, nil)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result common.DeleteResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Deleted {
		fmt.Println("Deleted")
	} else {
		fmt.Println("Key not found")
	}
}

func rangeCommand(args []string) {
	if len(args) != 2 {
		fmt.Println("Usage: client range <from> <to>")
		os.Exit(1)
	}

	from := args[0]
	to := args[1]

	resp, err := http.Get(fmt.Sprintf("%s/range?from=%s&to=%s", baseURL, from, to))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result common.RangeResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Data) == 0 {
		fmt.Println("No results")
		return
	}

	fmt.Println("Range results:")
	for _, kv := range result.Data {
		fmt.Printf("  %s -> %s\n", kv.Key, kv.Value)
	}
}

func rankCommand(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: client rank <key>")
		os.Exit(1)
	}

	key := args[0]

	resp, err := http.Get(fmt.Sprintf("%s/rank/%s", baseURL, key))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result common.RankResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Found {
		fmt.Printf("Rank: %d\n", result.Rank)
	} else {
		fmt.Println("Key not found")
	}
}
