package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/asn1-der-parser/pkg/api"
)

var serverURL = "http://localhost:8450"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "decode":
		decodeCmd(args)
	case "encode":
		encodeCmd(args)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("ASN.1 DER Parser Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client decode [options] <data>")
	fmt.Println("  client encode [options] <json-file>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  decode   Decode DER-encoded data")
	fmt.Println("  encode   Encode JSON to DER")
	fmt.Println()
	fmt.Println("Decode Options:")
	fmt.Println("  -f <file>     Read data from file")
	fmt.Println("  -format <fmt> Input format: hex, base64, raw (default: auto)")
	fmt.Println("  -url <url>    Server URL (default: http://localhost:8450)")
	fmt.Println("  -json         Output JSON instead of human-readable dump")
	fmt.Println()
	fmt.Println("Encode Options:")
	fmt.Println("  -f <file>     Read JSON from file")
	fmt.Println("  -url <url>    Server URL (default: http://localhost:8450)")
	fmt.Println("  -o <file>     Output file (default: stdout)")
	fmt.Println("  -format <fmt> Output format: hex, base64, raw (default: hex)")
}

func decodeCmd(args []string) {
	fs := flag.NewFlagSet("decode", flag.ExitOnError)
	fileFlag := fs.String("f", "", "input file")
	formatFlag := fs.String("format", "auto", "input format: hex, base64, raw")
	urlFlag := fs.String("url", serverURL, "server URL")
	jsonFlag := fs.Bool("json", false, "output JSON")

	fs.Parse(args)

	if *urlFlag != "" {
		serverURL = *urlFlag
	}

	var inputData []byte
	var err error

	if *fileFlag != "" {
		inputData, err = os.ReadFile(*fileFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
	} else if len(fs.Args()) > 0 {
		inputData = []byte(fs.Arg(0))
	} else {
		fmt.Fprintln(os.Stderr, "Error: no input data provided")
		os.Exit(1)
	}

	encodedData, err := normalizeInput(inputData, *formatFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing input: %v\n", err)
		os.Exit(1)
	}

	req := api.DecodeRequest{
		Data: base64.StdEncoding.EncodeToString(encodedData),
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/decode", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Server error (%d): %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var decoded api.DecodeResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if !decoded.Success {
		fmt.Fprintf(os.Stderr, "Decode failed: %s\n", decoded.Error)
		os.Exit(1)
	}

	if *jsonFlag {
		jsonBytes, _ := json.MarshalIndent(decoded.Result, "", "  ")
		fmt.Println(string(jsonBytes))
	} else {
		fmt.Println(decoded.Dump)
	}
}

func normalizeInput(data []byte, format string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "hex":
		s := strings.TrimSpace(string(data))
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, "\n", "")
		s = strings.ReplaceAll(s, "\r", "")
		s = strings.ReplaceAll(s, "\t", "")
		return hex.DecodeString(s)
	case "base64":
		s := strings.TrimSpace(string(data))
		return base64.StdEncoding.DecodeString(s)
	case "raw":
		return data, nil
	case "auto":
		s := strings.TrimSpace(string(data))
		if isHexString(s) {
			s = strings.ReplaceAll(s, " ", "")
			s = strings.ReplaceAll(s, "\n", "")
			s = strings.ReplaceAll(s, "\r", "")
			s = strings.ReplaceAll(s, "\t", "")
			return hex.DecodeString(s)
		}
		if isBase64String(s) {
			return base64.StdEncoding.DecodeString(s)
		}
		return data, nil
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}

func isHexString(s string) bool {
	if len(s) == 0 || len(s)%2 != 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' {
			continue
		}
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func isBase64String(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func encodeCmd(args []string) {
	fs := flag.NewFlagSet("encode", flag.ExitOnError)
	fileFlag := fs.String("f", "", "input JSON file")
	urlFlag := fs.String("url", serverURL, "server URL")
	outFlag := fs.String("o", "", "output file")
	formatFlag := fs.String("format", "hex", "output format: hex, base64, raw")

	fs.Parse(args)

	if *urlFlag != "" {
		serverURL = *urlFlag
	}

	var jsonData []byte
	var err error

	if *fileFlag != "" {
		jsonData, err = os.ReadFile(*fileFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
	} else if len(fs.Args()) > 0 {
		jsonData = []byte(fs.Arg(0))
	} else {
		fmt.Fprintln(os.Stderr, "Error: no input JSON provided")
		os.Exit(1)
	}

	var nodeJSON api.NodeJSON
	if err := json.Unmarshal(jsonData, &nodeJSON); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	req := api.EncodeRequest{
		Node: &nodeJSON,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/encode", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Server error (%d): %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var decoded api.EncodeResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if !decoded.Success {
		fmt.Fprintf(os.Stderr, "Encode failed: %s\n", decoded.Error)
		os.Exit(1)
	}

	rawData, err := base64.StdEncoding.DecodeString(decoded.Data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		os.Exit(1)
	}

	var output []byte
	switch strings.ToLower(*formatFlag) {
	case "hex":
		output = []byte(hex.EncodeToString(rawData))
	case "base64":
		output = []byte(base64.StdEncoding.EncodeToString(rawData))
	case "raw":
		output = rawData
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", *formatFlag)
		os.Exit(1)
	}

	if *outFlag != "" {
		if err := os.WriteFile(*outFlag, output, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
	} else {
		os.Stdout.Write(output)
		if *formatFlag != "raw" {
			fmt.Println()
		}
	}
}
