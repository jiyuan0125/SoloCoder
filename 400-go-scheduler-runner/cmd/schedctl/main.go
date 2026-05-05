package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"scheduler/internal/protocol"
)

const (
	defaultServerAddr = "localhost:8081"
)

var (
	serverAddr string
)

func main() {
	flag.StringVar(&serverAddr, "server", defaultServerAddr, "Scheduler server address (host:port)")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	cmd := args[0]
	switch cmd {
	case "list":
		listTasks()
	case "get":
		if len(args) < 2 {
			fmt.Println("Error: task name required")
			os.Exit(1)
		}
		getTask(args[1])
	case "add":
		addTask(args[1:])
	case "delete":
		if len(args) < 2 {
			fmt.Println("Error: task name required")
			os.Exit(1)
		}
		deleteTask(args[1])
	case "executions":
		if len(args) >= 2 {
			listExecutions(args[1])
		} else {
			listExecutions("")
		}
	case "ping":
		ping()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`Scheduler Client - Manage scheduled tasks

Usage:
  schedctl [flags] command [arguments]

Commands:
  list                    List all tasks
  get <name>              Get details of a specific task
  add [options]           Add a new task
    Options:
      -name=<name>        Task name (required)
      -command=<cmd>      Shell command to execute (required)
      -cron=<expr>        Cron expression (5 fields: min hour day month weekday) (required)
      -timeout=<sec>      Timeout in seconds (default: 60)
      -max-retry=<n>      Max retry attempts (default: 3)
      -retry-interval=<sec>  Retry interval in seconds (default: 60)
  delete <name>           Delete a task
  executions [name]       List execution records (all or for a specific task)
  ping                    Check if server is reachable

Flags:
  -server=<addr>          Server address (default: localhost:8081)

Examples:
  schedctl list
  schedctl get cleanup-task
  schedctl add -name=cleanup -command="rm -rf /tmp/*" -cron="0 0 * * *" -timeout=300
  schedctl add -name=health-check -command="curl http://localhost/health" -cron="*/5 * * * *"
  schedctl delete old-task
  schedctl executions cleanup
  schedctl -server=192.168.1.100:8081 list`)
}

func sendRequest(req protocol.Request) (*protocol.Response, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	data = append(data, '\n')

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := conn.Write(data); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	reader := bufio.NewReader(conn)
	respData, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resp protocol.Response
	if err := json.Unmarshal([]byte(respData), &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

func ping() {
	req := protocol.Request{
		Operation: protocol.OpPing,
	}

	resp, err := sendRequest(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Println("Server is reachable")
	} else {
		fmt.Printf("Ping failed: %s\n", resp.Error)
		os.Exit(1)
	}
}

func listTasks() {
	req := protocol.Request{
		Operation: protocol.OpListTasks,
	}

	resp, err := sendRequest(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	tasksRaw, ok := resp.Data["tasks"]
	if !ok {
		fmt.Println("No tasks found")
		return
	}

	tasksData, err := json.Marshal(tasksRaw)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var tasks []protocol.TaskStatus
	if err := json.Unmarshal(tasksData, &tasks); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found")
		return
	}

	fmt.Printf("Total %d tasks:\n\n", len(tasks))
	for _, task := range tasks {
		printTaskStatus(&task)
		fmt.Println("---")
	}
}

func getTask(name string) {
	req := protocol.Request{
		Operation: protocol.OpGetTask,
		Data: map[string]interface{}{
			"name": name,
		},
	}

	resp, err := sendRequest(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	taskRaw, ok := resp.Data["task"]
	if !ok {
		fmt.Println("Task not found")
		os.Exit(1)
	}

	taskData, err := json.Marshal(taskRaw)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var task protocol.TaskStatus
	if err := json.Unmarshal(taskData, &task); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printTaskStatus(&task)
}

func printTaskStatus(task *protocol.TaskStatus) {
	fmt.Printf("Name:        %s\n", task.Name)
	fmt.Printf("Command:     %s\n", task.Command)
	fmt.Printf("Cron:        %s\n", task.CronExpr)
	fmt.Printf("Timeout:     %s\n", task.Timeout)
	fmt.Printf("Max Retry:   %d\n", task.MaxRetry)
	fmt.Printf("Retry Intv:  %s\n", task.RetryInterval)
	fmt.Printf("Disabled:    %v\n", task.Disabled)
	fmt.Printf("Running:     %v\n", task.IsRunning)
	if !task.NextRun.IsZero() {
		fmt.Printf("Next Run:    %s\n", task.NextRun.Format("2006-01-02 15:04:05"))
	}
	if task.LastExecution != nil {
		last := task.LastExecution
		fmt.Printf("Last Exec:   %s (Success: %v, Retries: %d)\n",
			last.StartTime.Format("2006-01-02 15:04:05"),
			last.Success,
			last.RetryCount)
		fmt.Printf("  Duration:  %s\n", last.Duration)
		if !last.Success && last.Error != "" {
			fmt.Printf("  Error:     %s\n", strings.ReplaceAll(last.Error, "\n", " "))
		}
	}
}

func addTask(args []string) {
	params := parseArgs(args)

	name := params["name"]
	command := params["command"]
	cronExpr := params["cron"]

	if name == "" {
		fmt.Println("Error: -name is required")
		os.Exit(1)
	}
	if command == "" {
		fmt.Println("Error: -command is required")
		os.Exit(1)
	}
	if cronExpr == "" {
		fmt.Println("Error: -cron is required")
		os.Exit(1)
	}

	reqData := map[string]interface{}{
		"name":      name,
		"command":   command,
		"cron_expr": cronExpr,
	}

	if timeoutStr := params["timeout"]; timeoutStr != "" {
		if timeout, err := strconv.ParseFloat(timeoutStr, 64); err == nil {
			reqData["timeout"] = timeout
		}
	}
	if maxRetryStr := params["max-retry"]; maxRetryStr != "" {
		if maxRetry, err := strconv.ParseFloat(maxRetryStr, 64); err == nil {
			reqData["max_retry"] = maxRetry
		}
	}
	if retryIntvStr := params["retry-interval"]; retryIntvStr != "" {
		if interval, err := strconv.ParseFloat(retryIntvStr, 64); err == nil {
			reqData["retry_interval"] = interval
		}
	}

	req := protocol.Request{
		Operation: protocol.OpAddTask,
		Data:      reqData,
	}

	resp, err := sendRequest(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Printf("Task '%s' added successfully\n", name)
}

func deleteTask(name string) {
	req := protocol.Request{
		Operation: protocol.OpDeleteTask,
		Data: map[string]interface{}{
			"name": name,
		},
	}

	resp, err := sendRequest(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Printf("Task '%s' deleted successfully\n", name)
}

func listExecutions(taskName string) {
	req := protocol.Request{
		Operation: protocol.OpListExecutions,
		Data:      make(map[string]interface{}),
	}
	if taskName != "" {
		req.Data["task_name"] = taskName
	}

	resp, err := sendRequest(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	execRaw, ok := resp.Data["executions"]
	if !ok {
		fmt.Println("No execution records found")
		return
	}

	execData, err := json.Marshal(execRaw)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var executions []protocol.ExecutionRecord
	if err := json.Unmarshal(execData, &executions); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(executions) == 0 {
		fmt.Println("No execution records found")
		return
	}

	fmt.Printf("Total %d execution records:\n\n", len(executions))
	for _, exec := range executions {
		printExecution(&exec)
		fmt.Println("---")
	}
}

func printExecution(exec *protocol.ExecutionRecord) {
	status := "SUCCESS"
	if !exec.Success {
		status = "FAILED"
	}
	fmt.Printf("ID:          %s\n", exec.ID)
	fmt.Printf("Task:        %s\n", exec.TaskName)
	fmt.Printf("Start:       %s\n", exec.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Duration:    %s\n", exec.Duration)
	fmt.Printf("Status:      %s\n", status)
	fmt.Printf("Retries:     %d\n", exec.RetryCount)
	if exec.Output != "" {
		fmt.Printf("Output:      %s\n", strings.ReplaceAll(exec.Output, "\n", "\\n"))
	}
	if !exec.Success && exec.Error != "" {
		fmt.Printf("Error:       %s\n", strings.ReplaceAll(exec.Error, "\n", "\\n"))
	}
}

func parseArgs(args []string) map[string]string {
	result := make(map[string]string)
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			parts := strings.SplitN(arg[1:], "=", 2)
			if len(parts) == 2 {
				result[parts[0]] = parts[1]
			} else if len(parts) == 1 {
				result[parts[0]] = "true"
			}
		}
	}
	return result
}
