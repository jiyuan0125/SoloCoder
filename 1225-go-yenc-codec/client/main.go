package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"yenc-codec/common"
)

const (
	defaultServerURL = "http://localhost:8206"
	envServerURLKey  = "YENC_SERVER_URL"
)

func getServerURL() string {
	if url := os.Getenv(envServerURLKey); url != "" {
		return url
	}
	return defaultServerURL
}

func printUsage() {
	execName := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage: %s <command> [options]\n", execName)
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  encode <file>          Encode a file to yEnc format\n")
	fmt.Fprintf(os.Stderr, "  decode <file>          Decode a yEnc encoded file\n")
	fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
	fmt.Fprintf(os.Stderr, "  %s    Server URL (default: %s)\n", envServerURLKey, defaultServerURL)
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func doEncode(serverURL, filePath string) error {
	data, err := readFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	base64Data := base64.StdEncoding.EncodeToString(data)
	req := common.EncodeRequest{
		Data: base64Data,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(serverURL+"/encode", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var encodeResp common.EncodeResponse
	if err := json.Unmarshal(body, &encodeResp); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if !encodeResp.Success {
		return fmt.Errorf("server error: %s", encodeResp.Message)
	}

	yencData, err := base64.StdEncoding.DecodeString(encodeResp.Encoded)
	if err != nil {
		return fmt.Errorf("failed to decode base64 yEnc data: %v", err)
	}

	outputPath := filePath + ".yenc"
	if err := writeFile(outputPath, yencData); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	fmt.Printf("Successfully encoded %s -> %s\n", filePath, outputPath)
	fmt.Printf("Original size: %d bytes\n", len(data))
	return nil
}

func doDecode(serverURL, filePath string) error {
	data, err := readFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	base64YEnc := base64.StdEncoding.EncodeToString(data)

	req := common.DecodeRequest{
		Encoded: base64YEnc,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(serverURL+"/decode", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var decodeResp common.DecodeResponse
	if err := json.Unmarshal(body, &decodeResp); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if !decodeResp.Success {
		return fmt.Errorf("server error: %s", decodeResp.Message)
	}

	decodedData, err := base64.StdEncoding.DecodeString(decodeResp.Data)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %v", err)
	}

	var outputPath string
	ext := filepath.Ext(filePath)
	if ext == ".yenc" {
		outputPath = filePath[:len(filePath)-len(ext)]
	} else {
		outputPath = filePath + ".decoded"
	}

	if err := writeFile(outputPath, decodedData); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	fmt.Printf("Successfully decoded %s -> %s\n", filePath, outputPath)
	fmt.Printf("Decoded size: %d bytes\n", len(decodedData))
	return nil
}

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	filePath := os.Args[2]
	serverURL := getServerURL()

	switch command {
	case "encode":
		if err := doEncode(serverURL, filePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "decode":
		if err := doDecode(serverURL, filePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}
