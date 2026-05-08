package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/solocoder/resp-parser/api"
	"github.com/solocoder/resp-parser/resp"
)

type Client struct {
	serverURL string
	timeout   time.Duration
	client    *http.Client
}

func NewClient(serverURL string, timeout time.Duration) *Client {
	return &Client{
		serverURL: strings.TrimSuffix(serverURL, "/"),
		timeout:   timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Execute(commands [][]string) ([]resp.Value, error) {
	reqBody, err := resp.EncodeCommand(commands[0])
	if err != nil {
		return nil, fmt.Errorf("failed to encode command: %v", err)
	}
	
	for i := 1; i < len(commands); i++ {
		cmd, err := resp.EncodeCommand(commands[i])
		if err != nil {
			return nil, fmt.Errorf("failed to encode command: %v", err)
		}
		reqBody = append(reqBody, cmd...)
	}

	req, err := http.NewRequest("POST", c.serverURL+"/raw", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/resp")

	httpResp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("server error (%d): %s", httpResp.StatusCode, string(body))
	}

	parser := resp.NewParser(httpResp.Body)
	var results []resp.Value
	
	for {
		v, err := parser.Parse()
		if err == io.EOF {
			break
		}
		if err != nil {
			return results, fmt.Errorf("failed to parse response: %v", err)
		}
		results = append(results, v)
	}

	return results, nil
}

func (c *Client) ExecuteJSON(commands [][]string) (*api.Response, error) {
	req := api.Request{Commands: commands}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", c.serverURL+"/command", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("server error (%d): %s", httpResp.StatusCode, string(body))
	}

	var result api.Response
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

func formatValue(v resp.Value, indent int) string {
	prefix := strings.Repeat("  ", indent)
	
	switch v.Type {
	case resp.TypeSimpleString:
		return fmt.Sprintf("%s\"%s\"", prefix, v.Str)
	case resp.TypeError:
		return fmt.Sprintf("%s(error) %s", prefix, v.Str)
	case resp.TypeInteger:
		return fmt.Sprintf("%s(integer) %d", prefix, v.Int)
	case resp.TypeBulkString:
		return fmt.Sprintf("%s\"%s\"", prefix, v.Str)
	case resp.TypeNullBulkString, resp.TypeNullArray:
		return fmt.Sprintf("%s(nil)", prefix)
	case resp.TypeArray:
		if v.IsNull {
			return fmt.Sprintf("%s(nil array)", prefix)
		}
		if len(v.Array) == 0 {
			return fmt.Sprintf("%s(empty list)", prefix)
		}
		var lines []string
		lines = append(lines, fmt.Sprintf("%s%d) (list)", prefix, len(v.Array)))
		for i, elem := range v.Array {
			lines = append(lines, fmt.Sprintf("%s%d) %s", prefix, i+1, formatValue(elem, 0)))
		}
		return strings.Join(lines, "\n")
	default:
		return fmt.Sprintf("%s(unknown type %d)", prefix, v.Type)
	}
}

func printResult(v resp.Value, index int) {
	fmt.Printf("=== Response %d ===\n", index+1)
	fmt.Println(formatValue(v, 0))
	fmt.Println()
}

func runInteractive(client *Client) {
	fmt.Println("RESP Protocol Client")
	fmt.Println("Type 'help' for commands, 'exit' or 'quit' to exit")
	fmt.Println("Use 'pipeline' to enter pipeline mode")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt, exiting...")
		os.Exit(0)
	}()

	for {
		fmt.Print("resp> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println()
				break
			}
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch strings.ToLower(line) {
		case "exit", "quit":
			fmt.Println("Goodbye!")
			return
		case "help":
			printHelp()
			continue
		case "pipeline":
			runPipelineMode(client, reader)
			continue
		}

		args, err := resp.ParseInlineCommand(line)
		if err != nil {
			fmt.Printf("Command parse error: %v\n", err)
			continue
		}

		if len(args) == 0 {
			continue
		}

		results, err := client.Execute([][]string{args})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		for i, r := range results {
			printResult(r, i)
		}
	}
}

func runPipelineMode(client *Client, reader *bufio.Reader) {
	fmt.Println("Pipeline mode - enter commands, one per line")
	fmt.Println("Enter empty line or 'done' to execute all commands")
	fmt.Println()

	var commands [][]string

	for {
		fmt.Print("pipe> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return
		}

		line = strings.TrimSpace(line)
		
		if line == "" || strings.ToLower(line) == "done" {
			break
		}

		args, err := resp.ParseInlineCommand(line)
		if err != nil {
			fmt.Printf("Command parse error: %v\n", err)
			continue
		}

		if len(args) > 0 {
			commands = append(commands, args)
		}
	}

	if len(commands) == 0 {
		fmt.Println("No commands entered")
		return
	}

	fmt.Printf("\nExecuting %d commands...\n", len(commands))
	
	results, err := client.Execute(commands)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for i, r := range results {
		printResult(r, i)
	}
}

func printHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  SET key value      - Set a key to hold a string value")
	fmt.Println("  GET key            - Get the value of a key")
	fmt.Println("  DEL key [key ...]  - Delete one or more keys")
	fmt.Println("  EXISTS key [key ...] - Check if keys exist")
	fmt.Println("  INCR key           - Increment the integer value of a key")
	fmt.Println("  EXPIRE key seconds - Set a timeout on a key")
	fmt.Println("  TTL key            - Get the time to live for a key")
	fmt.Println("  PING [message]     - Ping the server")
	fmt.Println("  KEYS               - List all keys")
	fmt.Println("\nSpecial commands:")
	fmt.Println("  pipeline           - Enter pipeline mode")
	fmt.Println("  help               - Show this help message")
	fmt.Println("  exit / quit        - Exit the client")
	fmt.Println()
}

func main() {
	var serverURL string
	var timeout time.Duration
	var command string
	var showHelp bool

	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Server URL")
	flag.DurationVar(&timeout, "timeout", 30*time.Second, "Request timeout")
	flag.StringVar(&command, "c", "", "Execute a single command and exit")
	flag.BoolVar(&showHelp, "h", false, "Show help message")
	flag.Parse()

	if showHelp {
		fmt.Println("RESP Protocol Client")
		fmt.Println()
		fmt.Println("Usage: client [options]")
		fmt.Println()
		flag.PrintDefaults()
		return
	}

	client := NewClient(serverURL, timeout)

	if command != "" {
		args, err := resp.ParseInlineCommand(command)
		if err != nil {
			fmt.Printf("Command parse error: %v\n", err)
			os.Exit(1)
		}

		if len(args) == 0 {
			fmt.Println("Empty command")
			os.Exit(1)
		}

		results, err := client.Execute([][]string{args})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		for i, r := range results {
			printResult(r, i)
		}
		return
	}

	runInteractive(client)
}
