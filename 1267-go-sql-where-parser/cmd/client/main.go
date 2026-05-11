package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"sqlparser/pkg/common"
)

type Config struct {
	ServerAddr string
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg := getConfig()

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "parse":
		handleParse(cfg, args)
	case "validate":
		handleValidate(cfg, args)
	case "explain":
		handleExplain(cfg, args)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func getConfig() Config {
	addr := "http://localhost:8080"
	if envAddr := os.Getenv("SQL_PARSER_SERVER"); envAddr != "" {
		addr = envAddr
	}
	return Config{ServerAddr: addr}
}

func printUsage() {
	fmt.Println(`SQL WHERE Parser CLI

Usage:
  sql parse "WHERE age > 18 AND name LIKE '%张%'"   Parse and print AST
  sql validate "WHERE ..."                           Check syntax only
  sql explain "WHERE ..."                            Print parsing steps

Environment:
  SQL_PARSER_SERVER  Server address (default: http://localhost:8080)`)
}

func joinArgs(args []string) string {
	return strings.Join(args, " ")
}

func handleParse(cfg Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: parse command requires a WHERE clause")
		os.Exit(1)
	}

	where := joinArgs(args)
	req := common.ParseRequest{Where: where}

	resp, err := postJSON(cfg.ServerAddr+"/sql/parse", req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result common.ParseResponse
	json.Unmarshal(body, &result)

	if !result.Success {
		fmt.Fprintf(os.Stderr, "Parse failed: %s\n", result.Error)
		os.Exit(1)
	}

	var pretty bytes.Buffer
	json.Indent(&pretty, result.AST, "", "  ")
	fmt.Println(pretty.String())
}

func handleValidate(cfg Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: validate command requires a WHERE clause")
		os.Exit(1)
	}

	where := joinArgs(args)
	req := common.ParseRequest{Where: where}

	resp, err := postJSON(cfg.ServerAddr+"/sql/validate", req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result common.ValidateResponse
	json.Unmarshal(body, &result)

	if result.Success {
		fmt.Println("Syntax is valid")
	} else {
		fmt.Fprintf(os.Stderr, "Syntax error: %s\n", result.Error)
		os.Exit(1)
	}
}

func handleExplain(cfg Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: explain command requires a WHERE clause")
		os.Exit(1)
	}

	where := joinArgs(args)
	req := common.ParseRequest{Where: where}

	resp, err := postJSON(cfg.ServerAddr+"/sql/explain", req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result common.ExplainResponse
	json.Unmarshal(body, &result)

	fmt.Println("=== Parsing Steps ===")
	for i, step := range result.Steps {
		fmt.Printf("%d. %s\n", i+1, step)
	}

	if !result.Success {
		fmt.Fprintf(os.Stderr, "\nParse failed: %s\n", result.Error)
		os.Exit(1)
	}
}

func postJSON(url string, data interface{}) (*http.Response, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	return client.Do(req)
}
