package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/solocoder/base64codec/pkg"
)

type Config struct {
	Mode          pkg.Mode
	Action        string
	InputFile     string
	OutputFile    string
	ServerURL     string
	Padding       bool
	MIMELineWidth int
}

func main() {
	config := parseFlags()
	
	data, err := readInput(config.InputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	
	var result string
	if config.Action == "encode" {
		result, err = encode(config, data)
	} else {
		result, err = decode(config, data)
	}
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	if err := writeOutput(config.OutputFile, result); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() *Config {
	config := &Config{
		Mode:          pkg.ModeStandard,
		Action:        "encode",
		ServerURL:     "http://localhost:8200",
		MIMELineWidth: 76,
	}
	
	modeFlag := flag.String("mode", "standard", "Encoding mode: standard, urlsafe, mime")
	actionFlag := flag.String("action", "encode", "Action: encode or decode")
	inputFlag := flag.String("input", "", "Input file (empty for stdin)")
	outputFlag := flag.String("output", "", "Output file (empty for stdout)")
	serverFlag := flag.String("server", "http://localhost:8200", "Server URL")
	paddingFlag := flag.Bool("padding", true, "Include padding for URL-safe mode")
	lineWidthFlag := flag.Int("linewidth", 76, "Line width for MIME mode")
	
	flag.Parse()
	
	switch *modeFlag {
	case "standard":
		config.Mode = pkg.ModeStandard
	case "urlsafe":
		config.Mode = pkg.ModeURLSafe
	case "mime":
		config.Mode = pkg.ModeMIME
	default:
		fmt.Fprintf(os.Stderr, "Invalid mode: %s. Using standard mode.\n", *modeFlag)
		config.Mode = pkg.ModeStandard
	}
	
	switch *actionFlag {
	case "encode":
		config.Action = "encode"
	case "decode":
		config.Action = "decode"
	default:
		fmt.Fprintf(os.Stderr, "Invalid action: %s. Using encode.\n", *actionFlag)
		config.Action = "encode"
	}
	
	config.InputFile = *inputFlag
	config.OutputFile = *outputFlag
	config.ServerURL = *serverFlag
	config.Padding = *paddingFlag
	config.MIMELineWidth = *lineWidthFlag
	
	return config
}

func readInput(filename string) (string, error) {
	var reader io.Reader
	
	if filename == "" {
		reader = os.Stdin
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return "", fmt.Errorf("failed to open input file: %v", err)
		}
		defer file.Close()
		reader = file
	}
	
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read input: %v", err)
	}
	
	return string(data), nil
}

func writeOutput(filename string, data string) error {
	var writer io.Writer
	
	if filename == "" {
		writer = os.Stdout
	} else {
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer file.Close()
		writer = file
	}
	
	_, err := io.WriteString(writer, data)
	if err != nil {
		return fmt.Errorf("failed to write output: %v", err)
	}
	
	return nil
}

func encode(config *Config, data string) (string, error) {
	req := pkg.EncodeRequest{
		Mode:          config.Mode,
		Data:          data,
		Padding:       config.Padding,
		MIMELineWidth: config.MIMELineWidth,
	}
	
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}
	
	url := fmt.Sprintf("%s/api/encode", config.ServerURL)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}
	
	var encodeResp pkg.EncodeResponse
	if err := json.Unmarshal(respBody, &encodeResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}
	
	if !encodeResp.Success {
		return "", fmt.Errorf("server error: %s", encodeResp.Error)
	}
	
	return encodeResp.Result, nil
}

func decode(config *Config, data string) (string, error) {
	req := pkg.DecodeRequest{
		Mode: config.Mode,
		Data: data,
	}
	
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}
	
	url := fmt.Sprintf("%s/api/decode", config.ServerURL)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}
	
	var decodeResp pkg.DecodeResponse
	if err := json.Unmarshal(respBody, &decodeResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}
	
	if !decodeResp.Success {
		return "", fmt.Errorf("server error: %s", decodeResp.Error)
	}
	
	return decodeResp.Result, nil
}
