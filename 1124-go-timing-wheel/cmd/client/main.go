package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/timingwheel/pkg/api"
)

const defaultServerURL = "http://localhost:8080"

func getServerURL() string {
	if url := os.Getenv("TW_SERVER_URL"); url != "" {
		return url
	}
	return defaultServerURL
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "add":
		handleAdd(args)
	case "cancel":
		handleCancel(args)
	case "reset":
		handleReset(args)
	case "list":
		handleList(args)
	case "stats":
		handleStats(args)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("TimingWheel Client - A command line tool for the timing wheel timer server")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  add    <id> <delay> <callback>  Add a new timed task")
	fmt.Println("  cancel <id>                     Cancel a pending task")
	fmt.Println("  reset  <id> <new_delay>         Reset the delay of a task")
	fmt.Println("  list                            List all pending tasks")
	fmt.Println("  stats                           Show timing wheel statistics")
	fmt.Println("  help                            Show this help message")
	fmt.Println()
	fmt.Println("Time format for delay: 30s, 2m, 1h, 1h30m, etc.")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  TW_SERVER_URL    Server URL (default: http://localhost:8080)")
}

func handleAdd(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: client add <id> <delay> <callback>")
		os.Exit(1)
	}

	id := args[0]
	delayStr := args[1]
	callback := args[2]

	delay, err := time.ParseDuration(delayStr)
	if err != nil {
		fmt.Printf("Invalid delay format: %s\n", delayStr)
		os.Exit(1)
	}

	req := &api.AddTaskRequest{
		ID:       id,
		Callback: callback,
		Delay:    delay,
	}

	resp, err := httpPostJSON(getServerURL()+"/tasks/add", req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.AddTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error parsing response: %s\n", err)
		os.Exit(1)
	}

	if result.Success {
		fmt.Printf("Task '%s' added successfully (delay: %s, callback: %s)\n", id, delay, callback)
	} else {
		fmt.Printf("Error adding task: %s\n", result.Error)
		os.Exit(1)
	}
}

func handleCancel(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client cancel <id>")
		os.Exit(1)
	}

	id := args[0]

	req := &api.CancelTaskRequest{ID: id}

	resp, err := httpPostJSON(getServerURL()+"/tasks/cancel", req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.CancelTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error parsing response: %s\n", err)
		os.Exit(1)
	}

	if result.Success {
		fmt.Printf("Task '%s' cancelled successfully\n", id)
	} else {
		fmt.Printf("Error cancelling task: %s\n", result.Error)
		os.Exit(1)
	}
}

func handleReset(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: client reset <id> <new_delay>")
		os.Exit(1)
	}

	id := args[0]
	newDelayStr := args[1]

	newDelay, err := time.ParseDuration(newDelayStr)
	if err != nil {
		fmt.Printf("Invalid delay format: %s\n", newDelayStr)
		os.Exit(1)
	}

	req := &api.ResetTaskRequest{
		ID:       id,
		NewDelay: newDelay,
	}

	resp, err := httpPostJSON(getServerURL()+"/tasks/reset", req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.ResetTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error parsing response: %s\n", err)
		os.Exit(1)
	}

	if result.Success {
		fmt.Printf("Task '%s' reset successfully (new delay: %s)\n", id, newDelay)
	} else {
		fmt.Printf("Error resetting task: %s\n", result.Error)
		os.Exit(1)
	}
}

func handleList(args []string) {
	resp, err := http.Get(getServerURL() + "/tasks/list")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.ListTasksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error parsing response: %s\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("Error listing tasks: %s\n", result.Error)
		os.Exit(1)
	}

	if len(result.Tasks) == 0 {
		fmt.Println("No pending tasks")
		return
	}

	fmt.Printf("Pending tasks (%d):\n", len(result.Tasks))
	fmt.Println("----------------------------------------------------------------------")
	for i, t := range result.Tasks {
		fmt.Printf("[%d] ID: %s\n", i+1, t.ID)
		fmt.Printf("     Callback: %s\n", t.Callback)
		fmt.Printf("     Original delay: %s\n", t.Delay)
		fmt.Printf("     Remaining: %s\n", t.Remaining)
		fmt.Printf("     Status: %s\n", t.Status)
		fmt.Println("----------------------------------------------------------------------")
	}
}

func handleStats(args []string) {
	resp, err := http.Get(getServerURL() + "/stats")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error parsing response: %s\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("Error getting stats: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Println("Timing Wheel Statistics")
	fmt.Println("=======================")
	fmt.Printf("Total pending tasks: %d\n", result.TotalPending)
	fmt.Printf("Executed tasks: %d\n", result.ExecutedCount)
	fmt.Printf("Cancelled tasks: %d\n\n", result.CancelledCount)

	fmt.Println("Wheel distribution:")
	fmt.Printf("  Hour wheel   : %d tasks (current tick: %d/%d)\n", result.HourTasks, result.HourTick, 24)
	fmt.Printf("  Minute wheel : %d tasks (current tick: %d/%d)\n", result.MinuteTasks, result.MinuteTick, 60)
	fmt.Printf("  Second wheel : %d tasks (current tick: %d/%d)\n", result.SecondTasks, result.SecondTick, 60)
}

func httpPostJSON(url string, data interface{}) (*http.Response, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

func formatDuration(d time.Duration) string {
	return strconv.FormatFloat(d.Seconds(), 'f', -1, 64) + "s"
}
