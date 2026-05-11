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

	"cuckoo/pkg/common"
)

func getServerURL() string {
	if url := os.Getenv("SERVER_URL"); url != "" {
		return url
	}
	return "http://localhost:8214"
}

func makeJSONRequest(method, url string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return http.DefaultClient.Do(req)
}

func printUsage() {
	fmt.Println("Cuckoo Filter Client")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  client create <name> <capacity>")
	fmt.Println("  client insert <name> <element1> [element2 ...]")
	fmt.Println("  client lookup <name> <element>")
	fmt.Println("  client delete <name> <element>")
	fmt.Println("  client info <name>")
	fmt.Println("")
	fmt.Println("Environment Variables:")
	fmt.Println("  SERVER_URL - Server URL (default: http://localhost:8214)")
}

func cmdCreate(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: client create <name> <capacity>")
	}

	name := args[0]
	capacity, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid capacity: %v", err)
	}

	server := getServerURL()
	resp, err := makeJSONRequest(http.MethodPost, server+"/create", common.CreateRequest{
		Name:     name,
		Capacity: capacity,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result common.CreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if result.Success {
		fmt.Printf("Filter '%s' created successfully with capacity %d\n", name, capacity)
	} else {
		fmt.Printf("Failed to create filter: %s\n", result.Message)
	}
	return nil
}

func cmdInsert(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: client insert <name> <element1> [element2 ...]")
	}

	name := args[0]
	elements := args[1:]

	server := getServerURL()
	resp, err := makeJSONRequest(http.MethodPost, server+"/insert", common.InsertRequest{
		Name:     name,
		Elements: elements,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result common.InsertResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if result.Success {
		fmt.Printf("Inserted %d elements, failed %d elements\n", result.Inserted, result.Failed)
	} else {
		fmt.Printf("Failed to insert: %s\n", result.Message)
	}
	return nil
}

func cmdLookup(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: client lookup <name> <element>")
	}

	name := args[0]
	element := args[1]

	server := getServerURL()
	resp, err := makeJSONRequest(http.MethodPost, server+"/lookup", common.LookupRequest{
		Name:    name,
		Element: element,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result common.LookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if result.Exists {
		fmt.Printf("Element '%s' might exist in filter '%s'\n", element, name)
	} else {
		fmt.Printf("Element '%s' does NOT exist in filter '%s'\n", element, name)
	}
	return nil
}

func cmdDelete(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: client delete <name> <element>")
	}

	name := args[0]
	element := args[1]

	server := getServerURL()
	resp, err := makeJSONRequest(http.MethodPost, server+"/delete", common.DeleteRequest{
		Name:    name,
		Element: element,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result common.DeleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if result.Success {
		fmt.Printf("Element '%s' deleted from filter '%s'\n", element, name)
	} else {
		fmt.Printf("Failed to delete: %s\n", result.Message)
	}
	return nil
}

func cmdInfo(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: client info <name>")
	}

	name := args[0]

	server := getServerURL()
	resp, err := makeJSONRequest(http.MethodPost, server+"/info", common.InfoRequest{
		Name: name,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result common.InfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if !result.Exists {
		fmt.Printf("Filter '%s' does not exist\n", name)
		return nil
	}

	fmt.Printf("Filter: %s\n", result.Name)
	fmt.Printf("  Used slots: %d\n", result.UsedSlots)
	fmt.Printf("  Total slots: %d\n", result.TotalSlots)
	fmt.Printf("  Load rate: %.2f%%\n", result.LoadRate*100)
	return nil
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	var err error

	switch cmd {
	case "create":
		err = cmdCreate(args[1:])
	case "insert":
		err = cmdInsert(args[1:])
	case "lookup":
		err = cmdLookup(args[1:])
	case "delete":
		err = cmdDelete(args[1:])
	case "info":
		err = cmdInfo(args[1:])
	case "-h", "--help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
