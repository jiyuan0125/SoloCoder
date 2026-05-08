package main

import (
	"bytes"
	"dag-scheduler/common"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	return &Client{serverURL: serverURL}
}

func (c *Client) Submit(req common.SubmitRequest) (*common.SubmitResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.serverURL+"/submit", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.SubmitResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) Start(jobID string) (*common.StartResponse, error) {
	req := common.StartRequest{JobID: jobID}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.serverURL+"/start", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.StartResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) Status(jobID string) (*common.StatusResponse, error) {
	resp, err := http.Get(fmt.Sprintf("%s/status?job_id=%s", c.serverURL, jobID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.StatusResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) Cancel(jobID string) (*common.CancelResponse, error) {
	req := common.CancelRequest{JobID: jobID}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.serverURL+"/cancel", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.CancelResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func formatTime(ms int64) string {
	if ms == 0 {
		return "-"
	}
	t := time.UnixMilli(ms)
	return t.Format("15:04:05.000")
}

func formatDuration(ms int64) string {
	if ms == 0 {
		return "-"
	}
	return fmt.Sprintf("%.2fs", float64(ms)/1000.0)
}

func printReport(status *common.StatusResponse) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("EXECUTION REPORT")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Job ID: %s\n", status.JobID)
	fmt.Printf("Is Running: %v\n", status.IsRunning)
	fmt.Println()

	keys := make([]string, 0, len(status.Tasks))
	for k := range status.Tasks {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Println("Tasks:")
	fmt.Println("-")
	fmt.Printf("%-10s %-12s %-15s %-15s %-12s %-10s %s\n",
		"Task ID", "Status", "Start Time", "End Time", "Duration", "Attempts", "Dependencies")
	fmt.Println("-")

	completed := 0
	failed := 0
	skipped := 0
	running := 0
	pending := 0

	for _, id := range keys {
		task := status.Tasks[id]
		fmt.Printf("%-10s %-12s %-15s %-15s %-12s %-10d %v\n",
			id,
			task.Status,
			formatTime(task.StartTime),
			formatTime(task.EndTime),
			formatDuration(task.DurationMs),
			task.Attempts,
			task.Dependencies)

		switch task.Status {
		case "completed":
			completed++
		case "failed":
			failed++
		case "skipped":
			skipped++
		case "running":
			running++
		case "pending":
			pending++
		}
	}

	fmt.Println("-")
	fmt.Printf("\nSummary: Completed=%d, Failed=%d, Skipped=%d, Running=%d, Pending=%d\n",
		completed, failed, skipped, running, pending)

	if failed > 0 {
		fmt.Println("\nFailed Tasks Details:")
		for _, id := range keys {
			task := status.Tasks[id]
			if task.Status == "failed" && task.LastError != "" {
				fmt.Printf("  - %s: %s\n", id, task.LastError)
			}
		}
	}
}

func printDependencyGraph(status *common.StatusResponse) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("DEPENDENCY GRAPH")
	fmt.Println(strings.Repeat("=", 80))

	keys := make([]string, 0, len(status.Tasks))
	for k := range status.Tasks {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, id := range keys {
		task := status.Tasks[id]
		statusSymbol := "?"
		switch task.Status {
		case "completed":
			statusSymbol = "✓"
		case "failed":
			statusSymbol = "✗"
		case "running":
			statusSymbol = "▶"
		case "skipped":
			statusSymbol = "⊘"
		case "pending":
			statusSymbol = "◌"
		}

		if len(task.Dependencies) == 0 {
			fmt.Printf("  [%s] %s (no dependencies)\n", statusSymbol, id)
		} else {
			for i, dep := range task.Dependencies {
				if i == 0 {
					fmt.Printf("  [%s] %s\n", statusSymbol, id)
				}
				fmt.Printf("        └── depends on: %s\n", dep)
			}
		}
	}
}

func main() {
	var (
		serverURL  = flag.String("server", "http://localhost:8080", "Server URL")
		jsonFile   = flag.String("file", "", "JSON file containing task graph definition")
		jobID      = flag.String("job", "", "Job ID for status/cancel operations")
		operation  = flag.String("op", "run", "Operation: submit|start|status|cancel|run")
		poll       = flag.Bool("poll", true, "Poll for status until completion")
		pollInterval = flag.Int("interval", 1, "Poll interval in seconds")
	)

	flag.Parse()

	client := NewClient(*serverURL)

	switch *operation {
	case "submit":
		if *jsonFile == "" {
			fmt.Println("Error: -file is required for submit operation")
			os.Exit(1)
		}

		data, err := os.ReadFile(*jsonFile)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		var req common.SubmitRequest
		if err := json.Unmarshal(data, &req); err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			os.Exit(1)
		}

		resp, err := client.Submit(req)
		if err != nil {
			fmt.Printf("Error submitting job: %v\n", err)
			os.Exit(1)
		}

		if !resp.Success {
			fmt.Printf("Submit failed: %s\n", resp.Error)
			if len(resp.Cycle) > 0 {
				fmt.Printf("Cycle detected: %v\n", resp.Cycle)
			}
			os.Exit(1)
		}

		fmt.Printf("Job submitted successfully. Job ID: %s\n", resp.JobID)

	case "start":
		if *jobID == "" {
			fmt.Println("Error: -job is required for start operation")
			os.Exit(1)
		}

		resp, err := client.Start(*jobID)
		if err != nil {
			fmt.Printf("Error starting job: %v\n", err)
			os.Exit(1)
		}

		if !resp.Success {
			fmt.Printf("Start failed: %s\n", resp.Error)
			os.Exit(1)
		}

		fmt.Println("Job started successfully")

	case "status":
		if *jobID == "" {
			fmt.Println("Error: -job is required for status operation")
			os.Exit(1)
		}

		resp, err := client.Status(*jobID)
		if err != nil {
			fmt.Printf("Error getting status: %v\n", err)
			os.Exit(1)
		}

		if !resp.Success {
			fmt.Printf("Status request failed: %s\n", resp.Error)
			os.Exit(1)
		}

		printReport(resp)
		printDependencyGraph(resp)

	case "cancel":
		if *jobID == "" {
			fmt.Println("Error: -job is required for cancel operation")
			os.Exit(1)
		}

		resp, err := client.Cancel(*jobID)
		if err != nil {
			fmt.Printf("Error canceling job: %v\n", err)
			os.Exit(1)
		}

		if !resp.Success {
			fmt.Printf("Cancel failed: %s\n", resp.Error)
			os.Exit(1)
		}

		fmt.Println("Job canceled successfully")

	case "run":
		if *jsonFile == "" {
			fmt.Println("Error: -file is required for run operation")
			os.Exit(1)
		}

		data, err := os.ReadFile(*jsonFile)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		var req common.SubmitRequest
		if err := json.Unmarshal(data, &req); err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			os.Exit(1)
		}

		submitResp, err := client.Submit(req)
		if err != nil {
			fmt.Printf("Error submitting job: %v\n", err)
			os.Exit(1)
		}

		if !submitResp.Success {
			fmt.Printf("Submit failed: %s\n", submitResp.Error)
			if len(submitResp.Cycle) > 0 {
				fmt.Printf("Cycle detected: %v\n", submitResp.Cycle)
			}
			os.Exit(1)
		}

		jobID := submitResp.JobID
		fmt.Printf("Job submitted successfully. Job ID: %s\n", jobID)

		startResp, err := client.Start(jobID)
		if err != nil {
			fmt.Printf("Error starting job: %v\n", err)
			os.Exit(1)
		}

		if !startResp.Success {
			fmt.Printf("Start failed: %s\n", startResp.Error)
			os.Exit(1)
		}

		fmt.Println("Job started. Waiting for completion...")

		if *poll {
			var finalStatus *common.StatusResponse
			for {
				statusResp, err := client.Status(jobID)
				if err != nil {
					fmt.Printf("Error getting status: %v\n", err)
					os.Exit(1)
				}

				if !statusResp.Success {
					fmt.Printf("Status request failed: %s\n", statusResp.Error)
					os.Exit(1)
				}

				finalStatus = statusResp

				running := false
				for _, task := range statusResp.Tasks {
					if task.Status == "running" || task.Status == "pending" {
						running = true
						break
					}
				}

				if !running || !statusResp.IsRunning {
					break
				}

				completed := 0
				failed := 0
				skipped := 0
				runningCount := 0
				pendingCount := 0
				for _, task := range statusResp.Tasks {
					switch task.Status {
					case "completed":
						completed++
					case "failed":
						failed++
					case "skipped":
						skipped++
					case "running":
						runningCount++
					case "pending":
						pendingCount++
					}
				}

				fmt.Printf("\rStatus: Completed=%d, Failed=%d, Skipped=%d, Running=%d, Pending=%d",
					completed, failed, skipped, runningCount, pendingCount)

				time.Sleep(time.Duration(*pollInterval) * time.Second)
			}

			fmt.Println()
			printReport(finalStatus)
			printDependencyGraph(finalStatus)
		}

	default:
		fmt.Printf("Unknown operation: %s\n", *operation)
		fmt.Println("Valid operations: submit|start|status|cancel|run")
		os.Exit(1)
	}
}
