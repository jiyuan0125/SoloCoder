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

	"lz4service/internal/api"
)

func main() {
	compressCmd := flag.NewFlagSet("compress", flag.ExitOnError)
	decompressCmd := flag.NewFlagSet("decompress", flag.ExitOnError)

	compressText := compressCmd.String("text", "", "Text to compress")
	compressFile := compressCmd.String("file", "", "File path to compress")
	compressOutput := compressCmd.String("output", "", "Output file path")
	compressServer := compressCmd.String("server", "http://localhost:8300", "Server URL")

	decompressFile := decompressCmd.String("file", "", "Compressed file path")
	decompressOutput := decompressCmd.String("output", "", "Output file path")
	decompressServer := decompressCmd.String("server", "http://localhost:8300", "Server URL")

	if len(os.Args) < 2 {
		fmt.Println("Usage: lz4client <compress|decompress> [options]")
		fmt.Println("  compress -text <string> | -file <path> [-output <path>]")
		fmt.Println("  decompress -file <path> [-output <path>]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "compress":
		if err := compressCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		var data []byte
		var err error

		if *compressText != "" {
			data = []byte(*compressText)
		} else if *compressFile != "" {
			absPath, err := filepath.Abs(*compressFile)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			data, err = os.ReadFile(absPath)
			if err != nil {
				fmt.Printf("Error reading file: %v\n", err)
				os.Exit(1)
			}
		} else {
			data = []byte{}
		}

		compressed, err := compressData(*compressServer, data)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if *compressOutput != "" {
			absPath, err := filepath.Abs(*compressOutput)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			if err := os.WriteFile(absPath, compressed, 0644); err != nil {
				fmt.Printf("Error writing output: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Compressed data written to %s (%d bytes)\n", absPath, len(compressed))
		} else {
			fmt.Printf("Compressed data (%d bytes)\n", len(compressed))
		}

	case "decompress":
		if err := decompressCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if *decompressFile == "" {
			fmt.Println("Error: -file is required for decompress")
			os.Exit(1)
		}

		absPath, err := filepath.Abs(*decompressFile)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		compressed, err := os.ReadFile(absPath)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		decompressed, err := decompressData(*decompressServer, compressed)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if *decompressOutput != "" {
			outAbs, err := filepath.Abs(*decompressOutput)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			if err := os.WriteFile(outAbs, decompressed, 0644); err != nil {
				fmt.Printf("Error writing output: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Decompressed data written to %s (%d bytes)\n", outAbs, len(decompressed))
		} else {
			fmt.Printf("Decompressed data: %s\n", string(decompressed))
		}

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func compressData(server string, data []byte) ([]byte, error) {
	var body bytes.Buffer
	var contentType string

	req := api.CompressRequest{
		Data: data,
	}

	if len(data) > 0 {
		body.Write(data)
		contentType = "application/octet-stream"
	} else {
		contentType = "application/json"
		if err := json.NewEncoder(&body).Encode(req); err != nil {
			return nil, fmt.Errorf("failed to encode request: %v", err)
		}
	}

	resp, err := http.Post(server+"/compress", contentType, &body)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error: %s", string(respBody))
	}

	return io.ReadAll(resp.Body)
}

func decompressData(server string, data []byte) ([]byte, error) {
	var body bytes.Buffer
	var contentType string

	if len(data) > 0 {
		body.Write(data)
		contentType = "application/octet-stream"
	} else {
		req := api.DecompressRequest{}
		contentType = "application/json"
		if err := json.NewEncoder(&body).Encode(req); err != nil {
			return nil, fmt.Errorf("failed to encode request: %v", err)
		}
	}

	resp, err := http.Post(server+"/decompress", contentType, &body)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error: %s", string(respBody))
	}

	return io.ReadAll(resp.Body)
}
