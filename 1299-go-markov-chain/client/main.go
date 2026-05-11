package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"markov-chain/common"
)

type Config struct {
	ServerURL   string
	InputFile   string
	OutputFile  string
	Granularity string
	Order       int
	Length      int
	Seed        int64
	Timeout     time.Duration
}

func main() {
	config := parseFlags()

	text, err := readInputFile(config.InputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input file: %v\n", err)
		os.Exit(1)
	}

	granularity, err := parseGranularity(config.Granularity)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	req := common.GenerateRequest{
		Text:        text,
		Granularity: granularity,
		Order:       config.Order,
		Length:      config.Length,
		Seed:        config.Seed,
	}

	generated, err := callServer(config.ServerURL, req, config.Timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error calling server: %v\n", err)
		os.Exit(1)
	}

	if config.OutputFile != "" {
		if err := writeOutputFile(config.OutputFile, generated); err != nil {
			fmt.Fprintf(os.Stderr, "error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("generated text written to %s\n", config.OutputFile)
	} else {
		fmt.Println(generated)
	}
}

func parseFlags() Config {
	serverURL := flag.String("server", "http://localhost:8201", "server URL")
	inputFile := flag.String("input", "", "input text file (required)")
	outputFile := flag.String("output", "", "output file (default: stdout)")
	granularity := flag.String("granularity", "char", "granularity: char or word")
	order := flag.Int("order", 2, "n-gram order")
	length := flag.Int("length", 100, "generation length")
	seed := flag.Int64("seed", time.Now().UnixNano(), "random seed")
	timeout := flag.Int("timeout", 30, "timeout in seconds")

	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintln(os.Stderr, "error: -input flag is required")
		flag.Usage()
		os.Exit(1)
	}

	return Config{
		ServerURL:   *serverURL,
		InputFile:   *inputFile,
		OutputFile:  *outputFile,
		Granularity: *granularity,
		Order:       *order,
		Length:      *length,
		Seed:        *seed,
		Timeout:     time.Duration(*timeout) * time.Second,
	}
}

func parseGranularity(g string) (common.Granularity, error) {
	switch g {
	case "char":
		return common.GranularityChar, nil
	case "word":
		return common.GranularityWord, nil
	default:
		return "", fmt.Errorf("invalid granularity '%s', use 'char' or 'word'", g)
	}
}

func readInputFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func writeOutputFile(filename string, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

func callServer(serverURL string, req common.GenerateRequest, timeout time.Duration) (string, error) {
	url := fmt.Sprintf("%s/generate", serverURL)

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{
		Timeout: timeout,
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return "", fmt.Errorf("server error: %s", errResp.Error)
		}
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var genResp common.GenerateResponse
	if err := json.Unmarshal(respBody, &genResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !genResp.Success {
		return "", fmt.Errorf("generation failed: %s", genResp.Error)
	}

	return genResp.Text, nil
}
