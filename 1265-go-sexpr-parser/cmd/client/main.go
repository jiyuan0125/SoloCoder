package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"sexpr/internal/api"
)

const defaultServer = "http://localhost:8080"

func getServerURL() string {
	if url := os.Getenv("SEXP_SERVER"); url != "" {
		return url
	}
	return defaultServer
}

func callParse(input string) (*api.ParseResponse, error) {
	reqBody, err := json.Marshal(api.ParseRequest{Input: input})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(getServerURL()+"/sexp/parse", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result api.ParseResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &result, nil
}

func callEval(input string) (*api.EvalResponse, error) {
	reqBody, err := json.Marshal(api.EvalRequest{Input: input})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(getServerURL()+"/sexp/eval", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result api.EvalResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &result, nil
}

func callFormat(input string) (*api.FormatResponse, error) {
	reqBody, err := json.Marshal(api.FormatRequest{Input: input})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(getServerURL()+"/sexp/format", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result api.FormatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &result, nil
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func cmdParse(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: sexp parse <file>")
	}

	content, err := readFile(args[0])
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	result, err := callParse(content)
	if err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("parse error: %s", result.Error)
	}

	jsonData, err := json.MarshalIndent(result.Result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}

func cmdEval(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: sexp eval \"<expression>\"")
	}

	result, err := callEval(args[0])
	if err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("eval error: %s", result.Error)
	}

	fmt.Println(result.Result)
	return nil
}

func cmdFormat(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: sexp format <file>")
	}

	content, err := readFile(args[0])
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	result, err := callFormat(content)
	if err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("format error: %s", result.Error)
	}

	fmt.Println(result.Result)
	return nil
}

func printUsage() {
	fmt.Println("Usage: sexp <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  parse <file>    Parse a file and print JSON result")
	fmt.Println("  eval \"<expr>\"   Parse and evaluate an expression")
	fmt.Println("  format <file>   Reformat a file with proper indentation")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  SEXP_SERVER     Server URL (default: http://localhost:8080)")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error

	switch command {
	case "parse":
		err = cmdParse(args)
	case "eval":
		err = cmdEval(args)
	case "format":
		err = cmdFormat(args)
	case "help", "--help", "-h":
		printUsage()
		os.Exit(0)
	default:
		err = fmt.Errorf("unknown command: %s", command)
		printUsage()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
