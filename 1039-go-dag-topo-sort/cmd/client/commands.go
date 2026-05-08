package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"dag-topo-sort/pkg/api"
)

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func httpPost(url string, body interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func handleAdd(config Config) {
	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: client add <id> <name> [deps...]")
		os.Exit(1)
	}

	taskID := args[0]
	taskName := args[1]
	dependencies := args[2:]

	req := api.AddTaskRequest{
		Task: api.Task{
			ID:           taskID,
			Name:         taskName,
			Dependencies: dependencies,
		},
	}

	body, err := httpPost(config.ServerURL+"/tasks/add", req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var resp api.AddTaskResponse
	mustUnmarshalJSON(body, &resp)

	if resp.Success {
		fmt.Printf("Task '%s' added successfully\n", taskID)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleBatchAdd(config Config) {
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: client batch-add <json-file>")
		os.Exit(1)
	}

	filename := args[0]
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	var req api.BatchAddTasksRequest
	mustUnmarshalJSON(data, &req)

	body, err := httpPost(config.ServerURL+"/tasks/batch-add", req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var resp api.BatchAddTasksResponse
	mustUnmarshalJSON(body, &resp)

	if resp.Success {
		fmt.Printf("Batch add completed: %d tasks added\n", resp.Added)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleRemove(config Config) {
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: client remove <id>")
		os.Exit(1)
	}

	taskID := args[0]
	req := api.RemoveTaskRequest{TaskID: taskID}

	body, err := httpPost(config.ServerURL+"/tasks/remove", req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var resp api.RemoveTaskResponse
	mustUnmarshalJSON(body, &resp)

	if resp.Success {
		fmt.Printf("Task '%s' removed successfully\n", taskID)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleUpdateDeps(config Config) {
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: client update-deps <id> [deps...]")
		os.Exit(1)
	}

	taskID := args[0]
	dependencies := args[1:]

	req := api.UpdateDependenciesRequest{
		TaskID:       taskID,
		Dependencies: dependencies,
	}

	body, err := httpPost(config.ServerURL+"/tasks/update-deps", req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var resp api.UpdateDependenciesResponse
	mustUnmarshalJSON(body, &resp)

	if resp.Success {
		fmt.Printf("Dependencies for task '%s' updated successfully\n", taskID)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleList(config Config) {
	body, err := httpGet(config.ServerURL + "/tasks")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var resp api.ListTasksResponse
	mustUnmarshalJSON(body, &resp)

	if !resp.Success {
		fmt.Fprintln(os.Stderr, "Error listing tasks")
		os.Exit(1)
	}

	if len(resp.Tasks) == 0 {
		fmt.Println("No tasks found")
		return
	}

	fmt.Printf("Total tasks: %d\n\n", len(resp.Tasks))
	for _, task := range resp.Tasks {
		fmt.Printf("ID: %s\n", task.ID)
		fmt.Printf("  Name: %s\n", task.Name)
		if len(task.Dependencies) > 0 {
			fmt.Printf("  Dependencies: %s\n", strings.Join(task.Dependencies, ", "))
		} else {
			fmt.Println("  Dependencies: (none)")
		}
		fmt.Println()
	}
}

func handleClear(config Config) {
	body, err := httpPost(config.ServerURL+"/tasks/clear", struct{}{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var resp api.ClearTasksResponse
	mustUnmarshalJSON(body, &resp)

	if resp.Success {
		fmt.Println("All tasks cleared")
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleSort(config Config) {
	body, err := httpGet(config.ServerURL + "/sort")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var resp api.SortResponse
	mustUnmarshalJSON(body, &resp)

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "Sort failed: %s\n", resp.Error)
		if resp.ErrorType == "cyclic" && len(resp.CyclePath) > 0 {
			fmt.Fprintf(os.Stderr, "Cycle path: %s\n", strings.Join(resp.CyclePath, " -> "))
		}
		os.Exit(1)
	}

	fmt.Println("Topological Sort Result:")
	fmt.Println("========================")
	fmt.Printf("\nExecution Order:\n")
	for i, taskID := range resp.Order {
		fmt.Printf("  %d. %s\n", i+1, taskID)
	}

	fmt.Printf("\nMaximum Parallelism: %d\n", resp.MaxParallelism)
	fmt.Printf("\nCritical Path: %s\n", strings.Join(resp.CriticalPath, " -> "))
	fmt.Printf("Minimum Completion Time: %d units\n", resp.MinCompletionTime)
}

func handleDot(config Config) {
	body, err := httpGet(config.ServerURL + "/dot")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var resp api.DotResponse
	mustUnmarshalJSON(body, &resp)

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Println(resp.Dot)
}

func handleDotCritical(config Config) {
	body, err := httpGet(config.ServerURL + "/dot-critical")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var resp api.DotResponse
	mustUnmarshalJSON(body, &resp)

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Println(resp.Dot)
}

func handleLoad(config Config) {
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: client load <json-file>")
		os.Exit(1)
	}

	filename := args[0]
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	var req api.BatchAddTasksRequest
	mustUnmarshalJSON(data, &req)

	_, err = httpPost(config.ServerURL+"/tasks/clear", struct{}{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error clearing tasks: %v\n", err)
		os.Exit(1)
	}

	body, err := httpPost(config.ServerURL+"/tasks/batch-add", req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding tasks: %v\n", err)
		os.Exit(1)
	}

	var addResp api.BatchAddTasksResponse
	mustUnmarshalJSON(body, &addResp)
	if !addResp.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", addResp.Message)
		os.Exit(1)
	}

	body, err = httpGet(config.ServerURL + "/sort")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var sortResp api.SortResponse
	mustUnmarshalJSON(body, &sortResp)

	if !sortResp.Success {
		fmt.Fprintf(os.Stderr, "Sort failed: %s\n", sortResp.Error)
		if sortResp.ErrorType == "cyclic" && len(sortResp.CyclePath) > 0 {
			fmt.Fprintf(os.Stderr, "Cycle path: %s\n", strings.Join(sortResp.CyclePath, " -> "))
		}
		os.Exit(1)
	}

	fmt.Println("Topological Sort Result:")
	fmt.Println("========================")
	fmt.Printf("\nExecution Order:\n")
	for i, taskID := range sortResp.Order {
		fmt.Printf("  %d. %s\n", i+1, taskID)
	}

	fmt.Printf("\nMaximum Parallelism: %d\n", sortResp.MaxParallelism)
	fmt.Printf("\nCritical Path: %s\n", strings.Join(sortResp.CriticalPath, " -> "))
	fmt.Printf("Minimum Completion Time: %d units\n", sortResp.MinCompletionTime)
}
