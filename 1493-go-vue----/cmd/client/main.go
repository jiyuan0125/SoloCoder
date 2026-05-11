package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const defaultServer = "http://localhost:9013"

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	if serverURL == "" {
		serverURL = defaultServer
	}
	if !strings.HasPrefix(serverURL, "http") {
		serverURL = "http://" + serverURL
	}
	return &Client{serverURL: serverURL}
}

func (c *Client) do(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.serverURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

func (c *Client) get(path string) ([]byte, error) {
	return c.do(http.MethodGet, path, nil)
}

func (c *Client) post(path string, body interface{}) ([]byte, error) {
	return c.do(http.MethodPost, path, body)
}

func printHelp() {
	fmt.Println(`Cleaning Service CLI

Usage:
  client [options] <command> [args]

Options:
  --server <url>    Server URL (default: http://localhost:9013)

Commands:
  help                        Show this help message
  
  Client Management:
    client list               List all clients
    client get <id>           Get client by ID
    
  Zone & Service Area:
    zone list                 List all zones
    zone create <name>        Create a zone
    area list                 List all service areas
    
  Team & Cleaner:
    team list                 List all teams
    cleaner list              List all cleaners
    
  Schedule & Leave:
    leave list                List all leave requests
    
  Task Management:
    task list                 List all tasks
    task generate <start> <end>  Generate weekly tasks (dates: YYYY-MM-DD)
    task complete <task_id>   Complete a task
    
  Quality & Todos:
    qc list                   List all quality checks
    todo list                 List all todos
    todo complete <id>        Complete a todo`)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	args := os.Args[1:]
	serverURL := defaultServer

	if len(args) >= 2 && args[0] == "--server" {
		serverURL = args[1]
		args = args[2:]
	}

	if len(args) == 0 {
		printHelp()
		os.Exit(1)
	}

	client := NewClient(serverURL)
	cmd := args[0]

	switch cmd {
	case "help":
		printHelp()
	case "list":
		handleList(client, args[1:])
	case "get":
		handleGet(client, args[1:])
	case "zone":
		handleZone(client, args[1:])
	case "area":
		handleArea(client, args[1:])
	case "team":
		handleTeam(client, args[1:])
	case "cleaner":
		handleCleaner(client, args[1:])
	case "leave":
		handleLeave(client, args[1:])
	case "task":
		handleTask(client, args[1:])
	case "qc":
		handleQC(client, args[1:])
	case "todo":
		handleTodo(client, args[1:])
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printJSON(data []byte) {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err != nil {
		fmt.Println(string(data))
		return
	}
	fmt.Println(pretty.String())
}

func handleList(c *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client list <clients|zones|areas|teams|cleaners|leaves|tasks|qcs|todos>")
		os.Exit(1)
	}
	resource := args[0]

	var path string
	switch resource {
	case "clients":
		path = "/api/clients"
	case "zones":
		path = "/api/zones"
	case "areas":
		path = "/api/service-areas"
	case "teams":
		path = "/api/teams"
	case "cleaners":
		path = "/api/cleaners"
	case "leaves":
		path = "/api/leaves"
	case "tasks":
		path = "/api/tasks"
	case "qcs":
		path = "/api/quality-checks"
	case "todos":
		path = "/api/todos"
	default:
		fmt.Printf("Unknown resource: %s\n", resource)
		os.Exit(1)
	}

	data, err := c.get(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	printJSON(data)
}

func handleGet(c *Client, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: client get client <id>")
		os.Exit(1)
	}
	resource := args[0]
	id := args[1]

	var path string
	switch resource {
	case "client":
		path = "/api/clients/" + id
	default:
		fmt.Printf("Unknown resource: %s\n", resource)
		os.Exit(1)
	}

	data, err := c.get(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	printJSON(data)
}

func handleZone(c *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client zone list|create <name>")
		os.Exit(1)
	}
	sub := args[0]

	switch sub {
	case "list":
		data, err := c.get("/api/zones")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(data)
	case "create":
		if len(args) < 2 {
			fmt.Println("Usage: client zone create <name>")
			os.Exit(1)
		}
		data, err := c.post("/api/zones", map[string]string{"name": args[1]})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(data)
	default:
		fmt.Printf("Unknown subcommand: zone %s\n", sub)
		os.Exit(1)
	}
}

func handleArea(c *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client area list")
		os.Exit(1)
	}
	sub := args[0]

	if sub == "list" {
		data, err := c.get("/api/service-areas")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(data)
	} else {
		fmt.Printf("Unknown subcommand: area %s\n", sub)
		os.Exit(1)
	}
}

func handleTeam(c *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client team list")
		os.Exit(1)
	}
	sub := args[0]

	if sub == "list" {
		data, err := c.get("/api/teams")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(data)
	} else {
		fmt.Printf("Unknown subcommand: team %s\n", sub)
		os.Exit(1)
	}
}

func handleCleaner(c *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client cleaner list")
		os.Exit(1)
	}
	sub := args[0]

	if sub == "list" {
		data, err := c.get("/api/cleaners")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(data)
	} else {
		fmt.Printf("Unknown subcommand: cleaner %s\n", sub)
		os.Exit(1)
	}
}

func handleLeave(c *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client leave list")
		os.Exit(1)
	}
	sub := args[0]

	if sub == "list" {
		data, err := c.get("/api/leaves")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(data)
	} else {
		fmt.Printf("Unknown subcommand: leave %s\n", sub)
		os.Exit(1)
	}
}

func handleTask(c *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client task list|generate|complete")
		os.Exit(1)
	}
	sub := args[0]

	switch sub {
	case "list":
		data, err := c.get("/api/tasks")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(data)
	case "generate":
		if len(args) < 3 {
			fmt.Println("Usage: client task generate <start> <end> (dates: YYYY-MM-DD)")
			os.Exit(1)
		}
		body := map[string]interface{}{
			"week_start": args[1] + "T00:00:00Z",
			"week_end":   args[2] + "T23:59:59Z",
		}
		data, err := c.post("/api/tasks/generate", body)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(data)
	case "complete":
		if len(args) < 2 {
			fmt.Println("Usage: client task complete <task_id>")
			os.Exit(1)
		}
		body := map[string]string{"task_id": args[1]}
		data, err := c.post("/api/tasks/complete", body)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(data)
	default:
		fmt.Printf("Unknown subcommand: task %s\n", sub)
		os.Exit(1)
	}
}

func handleQC(c *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client qc list")
		os.Exit(1)
	}
	sub := args[0]

	if sub == "list" {
		data, err := c.get("/api/quality-checks")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(data)
	} else {
		fmt.Printf("Unknown subcommand: qc %s\n", sub)
		os.Exit(1)
	}
}

func handleTodo(c *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client todo list|complete <id>")
		os.Exit(1)
	}
	sub := args[0]

	switch sub {
	case "list":
		data, err := c.get("/api/todos")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(data)
	case "complete":
		if len(args) < 2 {
			fmt.Println("Usage: client todo complete <id>")
			os.Exit(1)
		}
		data, err := c.post("/api/todos/"+args[1], map[string]interface{}{})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(data)
	default:
		fmt.Printf("Unknown subcommand: todo %s\n", sub)
		os.Exit(1)
	}
}
