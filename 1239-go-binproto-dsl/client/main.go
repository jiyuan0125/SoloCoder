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

	"binproto/pkg/proto"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "parse":
		parseCommand(args)
	case "serialize":
		serializeCommand(args)
	case "validate":
		validateCommand(args)
	default:
		fmt.Printf("unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client parse <format_id> <binary_file> [--server http://localhost:8080]")
	fmt.Println("  client serialize <format_id> <json_file> <output_file> [--server http://localhost:8080]")
	fmt.Println("  client validate <dsl_file> [--server http://localhost:8080]")
}

func parseCommand(args []string) {
	fs := flag.NewFlagSet("parse", flag.ExitOnError)
	server := fs.String("server", "http://localhost:8080", "server URL")
	fs.Parse(args)

	if fs.NArg() != 2 {
		fmt.Println("Usage: client parse <format_id> <binary_file>")
		os.Exit(1)
	}

	formatID := fs.Arg(0)
	binaryFile := fs.Arg(1)

	data, err := os.ReadFile(binaryFile)
	if err != nil {
		fmt.Printf("error reading binary file: %v\n", err)
		os.Exit(1)
	}

	req := proto.ParseRequest{
		FormatID: formatID,
		Data:     data,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("error marshaling request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(*server+"/parse", "application/json", bytes.NewReader(reqBytes))
	if err != nil {
		fmt.Printf("error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var parseResp proto.ParseResponse
	if err := json.Unmarshal(respBytes, &parseResp); err != nil {
		fmt.Printf("error unmarshaling response: %v\n", err)
		os.Exit(1)
	}

	if !parseResp.Success {
		fmt.Printf("parse failed: %s\n", parseResp.Error)
		os.Exit(1)
	}

	resultJSON, _ := json.MarshalIndent(parseResp.Data, "", "  ")
	fmt.Println(string(resultJSON))
}

func serializeCommand(args []string) {
	fs := flag.NewFlagSet("serialize", flag.ExitOnError)
	server := fs.String("server", "http://localhost:8080", "server URL")
	fs.Parse(args)

	if fs.NArg() != 3 {
		fmt.Println("Usage: client serialize <format_id> <json_file> <output_file>")
		os.Exit(1)
	}

	formatID := fs.Arg(0)
	jsonFile := fs.Arg(1)
	outputFile := fs.Arg(2)

	jsonBytes, err := os.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("error reading json file: %v\n", err)
		os.Exit(1)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		fmt.Printf("error parsing json: %v\n", err)
		os.Exit(1)
	}

	req := proto.SerializeRequest{
		FormatID: formatID,
		Data:     data,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("error marshaling request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(*server+"/serialize", "application/json", bytes.NewReader(reqBytes))
	if err != nil {
		fmt.Printf("error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var serializeResp proto.SerializeResponse
	if err := json.Unmarshal(respBytes, &serializeResp); err != nil {
		fmt.Printf("error unmarshaling response: %v\n", err)
		os.Exit(1)
	}

	if !serializeResp.Success {
		fmt.Printf("serialize failed: %s\n", serializeResp.Error)
		os.Exit(1)
	}

	if err := os.WriteFile(outputFile, serializeResp.Data, 0644); err != nil {
		fmt.Printf("error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("wrote %d bytes to %s\n", len(serializeResp.Data), filepath.Base(outputFile))
}

func validateCommand(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	server := fs.String("server", "http://localhost:8080", "server URL")
	fs.Parse(args)

	if fs.NArg() != 1 {
		fmt.Println("Usage: client validate <dsl_file>")
		os.Exit(1)
	}

	dslFile := fs.Arg(0)

	dsl, err := os.ReadFile(dslFile)
	if err != nil {
		fmt.Printf("error reading dsl file: %v\n", err)
		os.Exit(1)
	}

	req := proto.ValidateRequest{
		DSL: string(dsl),
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("error marshaling request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(*server+"/validate", "application/json", bytes.NewReader(reqBytes))
	if err != nil {
		fmt.Printf("error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var validateResp proto.ValidateResponse
	if err := json.Unmarshal(respBytes, &validateResp); err != nil {
		fmt.Printf("error unmarshaling response: %v\n", err)
		os.Exit(1)
	}

	if !validateResp.Success {
		fmt.Printf("validation failed: %s\n", validateResp.Error)
		os.Exit(1)
	}

	fmt.Println("DSL is valid")
}
