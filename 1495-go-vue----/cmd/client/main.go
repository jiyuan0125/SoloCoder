package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"pest-control/api"
)

var baseURL string

func main() {
	flag.StringVar(&baseURL, "server", "http://localhost:8080", "Server base URL")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "customers":
		handleCustomers(rest)
	case "contracts":
		handleContracts(rest)
	case "staff":
		handleStaff(rest)
	case "chemicals":
		handleChemicals(rest)
	case "inspections":
		handleInspections(rest)
	case "operations":
		handleOperations(rest)
	case "pending":
		handlePending()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Pest Control CLI - Usage:")
	fmt.Println("  pestctl customers list                    List all customers")
	fmt.Println("  pestctl contracts list                 List all contracts")
	fmt.Println("  pestctl staff list                     List active staff")
	fmt.Println("  pestctl chemicals list                 List all chemicals")
	fmt.Println("  pestctl inspections list <contract-id> List inspections for a contract")
	fmt.Println("  pestctl operations list <contract-id>  List operations for a contract")
	fmt.Println("  pestctl pending                        List pending services")
}

func handleCustomers(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: pestctl customers list")
		return
	}
	if args[0] == "list" {
		resp, err := http.Get(baseURL + "/customers")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var result api.Response
		json.Unmarshal(body, &result)
		printJSON(result)
	}
}

func handleContracts(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: pestctl contracts list")
		return
	}
	if args[0] == "list" {
		resp, err := http.Get(baseURL + "/contracts")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var result api.Response
		json.Unmarshal(body, &result)
		printJSON(result)
	}
}

func handleStaff(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: pestctl staff list")
		return
	}
	if args[0] == "list" {
		resp, err := http.Get(baseURL + "/staff")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var result api.Response
		json.Unmarshal(body, &result)
		printJSON(result)
	}
}

func handleChemicals(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: pestctl chemicals list")
		return
	}
	if args[0] == "list" {
		resp, err := http.Get(baseURL + "/chemicals")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var result api.Response
		json.Unmarshal(body, &result)
		printJSON(result)
	}
}

func handleInspections(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: pestctl inspections list <contract-id>")
		return
	}
	if args[0] == "list" {
		contractID := args[1]
		resp, err := http.Get(baseURL + "/contracts/" + contractID + "/inspections")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var result api.Response
		json.Unmarshal(body, &result)
		printJSON(result)
	}
}

func handleOperations(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: pestctl operations list <contract-id>")
		return
	}
	if args[0] == "list" {
		contractID := args[1]
		resp, err := http.Get(baseURL + "/contracts/" + contractID + "/operations")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var result api.Response
		json.Unmarshal(body, &result)
		printJSON(result)
	}
}

func handlePending() {
	resp, err := http.Get(baseURL + "/pending-services")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result api.Response
	json.Unmarshal(body, &result)
	printJSON(result)
}

func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func httpGet(path string) (*api.Response, error) {
	resp, err := http.Get(baseURL + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result api.Response
	err = json.Unmarshal(body, &result)
	return &result, err
}

func httpPost(path string, body interface{}) (*api.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(baseURL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result api.Response
	err = json.Unmarshal(respBody, &result)
	return &result, err
}
