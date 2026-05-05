package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"go-log-cleaner/common"
)

func main() {
	host := flag.String("host", common.DefaultServerHost, "Server host")
	port := flag.String("port", common.DefaultServerPort, "Server port")

	cmd := flag.String("cmd", "submit", "Command: submit, history, task, list, health")
	taskID := flag.String("task-id", "", "Task ID for 'task' command")

	targetDir := flag.String("dir", "", "Target directory to clean")
	mode := flag.String("mode", "time", "Clean mode: time, capacity, both")
	retainDays := flag.Int("retain-days", common.DefaultRetainDays, "Number of days to retain logs")
	capacityThreshold := flag.Float64("capacity-threshold", common.DefaultCapacityThreshold, "Capacity threshold (0.0-1.0)")
	safeWaterLevel := flag.Float64("safe-water-level", common.DefaultSafeWaterLevel, "Safe water level (0.0-1.0)")
	extensions := flag.String("extensions", common.DefaultFileExtension, "File extensions to clean (comma-separated)")
	recursive := flag.Bool("recursive", false, "Recurse into subdirectories")
	dryRun := flag.Bool("dry-run", false, "Only show what would be deleted")
	confirm := flag.Bool("confirm", false, "Actually perform deletion")
	maxDelete := flag.Int("max-delete", common.DefaultMaxDelete, "Maximum number of files to delete (0 for unlimited)")

	flag.Parse()

	client := NewClient(*host, *port)

	switch *cmd {
	case "submit":
		handleSubmit(client, targetDir, mode, retainDays, capacityThreshold, safeWaterLevel, extensions, recursive, dryRun, confirm, maxDelete)
	case "history":
		handleHistory(client)
	case "task":
		handleTask(client, taskID)
	case "list":
		handleList(client)
	case "health":
		handleHealth(client)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", *cmd)
		os.Exit(1)
	}
}

func handleSubmit(client *Client, targetDir *string, mode *string, retainDays *int,
	capacityThreshold *float64, safeWaterLevel *float64, extensions *string,
	recursive *bool, dryRun *bool, confirm *bool, maxDelete *int) {

	if *targetDir == "" {
		fmt.Fprintln(os.Stderr, "Error: target directory is required")
		os.Exit(1)
	}

	if info, err := os.Stat(*targetDir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: target directory '%s' does not exist or is not a directory\n", *targetDir)
		os.Exit(1)
	}

	cleanMode := parseMode(*mode)
	fileExtensions := parseExtensions(*extensions)

	req := common.CleanTaskRequest{
		TargetDir:         *targetDir,
		Mode:              cleanMode,
		RetainDays:        *retainDays,
		CapacityThreshold: *capacityThreshold,
		SafeWaterLevel:    *safeWaterLevel,
		FileExtensions:    fileExtensions,
		Recursive:         *recursive,
		DryRun:            *dryRun,
		Confirm:           *confirm,
		MaxDelete:         *maxDelete,
	}

	resp, err := client.SubmitTask(&req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error submitting task: %v\n", err)
		os.Exit(1)
	}

	if resp.Status == common.ResponseStatusError {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", resp.Message)
		os.Exit(1)
	}

	payloadBytes, _ := json.Marshal(resp.Payload)
	var submitResp common.SubmitTaskResponse
	json.Unmarshal(payloadBytes, &submitResp)

	fmt.Printf("Task submitted successfully\n")
	fmt.Printf("Task ID: %s\n", submitResp.TaskID)
	fmt.Printf("Use 'log-cleaner-client -cmd task -task-id %s' to check status\n", submitResp.TaskID)
}

func handleHistory(client *Client) {
	resp, err := client.GetHistory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting history: %v\n", err)
		os.Exit(1)
	}

	if resp.Status == common.ResponseStatusError {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", resp.Message)
		os.Exit(1)
	}

	payloadBytes, _ := json.Marshal(resp.Payload)
	var historyResp common.GetHistoryResponse
	json.Unmarshal(payloadBytes, &historyResp)

	if len(historyResp.Records) == 0 {
		fmt.Println("No history records found")
		return
	}

	fmt.Println("Cleaning History:")
	fmt.Println("==================")
	for _, record := range historyResp.Records {
		fmt.Printf("\nTask ID: %s\n", record.TaskID)
		fmt.Printf("  Target Dir: %s\n", record.TargetDir)
		fmt.Printf("  Mode: %s\n", modeToString(record.Mode))
		fmt.Printf("  Status: %s\n", record.Status)
		fmt.Printf("  Files Deleted: %d\n", record.FilesDeleted)
		fmt.Printf("  Size Deleted: %d bytes\n", record.SizeDeleted)
		fmt.Printf("  Exec Time: %s\n", record.ExecTime)
	}
}

func handleTask(client *Client, taskID *string) {
	if *taskID == "" {
		fmt.Fprintln(os.Stderr, "Error: task ID is required")
		os.Exit(1)
	}

	resp, err := client.GetTask(*taskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting task: %v\n", err)
		os.Exit(1)
	}

	if resp.Status == common.ResponseStatusError {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", resp.Message)
		os.Exit(1)
	}

	payloadBytes, _ := json.Marshal(resp.Payload)
	var taskResp common.GetTaskResponse
	json.Unmarshal(payloadBytes, &taskResp)

	result := taskResp.Result

	fmt.Printf("Task Status: %s\n", result.Status)
	fmt.Printf("Task ID: %s\n", result.TaskID)
	fmt.Printf("Start Time: %s\n", result.StartTime)
	if result.EndTime.After(result.StartTime) {
		fmt.Printf("End Time: %s\n", result.EndTime)
	}

	if result.Error != "" {
		fmt.Printf("Error: %s\n", result.Error)
	}

	fmt.Printf("\nFiles to Delete: %d (Total: %d bytes)\n", len(result.FilesToDelete), result.TotalSizeToDelete)
	for _, file := range result.FilesToDelete {
		fmt.Printf("  - %s (%d bytes)\n", file.Path, file.DiskUsage)
	}

	fmt.Printf("\nFiles Deleted: %d (Total: %d bytes)\n", len(result.FilesDeleted), result.TotalSizeDeleted)
	for _, file := range result.FilesDeleted {
		fmt.Printf("  - %s (%d bytes)\n", file.Path, file.DiskUsage)
	}

	if len(result.FilesSkipped) > 0 {
		fmt.Printf("\nFiles Skipped: %d\n", len(result.FilesSkipped))
		for _, file := range result.FilesSkipped {
			reason := "in use"
			fmt.Printf("  - %s (%s)\n", file.Path, reason)
		}
	}
}

func handleList(client *Client) {
	resp, err := client.ListTasks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing tasks: %v\n", err)
		os.Exit(1)
	}

	if resp.Status == common.ResponseStatusError {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", resp.Message)
		os.Exit(1)
	}

	payloadBytes, _ := json.Marshal(resp.Payload)
	var listResp common.ListTasksResponse
	json.Unmarshal(payloadBytes, &listResp)

	if len(listResp.Tasks) == 0 {
		fmt.Println("No tasks found")
		return
	}

	fmt.Println("Tasks:")
	fmt.Println("======")
	for _, task := range listResp.Tasks {
		fmt.Printf("\nTask ID: %s\n", task.TaskID)
		fmt.Printf("  Status: %s\n", task.Status)
		fmt.Printf("  Files to Delete: %d\n", len(task.FilesToDelete))
		fmt.Printf("  Files Deleted: %d\n", len(task.FilesDeleted))
	}
}

func handleHealth(client *Client) {
	resp, err := client.HealthCheck()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking health: %v\n", err)
		os.Exit(1)
	}

	if resp.Status == common.ResponseStatusError {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Printf("Server Status: %s\n", resp.Status)
	if resp.Message != "" {
		fmt.Printf("Message: %s\n", resp.Message)
	}
}

func parseMode(mode string) common.CleanMode {
	switch mode {
	case "time":
		return common.CleanModeTime
	case "capacity":
		return common.CleanModeCapacity
	case "both":
		return common.CleanModeTime | common.CleanModeCapacity
	default:
		return common.CleanModeTime
	}
}

func parseExtensions(extensions string) []string {
	if extensions == "" {
		return []string{}
	}
	parts := strings.Split(extensions, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			if !strings.HasPrefix(part, ".") {
				part = "." + part
			}
			result = append(result, part)
		}
	}
	return result
}

func modeToString(mode common.CleanMode) string {
	switch {
	case mode&common.CleanModeTime != 0 && mode&common.CleanModeCapacity != 0:
		return "time+capacity"
	case mode&common.CleanModeTime != 0:
		return "time"
	case mode&common.CleanModeCapacity != 0:
		return "capacity"
	default:
		return "unknown"
	}
}
