package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"semver-server/api"
)

const serverURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "parse":
		handleParse()
	case "compare":
		handleCompare()
	case "range":
		handleRange()
	case "increment":
		handleIncrement()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client parse <version>")
	fmt.Println("  client compare <version1> <version2>")
	fmt.Println("  client range <version> <range>")
	fmt.Println("  client increment <version> <part>")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  client parse v1.2.3")
	fmt.Println("  client compare 1.0.0-alpha 1.0.0")
	fmt.Println("  client range 1.5.0 [1.0.0,2.0.0)")
	fmt.Println("  client increment 1.2.3 minor")
}

func handleParse() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: client parse <version>")
		os.Exit(1)
	}

	version := os.Args[2]
	req := api.ParseRequest{Version: version}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/parse", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.ParseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("Parsed version: %s\n", result.Version.String)
	fmt.Printf("  Major: %d\n", result.Version.Major)
	fmt.Printf("  Minor: %d\n", result.Version.Minor)
	fmt.Printf("  Patch: %d\n", result.Version.Patch)
	if len(result.Version.PreRelease) > 0 {
		fmt.Printf("  Pre-release: %s\n", strings.Join(result.Version.PreRelease, "."))
	}
}

func handleCompare() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: client compare <version1> <version2>")
		os.Exit(1)
	}

	version1 := os.Args[2]
	version2 := os.Args[3]
	req := api.CompareRequest{Version1: version1, Version2: version2}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/compare", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.CompareResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	switch result.Result {
	case -1:
		fmt.Printf("%s is less than %s\n", version1, version2)
	case 0:
		fmt.Printf("%s is equal to %s\n", version1, version2)
	case 1:
		fmt.Printf("%s is greater than %s\n", version1, version2)
	}
}

func handleRange() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: client range <version> <range>")
		os.Exit(1)
	}

	version := os.Args[2]
	versionRange := os.Args[3]
	req := api.RangeCheckRequest{Version: version, Range: versionRange}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/range", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.RangeCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	if result.InRange {
		fmt.Printf("%s is in range %s\n", version, versionRange)
	} else {
		fmt.Printf("%s is NOT in range %s\n", version, versionRange)
	}
}

func handleIncrement() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: client increment <version> <part>")
		os.Exit(1)
	}

	version := os.Args[2]
	part := os.Args[3]
	req := api.IncrementRequest{Version: version, Part: part}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/increment", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.IncrementResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("Incremented %s version of %s: %s\n", part, version, result.Version.String)
}
