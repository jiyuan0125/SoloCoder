package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type Config struct {
	ServerURL string
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	flag.CommandLine.Parse(os.Args[2:])

	config := Config{
		ServerURL: *serverURL,
	}

	cmd := os.Args[1]
	switch cmd {
	case "add":
		handleAdd(config)
	case "batch-add":
		handleBatchAdd(config)
	case "remove":
		handleRemove(config)
	case "update-deps":
		handleUpdateDeps(config)
	case "list":
		handleList(config)
	case "clear":
		handleClear(config)
	case "sort":
		handleSort(config)
	case "dot":
		handleDot(config)
	case "dot-critical":
		handleDotCritical(config)
	case "load":
		handleLoad(config)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("DAG Topological Sort Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  add <id> <name> [deps...]    Add a single task")
	fmt.Println("  batch-add <json-file>        Add tasks from JSON file")
	fmt.Println("  remove <id>                  Remove a task")
	fmt.Println("  update-deps <id> [deps...]   Update task dependencies")
	fmt.Println("  list                         List all tasks")
	fmt.Println("  clear                        Clear all tasks")
	fmt.Println("  sort                         Execute topological sort")
	fmt.Println("  dot                          Generate DOT graph")
	fmt.Println("  dot-critical                 Generate DOT graph with critical path highlighted")
	fmt.Println("  load <json-file>             Load tasks from file and execute sort")
	fmt.Println("  help                         Show this help")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -server <url>                Server URL (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("JSON File Format:")
	fmt.Println(`  {
    "tasks": [
      {"id": "A", "name": "Task A", "dependencies": []},
      {"id": "B", "name": "Task B", "dependencies": ["A"]}
    ]
  }`)
}

func mustUnmarshalJSON(data []byte, v interface{}) {
	if err := json.Unmarshal(data, v); err != nil {
		fmt.Fprintf(os.Stderr, "JSON parse error: %v\n", err)
		os.Exit(1)
	}
}
