package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"trie-router/common"
)

const serverURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "register":
		handleRegister(args)
	case "query":
		handleQuery(args)
	case "list":
		handleList(args)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client register <path> <method> <handler_name>")
	fmt.Println("  client query <path> <method>")
	fmt.Println("  client list [sort_by] (sort_by: registration|priority|path)")
}

func handleRegister(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: client register <path> <method> <handler_name>")
		os.Exit(1)
	}

	req := common.RegisterRequest{
		Path:        args[0],
		Method:      args[1],
		HandlerName: args[2],
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/routes/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result common.RegisterResponse
	json.Unmarshal(data, &result)

	if result.Success {
		fmt.Println("Route registered successfully!")
	} else {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}
}

func handleQuery(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: client query <path> <method>")
		os.Exit(1)
	}

	req := common.QueryRequest{
		Path:   args[0],
		Method: args[1],
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/routes/query", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result common.QueryResponse
	json.Unmarshal(data, &result)

	fmt.Printf("Status: %d\n", result.Status)
	if result.Found {
		fmt.Printf("Found: Yes\n")
		if result.Route != nil {
			fmt.Printf("Route: %s %s -> %s\n", result.Route.Method, result.Route.Path, result.Route.HandlerName)
		}
		if len(result.Params) > 0 {
			fmt.Println("Params:")
			for k, v := range result.Params {
				fmt.Printf("  %s = %s\n", k, v)
			}
		}
	} else {
		fmt.Printf("Found: No\n")
		if len(result.AllowedMethods) > 0 {
			fmt.Printf("Allowed methods: %s\n", strings.Join(result.AllowedMethods, ", "))
		}
		if result.Error != "" {
			fmt.Printf("Error: %s\n", result.Error)
		}
	}
}

func handleList(args []string) {
	sortBy := ""
	if len(args) > 0 {
		sortBy = args[0]
	}

	var body io.Reader
	if sortBy != "" {
		req := common.ListRequest{SortBy: sortBy}
		b, _ := json.Marshal(req)
		body = bytes.NewBuffer(b)
	}

	var resp *http.Response
	var err error
	if body != nil {
		resp, err = http.Post(serverURL+"/api/routes/list", "application/json", body)
	} else {
		resp, err = http.Get(serverURL + "/api/routes/list")
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result common.ListResponse
	json.Unmarshal(data, &result)

	fmt.Printf("Total routes: %d\n", result.Count)
	fmt.Println("Routes:")
	for i, route := range result.Routes {
		fmt.Printf("  %d. %s %s -> %s\n", i+1, route.Method, route.Path, route.HandlerName)
	}
}
