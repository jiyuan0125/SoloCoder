package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/solocoder/fenwickbit/api"
)

var serverURL = "http://localhost:8080"

func doRequest(method, path string, body interface{}, response interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, serverURL+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return fmt.Errorf(errResp.Error)
		}
		return fmt.Errorf("server error: %s", resp.Status)
	}

	if response != nil {
		return json.NewDecoder(resp.Body).Decode(response)
	}
	return nil
}

func cmdCreate(args []string) string {
	if len(args) != 1 {
		fmt.Println("Usage: create <size>")
		return ""
	}

	size, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Error: invalid size")
		return ""
	}

	var resp api.CreateResponse
	if err := doRequest(http.MethodPost, "/create", api.CreateRequest{Size: size}, &resp); err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	fmt.Println("Created tree with ID:", resp.ID)
	return resp.ID
}

func cmdUpdate(args []string, treeID string) {
	if len(args) != 2 {
		fmt.Println("Usage: update <index> <delta>")
		return
	}

	index, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Error: invalid index")
		return
	}

	delta, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		fmt.Println("Error: invalid delta")
		return
	}

	var resp api.UpdateResponse
	if err := doRequest(http.MethodPost, "/update", api.UpdateRequest{ID: treeID, Index: index, Delta: delta}, &resp); err != nil {
		fmt.Println("Error:", err)
		return
	}

	if resp.Success {
		fmt.Println("Update successful")
	} else {
		fmt.Println("Update failed:", resp.Error)
	}
}

func cmdQueryPrefix(args []string, treeID string) {
	if len(args) != 1 {
		fmt.Println("Usage: query-prefix <index>")
		return
	}

	index, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Error: invalid index")
		return
	}

	var resp api.QueryPrefixResponse
	if err := doRequest(http.MethodPost, "/query-prefix", api.QueryPrefixRequest{ID: treeID, Index: index}, &resp); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Prefix sum:", resp.Result)
}

func cmdQueryRange(args []string, treeID string) {
	if len(args) != 2 {
		fmt.Println("Usage: query <l> <r>")
		return
	}

	l, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Error: invalid l")
		return
	}

	r, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Error: invalid r")
		return
	}

	var resp api.QueryRangeResponse
	if err := doRequest(http.MethodPost, "/query-range", api.QueryRangeRequest{ID: treeID, L: l, R: r}, &resp); err != nil {
		fmt.Println("Error:", err)
		return
	}

	if resp.Error != "" {
		fmt.Println("Query failed:", resp.Error)
	} else {
		fmt.Println("Range sum:", resp.Result)
	}
}

func cmdInversions(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: inversions <v1> <v2> ...")
		return
	}

	arr := make([]int64, len(args))
	for i, v := range args {
		num, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			fmt.Printf("Error: invalid number %q\n", v)
			return
		}
		arr[i] = num
	}

	var resp api.InversionsResponse
	if err := doRequest(http.MethodPost, "/inversions", api.InversionsRequest{Array: arr}, &resp); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Inversion count:", resp.Count)
}

func printHelp() {
	fmt.Println("Fenwick Tree Client")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create <size>           - Create a new Fenwick Tree")
	fmt.Println("  update <index> <delta>  - Update a position (requires active tree)")
	fmt.Println("  query-prefix <index>    - Query prefix sum (requires active tree)")
	fmt.Println("  query <l> <r>           - Query range sum [l, r] (requires active tree)")
	fmt.Println("  inversions <v1> <v2>... - Count inversions in array")
	fmt.Println("  help                    - Show this help")
	fmt.Println("  exit                    - Exit")
}

func main() {
	serverAddr := flag.String("server", "http://localhost:8080", "server URL")
	flag.Parse()

	if envServer := os.Getenv("SERVER_URL"); envServer != "" {
		*serverAddr = envServer
	}
	serverURL = *serverAddr

	var treeID string

	fmt.Println("Fenwick Tree Client. Type 'help' for commands.")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToLower(parts[0])
		args := parts[1:]

		switch cmd {
		case "help", "?":
			printHelp()
		case "exit", "quit":
			return
		case "create":
			treeID = cmdCreate(args)
		case "update":
			cmdUpdate(args, treeID)
		case "query-prefix":
			cmdQueryPrefix(args, treeID)
		case "query":
			cmdQueryRange(args, treeID)
		case "inversions":
			cmdInversions(args)
		default:
			fmt.Println("Unknown command. Type 'help' for available commands.")
		}
	}
}
