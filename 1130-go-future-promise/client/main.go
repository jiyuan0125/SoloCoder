package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"future-promise/api"
	"io"
	"net/http"
	"os"
	"time"
)

const ServerURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "run":
		handleRun(args)
	case "chain":
		handleChain(args)
	case "timeout":
		handleTimeout(args)
	case "cancel":
		handleCancel(args)
	case "log":
		handleLog(args)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Future/Promise Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  run       Submit a simple task and wait for result")
	fmt.Println("  chain     Submit a chained task with multiple steps")
	fmt.Println("  timeout   Test timeout behavior")
	fmt.Println("  cancel    Cancel a running task")
	fmt.Println("  log       View all task execution logs")
	fmt.Println("  help      Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client run --input \"hello\"")
	fmt.Println("  client chain --steps 3")
	fmt.Println("  client timeout --duration 1s")
	fmt.Println("  client cancel --task-id 1")
	fmt.Println("  client log")
}

func handleRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	input := fs.String("input", "", "Input data for the task")
	wait := fs.Bool("wait", true, "Wait for task completion")
	fs.Parse(args)

	req := api.SubmitTaskRequest{
		Type:  "simple",
		Input: *input,
	}

	taskID, err := submitTask(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task submitted. ID: %s\n", taskID)

	if *wait {
		waitForTask(taskID)
	}
}

func handleChain(args []string) {
	fs := flag.NewFlagSet("chain", flag.ExitOnError)
	stepsCount := fs.Int("steps", 3, "Number of steps in chain")
	delay := fs.Duration("delay", 500*time.Millisecond, "Delay between steps")
	failAt := fs.Int("fail-at", 0, "Step number to fail at (0 for no failure)")
	wait := fs.Bool("wait", true, "Wait for task completion")
	fs.Parse(args)

	steps := make([]api.TaskStep, *stepsCount)
	for i := 0; i < *stepsCount; i++ {
		steps[i] = api.TaskStep{
			Name:  fmt.Sprintf("step-%d", i+1),
			Input: map[string]interface{}{"step": i + 1},
			Delay: *delay,
			Fail:  *failAt == i+1,
		}
	}

	req := api.SubmitTaskRequest{
		Type:  "chain",
		Steps: steps,
	}

	taskID, err := submitTask(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Chain task submitted. ID: %s, Steps: %d\n", taskID, *stepsCount)
	if *failAt > 0 {
		fmt.Printf("Will fail at step: %d\n", *failAt)
	}

	if *wait {
		waitForTask(taskID)
	}
}

func handleTimeout(args []string) {
	fs := flag.NewFlagSet("timeout", flag.ExitOnError)
	duration := fs.Duration("duration", 1*time.Second, "Timeout duration")
	wait := fs.Bool("wait", true, "Wait for task completion")
	fs.Parse(args)

	req := api.SubmitTaskRequest{
		Type:    "simple",
		Input:   "timeout-test",
		Timeout: *duration,
	}

	taskID, err := submitTask(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Timeout test submitted. ID: %s, Timeout: %v\n", taskID, *duration)
	fmt.Println("Note: Task takes 2 seconds to complete, should timeout.")

	if *wait {
		waitForTask(taskID)
	}
}

func handleCancel(args []string) {
	fs := flag.NewFlagSet("cancel", flag.ExitOnError)
	taskID := fs.String("task-id", "", "Task ID to cancel")
	fs.Parse(args)

	if *taskID == "" {
		fmt.Println("Error: --task-id is required")
		fs.Usage()
		os.Exit(1)
	}

	err := cancelTask(*taskID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Cancel request sent for task: %s\n", *taskID)
}

func handleLog(args []string) {
	logs, err := getLogs()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Task Execution Logs ===")
	fmt.Println()

	if len(logs) == 0 {
		fmt.Println("No logs found.")
		return
	}

	for i, entry := range logs {
		fmt.Printf("[%d] %s\n", i+1, entry.Timestamp.Format(time.RFC3339Nano))
		fmt.Printf("    Task: %s\n", entry.TaskID)
		fmt.Printf("    Action: %s\n", entry.Action)
		if entry.Details != "" {
			fmt.Printf("    Details: %s\n", entry.Details)
		}
		fmt.Println()
	}
}

func submitTask(req api.SubmitTaskRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(ServerURL+"/submit", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	var result api.SubmitTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.TaskID, nil
}

func getTask(taskID string) (*api.GetTaskResponse, error) {
	resp, err := http.Get(ServerURL + "/task/" + taskID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	var result api.GetTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func cancelTask(taskID string) error {
	req, err := http.NewRequest("POST", ServerURL+"/cancel/"+taskID, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	var result api.CancelTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf(result.Message)
	}

	return nil
}

func getLogs() ([]api.TaskLogEntry, error) {
	resp, err := http.Get(ServerURL + "/logs")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	var result api.GetLogsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Logs, nil
}

func waitForTask(taskID string) {
	fmt.Println("Waiting for task to complete...")

	for {
		task, err := getTask(taskID)
		if err != nil {
			fmt.Printf("Error polling task: %v\n", err)
			return
		}

		fmt.Printf("Status: %s\n", task.State)

		switch task.State {
		case api.StateCompleted:
			fmt.Println("\n=== Task Completed Successfully ===")
			if task.Result != nil {
				resultJSON, _ := json.MarshalIndent(task.Result, "", "  ")
				fmt.Printf("Result:\n%s\n", string(resultJSON))
			}
			return
		case api.StateFailed, api.StateCancelled, api.StateTimeout:
			fmt.Printf("\n=== Task %s ===\n", task.State)
			if task.Error != "" {
				fmt.Printf("Error: %s\n", task.Error)
			}
			return
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}
}
