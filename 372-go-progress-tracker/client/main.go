package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"progress-tracker/common"
)

const (
	defaultServerURL = "http://localhost:8080"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := os.Getenv("PROGRESS_SERVER_URL")
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	client := NewAPIClient(serverURL)

	command := os.Args[1]
	switch command {
	case "create":
		handleCreate(client, os.Args[2:])
	case "get":
		handleGet(client, os.Args[2:])
	case "list":
		handleList(client)
	case "update":
		handleUpdate(client, os.Args[2:])
	case "set":
		handleSet(client, os.Args[2:])
	case "cancel":
		handleCancel(client, os.Args[2:])
	case "close":
		handleClose(client, os.Args[2:])
	case "subtask":
		handleSubTask(client, os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Progress Tracker Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client create <total> [description]  - Create new progress")
	fmt.Println("  client get <id>                       - Get specific progress")
	fmt.Println("  client list                           - List all progress")
	fmt.Println("  client update <id> <delta>            - Update progress by delta")
	fmt.Println("  client set <id> <current>             - Set progress to specific value")
	fmt.Println("  client cancel <id>                     - Cancel progress")
	fmt.Println("  client close <id>                      - Close progress")
	fmt.Println("  client subtask <parent_id> <total> [weight] - Add subtask")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  PROGRESS_SERVER_URL - Server URL (default: http://localhost:8080)")
}

func handleCreate(client *APIClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Error: total is required")
		os.Exit(1)
	}

	total, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Printf("Error: invalid total value: %v\n", err)
		os.Exit(1)
	}

	description := ""
	if len(args) > 1 {
		description = args[1]
	}

	resp, err := client.CreateProgress(total, description)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Progress created: ID=%s, Status=%s\n", resp.ID, resp.Status)
}

func handleGet(client *APIClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Error: progress ID is required")
		os.Exit(1)
	}

	id := args[0]
	status, err := client.GetProgress(id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printProgressStatus(status)
}

func handleList(client *APIClient) {
	progresses, err := client.ListProgress()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(progresses) == 0 {
		fmt.Println("No progress tasks found")
		return
	}

	fmt.Printf("Found %d progress task(s):\n", len(progresses))
	fmt.Println("---")
	for _, p := range progresses {
		printProgressStatus(&p)
		fmt.Println("---")
	}
}

func handleUpdate(client *APIClient, args []string) {
	if len(args) < 2 {
		fmt.Println("Error: progress ID and delta are required")
		os.Exit(1)
	}

	id := args[0]
	delta, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		fmt.Printf("Error: invalid delta value: %v\n", err)
		os.Exit(1)
	}

	status, err := client.UpdateProgress(id, delta)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printProgressStatus(status)
}

func handleSet(client *APIClient, args []string) {
	if len(args) < 2 {
		fmt.Println("Error: progress ID and current value are required")
		os.Exit(1)
	}

	id := args[0]
	current, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		fmt.Printf("Error: invalid current value: %v\n", err)
		os.Exit(1)
	}

	status, err := client.SetProgress(id, current)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printProgressStatus(status)
}

func handleCancel(client *APIClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Error: progress ID is required")
		os.Exit(1)
	}

	id := args[0]
	status, err := client.CancelProgress(id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printProgressStatus(status)
}

func handleClose(client *APIClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Error: progress ID is required")
		os.Exit(1)
	}

	id := args[0]
	err := client.CloseProgress(id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Progress %s closed successfully\n", id)
}

func handleSubTask(client *APIClient, args []string) {
	if len(args) < 2 {
		fmt.Println("Error: parent ID and total are required")
		os.Exit(1)
	}

	parentID := args[0]
	total, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		fmt.Printf("Error: invalid total value: %v\n", err)
		os.Exit(1)
	}

	weight := 0.0
	if len(args) > 2 {
		weight, err = strconv.ParseFloat(args[2], 64)
		if err != nil {
			fmt.Printf("Error: invalid weight value: %v\n", err)
			os.Exit(1)
		}
	}

	resp, err := client.AddSubTask(parentID, total, weight)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Subtask created: ID=%s, Status=%s\n", resp.ID, resp.Status)
}

func printProgressStatus(status *common.ProgressStatus) {
	fmt.Printf("ID:          %s\n", status.ID)
	fmt.Printf("Current:     %d\n", status.Current)
	fmt.Printf("Total:       %d\n", status.Total)
	fmt.Printf("Percentage:  %.1f%%\n", status.Percentage)
	
	if status.Remaining < 0 {
		fmt.Printf("Remaining:   N/A (not started yet)\n")
	} else if status.Remaining == 0 {
		fmt.Printf("Remaining:   0 (completed)\n")
	} else {
		fmt.Printf("Remaining:   %s\n", formatDuration(status.Remaining))
	}
	
	if status.Description != "" {
		fmt.Printf("Description: %s\n", status.Description)
	}
	fmt.Printf("Is Done:     %v\n", status.IsDone)
	fmt.Printf("Is Cancelled:%v\n", status.IsCancelled)
	fmt.Printf("Created At:  %s\n", status.CreatedAt.Format(time.RFC3339))
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
