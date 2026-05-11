package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/example/msgpack-codec/pkg/protocol"
)

type config struct {
	serverURL  string
	operation  string
	inputFile  string
	outputFile string
}

func main() {
	cfg := parseFlags()

	if cfg.operation == "" {
		fmt.Println("Error: operation is required (encode or decode)")
		os.Exit(1)
	}

	if cfg.inputFile == "" {
		fmt.Println("Error: input file is required")
		os.Exit(1)
	}

	if cfg.outputFile == "" {
		cfg.outputFile = determineOutputFile(cfg.inputFile, cfg.operation)
	}

	inputData, err := os.ReadFile(cfg.inputFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var endpoint string
	var contentType string

	switch cfg.operation {
	case "encode":
		endpoint = cfg.serverURL + protocol.EndpointEncode
		contentType = "application/json"
	case "decode":
		endpoint = cfg.serverURL + protocol.EndpointDecode
		contentType = "application/octet-stream"
	default:
		fmt.Printf("Error: unknown operation '%s'\n", cfg.operation)
		os.Exit(1)
	}

	resp, err := http.Post(endpoint, contentType, bytes.NewReader(inputData))
	if err != nil {
		fmt.Printf("Error sending request to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp protocol.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != "" {
			fmt.Printf("Server error: %s\n", errResp.Error)
		} else {
			fmt.Printf("Server returned status %d: %s\n", resp.StatusCode, string(respBody))
		}
		os.Exit(1)
	}

	if err := os.WriteFile(cfg.outputFile, respBody, 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Success! %s written to %s\n", cfg.operation, cfg.outputFile)
}

func parseFlags() config {
	var cfg config

	flag.StringVar(&cfg.serverURL, "server", "http://localhost:8801", "Server URL (default: http://localhost:8801)")
	flag.StringVar(&cfg.operation, "op", "", "Operation: encode or decode")
	flag.StringVar(&cfg.inputFile, "in", "", "Input file path")
	flag.StringVar(&cfg.outputFile, "out", "", "Output file path (optional, auto-determined if not provided)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "MsgPack Codec Client\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s -op encode -in input.json [-out output.msgpack]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -op decode -in input.msgpack [-out output.json]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	return cfg
}

func determineOutputFile(inputFile, operation string) string {
	ext := filepath.Ext(inputFile)
	base := inputFile[:len(inputFile)-len(ext)]

	switch operation {
	case "encode":
		if ext == ".json" {
			return base + ".msgpack"
		}
		return inputFile + ".msgpack"
	case "decode":
		if ext == ".msgpack" {
			return base + ".json"
		}
		return inputFile + ".json"
	default:
		return inputFile + ".out"
	}
}
