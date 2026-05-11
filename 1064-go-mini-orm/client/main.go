package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"mini-orm/common"
)

var serverURL = "http://localhost:8204"

func sendRequest(endpoint string, reqBody interface{}, respBody interface{}) error {
	var body io.Reader
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	resp, err := http.Post(serverURL+endpoint, "application/json", body)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}

func parseDataFlag(dataStr string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	if dataStr == "" {
		return result, nil
	}

	parts := strings.Split(dataStr, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid data format: %s", part)
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		if intVal, err := strconv.Atoi(value); err == nil {
			result[key] = intVal
		} else if boolVal, err := strconv.ParseBool(value); err == nil {
			result[key] = boolVal
		} else if value == "null" || value == "NULL" {
			result[key] = nil
		} else {
			result[key] = value
		}
	}
	return result, nil
}

func usage() {
	fmt.Println("Usage: orm-client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  insert   Insert a new record")
	fmt.Println("  get      Get a record by ID")
	fmt.Println("  update   Update a record by ID")
	fmt.Println("  delete   Delete a record by ID")
	fmt.Println("  list     List records with optional filters")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -model    Model name (default: user)")
	fmt.Println("  -id       Record ID (for get, update, delete)")
	fmt.Println("  -data     Key=value pairs, comma-separated (for insert, update)")
	fmt.Println("  -where    Where conditions, key=value pairs (for list)")
	fmt.Println("  -order    Order by column (for list)")
	fmt.Println("  -limit    Limit number of results (for list)")
	fmt.Println("  -offset   Offset for pagination (for list)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  orm-client insert -model user -data name=John,email=john@example.com,age=30,active=true")
	fmt.Println("  orm-client get -model user -id 1")
	fmt.Println("  orm-client update -model user -id 1 -data name=Johnny")
	fmt.Println("  orm-client delete -model user -id 1")
	fmt.Println("  orm-client list -model user -where active=true -order id -limit 10")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]

	flagSet := flag.NewFlagSet("orm-client", flag.ExitOnError)
	modelFlag := flagSet.String("model", "user", "Model name")
	idFlag := flagSet.String("id", "", "Record ID")
	dataFlag := flagSet.String("data", "", "Data (key=value pairs)")
	whereFlag := flagSet.String("where", "", "Where conditions")
	orderFlag := flagSet.String("order", "", "Order by")
	limitFlag := flagSet.Int("limit", 0, "Limit")
	offsetFlag := flagSet.Int("offset", 0, "Offset")

	if err := flagSet.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var err error
	switch command {
	case "insert":
		data, err := parseDataFlag(*dataFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		req := common.InsertRequest{
			Model: *modelFlag,
			Data:  data,
		}

		var resp common.InsertResponse
		if err := sendRequest("/insert", &req, &resp); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printResponse(resp)

	case "get":
		if *idFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: -id is required for get command\n")
			os.Exit(1)
		}

		var id interface{}
		if intID, err := strconv.ParseInt(*idFlag, 10, 64); err == nil {
			id = intID
		} else {
			id = *idFlag
		}

		req := common.GetByIDRequest{
			Model: *modelFlag,
			ID:    id,
		}

		var resp common.GetByIDResponse
		if err := sendRequest("/get", &req, &resp); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printResponse(resp)

	case "update":
		if *idFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: -id is required for update command\n")
			os.Exit(1)
		}

		data, err := parseDataFlag(*dataFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var id interface{}
		if intID, err := strconv.ParseInt(*idFlag, 10, 64); err == nil {
			id = intID
		} else {
			id = *idFlag
		}

		req := common.UpdateRequest{
			Model: *modelFlag,
			ID:    id,
			Data:  data,
		}

		var resp common.UpdateResponse
		if err := sendRequest("/update", &req, &resp); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printResponse(resp)

	case "delete":
		if *idFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: -id is required for delete command\n")
			os.Exit(1)
		}

		var id interface{}
		if intID, err := strconv.ParseInt(*idFlag, 10, 64); err == nil {
			id = intID
		} else {
			id = *idFlag
		}

		req := common.DeleteRequest{
			Model: *modelFlag,
			ID:    id,
		}

		var resp common.DeleteResponse
		if err := sendRequest("/delete", &req, &resp); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printResponse(resp)

	case "list":
		where, err := parseDataFlag(*whereFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		req := common.ListRequest{
			Model:   *modelFlag,
			Where:   where,
			OrderBy: *orderFlag,
			Limit:   *limitFlag,
			Offset:  *offsetFlag,
		}

		var resp common.ListResponse
		if err := sendRequest("/list", &req, &resp); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printResponse(resp)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printResponse(resp interface{}) {
	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}
