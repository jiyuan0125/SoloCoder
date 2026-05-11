package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"uuencode/api"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := flag.String("server", "http://localhost:8204", "server URL")
	flag.CommandLine.Parse(os.Args[2:])

	args := flag.CommandLine.Args()

	switch os.Args[1] {
	case "encode":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "error: input file required")
			os.Exit(1)
		}
		inputFile := args[0]
		var outputFile string
		if len(args) >= 2 {
			outputFile = args[1]
		}
		encode(*serverURL, inputFile, outputFile)
	case "decode":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "error: input file required")
			os.Exit(1)
		}
		inputFile := args[0]
		var outputDir string
		if len(args) >= 2 {
			outputDir = args[1]
		}
		decode(*serverURL, inputFile, outputDir)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command %q\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: client [flags] <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  encode <input> [output]    Encode a file")
	fmt.Println("  decode <input> [output]    Decode a file")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -server string   Server URL (default \"http://localhost:8204\")")
}

func encode(serverURL, inputFile, outputFile string) {
	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input file: %v\n", err)
		os.Exit(1)
	}

	base64Data := base64.StdEncoding.EncodeToString(data)
	filename := filepath.Base(inputFile)

	req := api.EncodeRequest{
		Data:     base64Data,
		Filename: filename,
		Mode:     644,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error encoding request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/encode", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error calling server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		fmt.Fprintf(os.Stderr, "server error: %s\n", errResp.Error)
		os.Exit(1)
	}

	var respData api.EncodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		fmt.Fprintf(os.Stderr, "error decoding response: %v\n", err)
		os.Exit(1)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(respData.Encoded), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Encoded:", outputFile)
	} else {
		fmt.Print(respData.Encoded)
	}
}

func decode(serverURL, inputFile, outputDir string) {
	encoded, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input file: %v\n", err)
		os.Exit(1)
	}

	req := api.DecodeRequest{
		Encoded: string(encoded),
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error encoding request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/decode", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error calling server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "server error: %s (status %d)\n", string(body), resp.StatusCode)
		os.Exit(1)
	}

	var respData api.DecodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		fmt.Fprintf(os.Stderr, "error decoding response: %v\n", err)
		os.Exit(1)
	}

	data, err := base64.StdEncoding.DecodeString(respData.Data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error decoding base64 data: %v\n", err)
		os.Exit(1)
	}

	outputPath := respData.Filename
	if outputDir != "" {
		outputPath = filepath.Join(outputDir, respData.Filename)
	}

	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "error creating output directory: %v\n", err)
			os.Exit(1)
		}
	}

	mode, err := strconv.ParseInt(fmt.Sprintf("%d", respData.Mode), 8, 32)
	if err != nil {
		mode = 0644
	}

	if err := os.WriteFile(outputPath, data, os.FileMode(mode)); err != nil {
		fmt.Fprintf(os.Stderr, "error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Decoded:", outputPath)
}
