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

	"encoding-detector/pkg/api"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	serverAddr := getServerAddr()

	switch command {
	case "detect":
		handleDetect(serverAddr)
	case "convert":
		handleConvert(serverAddr)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func getServerAddr() string {
	envAddr := os.Getenv("SERVER_ADDR")
	if envAddr != "" {
		return envAddr
	}
	return "http://localhost:8200"
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client detect <file>          - Detect file encoding")
	fmt.Println("  client convert <file> --to <encoding>  - Convert file encoding (default: utf-8)")
	fmt.Println("")
	fmt.Println("Environment variables:")
	fmt.Println("  SERVER_ADDR    - Server address (default: http://localhost:8200)")
}

func handleDetect(serverAddr string) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: client detect <file>")
		os.Exit(1)
	}

	filename := os.Args[2]
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	req := api.DetectRequest{
		Content: content,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error encoding request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverAddr+"/api/detect", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server error: %s\n", string(respBody))
		os.Exit(1)
	}

	var result api.DetectResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Detected encoding: %s\n", result.Encoding)
	fmt.Printf("Confidence: %.2f%%\n", result.Confidence*100)
	fmt.Printf("Has BOM: %v\n", result.HasBOM)
}

func handleConvert(serverAddr string) {
	fs := flag.NewFlagSet("convert", flag.ExitOnError)
	toEncoding := fs.String("to", "utf-8", "Target encoding")
	fs.Parse(os.Args[2:])

	if fs.NArg() < 1 {
		fmt.Println("Usage: client convert <file> [--to <encoding>]")
		os.Exit(1)
	}

	filename := fs.Arg(0)
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	req := api.ConvertRequest{
		Content:    content,
		ToEncoding: strings.ToLower(*toEncoding),
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error encoding request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverAddr+"/api/convert", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server error: %s\n", string(respBody))
		os.Exit(1)
	}

	var result api.ConvertResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("Conversion error: %s\n", result.Error)
		os.Exit(1)
	}

	os.Stdout.Write(result.Content)
}
