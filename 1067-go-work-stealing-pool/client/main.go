package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/workstealing/pool/common"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverAddr := flag.String("server", "http://localhost:8300", "Server address")
	flag.Parse()

	client := NewTaskClient(*serverAddr)

	command := strings.ToLower(os.Args[1])
	args := os.Args[2:]

	switch command {
	case "submit":
		handleSubmit(client, args)
	case "result":
		handleResult(client, args)
	case "stats":
		handleStats(client)
	case "add-workers":
		handleAddWorkers(client, args)
	case "remove-workers":
		handleRemoveWorkers(client, args)
	case "shutdown":
		handleShutdown(client, args)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleSubmit(client *TaskClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: submit <task_type> [payload] [timeout_ms]")
		fmt.Println("  task_type: fib | echo | sleep | panic")
		os.Exit(1)
	}

	taskType := args[0]
	var payload interface{}
	var timeout time.Duration

	if len(args) >= 2 {
		var raw json.RawMessage
		if err := json.Unmarshal([]byte(args[1]), &raw); err == nil {
			json.Unmarshal(raw, &payload)
		} else {
			payload = args[1]
		}
	}

	if len(args) >= 3 {
		var ms int
		fmt.Sscanf(args[2], "%d", &ms)
		timeout = time.Duration(ms) * time.Millisecond
	}

	resp, err := client.SubmitTask(taskType, payload, timeout)
	if err != nil {
		log.Fatalf("Submit failed: %v", err)
	}

	if resp.Success {
		fmt.Printf("Task submitted: %s\n", resp.TaskID)
	} else {
		fmt.Printf("Submit failed: %s\n", resp.Error)
	}
}

func handleResult(client *TaskClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: result <task_id> [--wait]")
		os.Exit(1)
	}

	taskID := args[0]
	wait := false
	if len(args) >= 2 && args[1] == "--wait" {
		wait = true
	}

	if wait {
		for {
			resp, err := client.GetResult(taskID)
			if err != nil {
				log.Fatalf("Get result failed: %v", err)
			}
			if resp.Done {
				printResult(resp)
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	} else {
		resp, err := client.GetResult(taskID)
		if err != nil {
			log.Fatalf("Get result failed: %v", err)
		}
		printResult(resp)
	}
}

func printResult(resp *common.GetTaskResultResponse) {
	out, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(out))
}

func handleStats(client *TaskClient) {
	resp, err := client.GetStats()
	if err != nil {
		log.Fatalf("Get stats failed: %v", err)
	}
	out, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(out))
}

func handleAddWorkers(client *TaskClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: add-workers <count>")
		os.Exit(1)
	}
	var count int
	fmt.Sscanf(args[0], "%d", &count)

	resp, err := client.AdjustWorkers(count, 0)
	if err != nil {
		log.Fatalf("Adjust workers failed: %v", err)
	}
	out, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(out))
}

func handleRemoveWorkers(client *TaskClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: remove-workers <count>")
		os.Exit(1)
	}
	var count int
	fmt.Sscanf(args[0], "%d", &count)

	resp, err := client.AdjustWorkers(0, count)
	if err != nil {
		log.Fatalf("Adjust workers failed: %v", err)
	}
	out, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(out))
}

func handleShutdown(client *TaskClient, args []string) {
	force := false
	if len(args) >= 1 && args[0] == "--force" {
		force = true
	}

	resp, err := client.Shutdown(force)
	if err != nil {
		log.Fatalf("Shutdown failed: %v", err)
	}
	out, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(out))
}

func printUsage() {
	fmt.Println("Work-stealing Pool Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client submit <task_type> [payload] [timeout_ms]   Submit a task")
	fmt.Println("  client result <task_id> [--wait]                   Get task result")
	fmt.Println("  client stats                                       Get pool stats")
	fmt.Println("  client add-workers <count>                         Add workers")
	fmt.Println("  client remove-workers <count>                      Remove workers")
	fmt.Println("  client shutdown [--force]                          Shutdown server")
	fmt.Println()
	fmt.Println("Task types:")
	fmt.Println("  fib <number>          - Compute Fibonacci number")
	fmt.Println("  echo <text>           - Echo text (logs it)")
	fmt.Println("  sleep <ms>            - Sleep for milliseconds")
	fmt.Println("  panic <msg>           - Trigger a panic (for testing)")
}
