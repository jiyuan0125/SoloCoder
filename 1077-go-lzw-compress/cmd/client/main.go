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

	"github.com/lzwtool/lzwcompress/pkg/api"
)

func printUsage() {
	fmt.Println("LZW Compression Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  lzw compress -in <input_file> -out <output_file> [-min 8] [-server http://localhost:8080]")
	fmt.Println("  lzw decompress -in <input_file> -out <output_file> [-min 8] [-server http://localhost:8080]")
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "compress":
		compress()
	case "decompress":
		decompress()
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func compress() {
	fs := flag.NewFlagSet("compress", flag.ExitOnError)
	inFile := fs.String("in", "", "Input file path (required)")
	outFile := fs.String("out", "", "Output file path (required)")
	minCodeSize := fs.Int("min", 8, "Minimum code size (2-8)")
	server := fs.String("server", "http://localhost:8080", "Server address")

	err := fs.Parse(os.Args[2:])
	if err != nil {
		os.Exit(1)
	}

	if *inFile == "" || *outFile == "" {
		fmt.Println("Error: -in and -out are required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	data, err := os.ReadFile(*inFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	req := api.CompressRequest{
		Data:        data,
		MinCodeSize: *minCodeSize,
	}

	resp, err := sendRequest(*server+api.EndpointCompress, req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}

	var compressResp api.CompressResponse
	err = json.Unmarshal(resp, &compressResp)
	if err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}

	if compressResp.Error != "" {
		fmt.Printf("Server error: %s\n", compressResp.Error)
		os.Exit(1)
	}

	outDir := filepath.Dir(*outFile)
	if outDir != "." && outDir != "" {
		err = os.MkdirAll(outDir, 0755)
		if err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}
	}

	err = os.WriteFile(*outFile, compressResp.Data, 0644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Compressed %s -> %s\n", *inFile, *outFile)
	fmt.Printf("Original size: %d bytes\n", len(data))
	fmt.Printf("Compressed size: %d bytes\n", len(compressResp.Data))
}

func decompress() {
	fs := flag.NewFlagSet("decompress", flag.ExitOnError)
	inFile := fs.String("in", "", "Input file path (required)")
	outFile := fs.String("out", "", "Output file path (required)")
	minCodeSize := fs.Int("min", 8, "Minimum code size (2-8)")
	server := fs.String("server", "http://localhost:8080", "Server address")

	err := fs.Parse(os.Args[2:])
	if err != nil {
		os.Exit(1)
	}

	if *inFile == "" || *outFile == "" {
		fmt.Println("Error: -in and -out are required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	data, err := os.ReadFile(*inFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	req := api.DecompressRequest{
		Data:        data,
		MinCodeSize: *minCodeSize,
	}

	resp, err := sendRequest(*server+api.EndpointDecompress, req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}

	var decompressResp api.DecompressResponse
	err = json.Unmarshal(resp, &decompressResp)
	if err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}

	if decompressResp.Error != "" {
		fmt.Printf("Server error: %s\n", decompressResp.Error)
		os.Exit(1)
	}

	outDir := filepath.Dir(*outFile)
	if outDir != "." && outDir != "" {
		err = os.MkdirAll(outDir, 0755)
		if err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}
	}

	err = os.WriteFile(*outFile, decompressResp.Data, 0644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Decompressed %s -> %s\n", *inFile, *outFile)
	fmt.Printf("Compressed size: %d bytes\n", len(data))
	fmt.Printf("Original size: %d bytes\n", len(decompressResp.Data))
}

func sendRequest(url string, req interface{}) ([]byte, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
