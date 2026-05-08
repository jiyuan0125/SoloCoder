package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/uuid-generator/common"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	os.Args = append([]string{os.Args[0]}, os.Args[2:]...)

	switch cmd {
	case "generate":
		handleGenerate()
	case "parse":
		handleParse()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client generate [options]  - Generate UUID(s)")
	fmt.Println("  client parse <uuid>        - Parse a UUID string")
	fmt.Println()
	fmt.Println("Generate options:")
	fmt.Println("  -version int    UUID version (1-5, default: 4)")
	fmt.Println("  -count int      Number of UUIDs to generate (default: 1)")
	fmt.Println("  -namespace str  Namespace for v3/v5 (dns|url|oid|x500|custom-uuid)")
	fmt.Println("  -name str       Name for v3/v5")
	fmt.Println("  -compact        Output without hyphens")
	fmt.Println("  -server addr    Server address (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("Parse options:")
	fmt.Println("  -server addr    Server address (default: http://localhost:8080)")
}

func handleGenerate() {
	version := flag.Int("version", 4, "UUID version")
	count := flag.Int("count", 1, "Number of UUIDs")
	namespace := flag.String("namespace", "", "Namespace for v3/v5")
	name := flag.String("name", "", "Name for v3/v5")
	compact := flag.Bool("compact", false, "Output without hyphens")
	server := flag.String("server", "http://localhost:8080", "Server address")
	flag.Parse()

	req := common.GenerateRequest{
		Version:   *version,
		Namespace: *namespace,
		Name:      *name,
		Count:     *count,
	}

	resp, err := postJSON(*server+"/generate", req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var genResp common.GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		os.Exit(1)
	}

	if genResp.Error != "" {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", genResp.Error)
		os.Exit(1)
	}

	for _, u := range genResp.UUIDs {
		if *compact {
			u = strings.ReplaceAll(u, "-", "")
		}
		fmt.Println(u)
	}
}

func handleParse() {
	server := flag.String("server", "http://localhost:8080", "Server address")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: UUID string required")
		os.Exit(1)
	}

	uuidStr := flag.Arg(0)
	req := common.ParseRequest{UUID: uuidStr}

	resp, err := postJSON(*server+"/parse", req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var parseResp common.ParseResponse
	if err := json.NewDecoder(resp.Body).Decode(&parseResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		os.Exit(1)
	}

	if parseResp.Error != "" {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", parseResp.Error)
		os.Exit(1)
	}

	fmt.Printf("UUID: %s\n", uuidStr)
	fmt.Printf("Version: %d\n", parseResp.Version)
	fmt.Printf("Variant: %d\n", parseResp.Variant)
}

func postJSON(url string, data interface{}) (*http.Response, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(errorBody))
	}

	return resp, nil
}
