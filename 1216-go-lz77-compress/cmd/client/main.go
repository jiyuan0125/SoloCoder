package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"lz77-compressor/pkg/api"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := strings.ToLower(os.Args[1])

	switch command {
	case "compress":
		if len(os.Args) < 3 {
			fmt.Println("Usage: client compress <text>")
			os.Exit(1)
		}
		text := strings.Join(os.Args[2:], " ")
		compress(text)
	case "decompress":
		if len(os.Args) < 3 {
			fmt.Println("Usage: client decompress <encoded_data>")
			os.Exit(1)
		}
		encodedData := strings.Join(os.Args[2:], " ")
		decompress(encodedData)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("LZ77 Compression Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client compress <text>      - Compress text using LZ77")
	fmt.Println("  client decompress <data>    - Decompress LZ77 encoded data")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  LZ77_SERVER_URL             - Server URL (default: http://localhost:8080)")
}

func getServerURL() string {
	if url := os.Getenv("LZ77_SERVER_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

func compress(text string) {
	serverURL := getServerURL()

	req := api.CompressRequest{
		Text: text,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("Failed to encode request: %v", err)
	}

	resp, err := http.Post(serverURL+"/compress", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	var result api.CompressResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Println("Compression Result:")
	fmt.Printf("  Original Size:  %d bytes\n", result.OriginalSize)
	fmt.Printf("  Compressed Size: %d bytes\n", result.CompressedSize)
	fmt.Printf("  Compression Ratio: %.2f%%\n", result.Ratio*100)
	fmt.Println()
	fmt.Println("Compressed Data (base64 encoded):")
	fmt.Println(result.CompressedData)
}

func decompress(encodedData string) {
	serverURL := getServerURL()

	req := api.DecompressRequest{
		CompressedData: encodedData,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("Failed to encode request: %v", err)
	}

	resp, err := http.Post(serverURL+"/decompress", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	var result api.DecompressResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Println("Decompression Result:")
	fmt.Println()
	fmt.Println("Original Text:")
	fmt.Println(result.Text)
}
