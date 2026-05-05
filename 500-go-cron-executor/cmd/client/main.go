package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"cron-executor/pkg/common"
)

type Client struct {
	host string
	port int
}

func NewClient(host string, port int) *Client {
	return &Client{
		host: host,
		port: port,
	}
}

func (c *Client) sendMessage(msg *common.Message) (*common.Message, error) {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(conn)
	writer.Write(msgBytes)
	writer.WriteByte('\n')
	if err := writer.Flush(); err != nil {
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	var resp common.Message
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) Ping() error {
	msg := &common.Message{Type: common.MsgTypePing}
	resp, err := c.sendMessage(msg)
	if err != nil {
		return err
	}

	if resp.Type != common.MsgTypePong {
		return fmt.Errorf("unexpected response: %s", resp.Type)
	}

	return nil
}

func (c *Client) ListTasks() (*common.TaskListResponse, error) {
	msg := &common.Message{Type: common.MsgTypeListTasks}
	resp, err := c.sendMessage(msg)
	if err != nil {
		return nil, err
	}

	if resp.Type == common.MsgTypeError {
		var errResp common.ErrorResponse
		json.Unmarshal(resp.Payload, &errResp)
		return nil, fmt.Errorf(errResp.Message)
	}

	var taskList common.TaskListResponse
	if err := json.Unmarshal(resp.Payload, &taskList); err != nil {
		return nil, err
	}

	return &taskList, nil
}

func (c *Client) TriggerTask(taskName string) (*common.TriggerResultResponse, error) {
	req := common.TriggerTaskRequest{TaskName: taskName}
	payload, _ := json.Marshal(req)

	msg := &common.Message{
		Type:    common.MsgTypeTriggerTask,
		Payload: payload,
	}

	resp, err := c.sendMessage(msg)
	if err != nil {
		return nil, err
	}

	if resp.Type == common.MsgTypeError {
		var errResp common.ErrorResponse
		json.Unmarshal(resp.Payload, &errResp)
		return nil, fmt.Errorf(errResp.Message)
	}

	var result common.TriggerResultResponse
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetStatus(taskName string) (*common.TaskStatusResponse, error) {
	req := common.GetStatusRequest{TaskName: taskName}
	payload, _ := json.Marshal(req)

	msg := &common.Message{
		Type:    common.MsgTypeGetStatus,
		Payload: payload,
	}

	resp, err := c.sendMessage(msg)
	if err != nil {
		return nil, err
	}

	if resp.Type == common.MsgTypeError {
		var errResp common.ErrorResponse
		json.Unmarshal(resp.Payload, &errResp)
		return nil, fmt.Errorf(errResp.Message)
	}

	var status common.TaskStatusResponse
	if err := json.Unmarshal(resp.Payload, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (c *Client) GetExecutions(taskName string, limit int) (*common.ExecutionListResponse, error) {
	req := common.GetExecutionsRequest{
		TaskName: taskName,
		Limit:    limit,
	}
	payload, _ := json.Marshal(req)

	msg := &common.Message{
		Type:    common.MsgTypeGetExecutions,
		Payload: payload,
	}

	resp, err := c.sendMessage(msg)
	if err != nil {
		return nil, err
	}

	if resp.Type == common.MsgTypeError {
		var errResp common.ErrorResponse
		json.Unmarshal(resp.Payload, &errResp)
		return nil, fmt.Errorf(errResp.Message)
	}

	var list common.ExecutionListResponse
	if err := json.Unmarshal(resp.Payload, &list); err != nil {
		return nil, err
	}

	return &list, nil
}

func (c *Client) GetStats() (*common.StatsResponse, error) {
	msg := &common.Message{Type: common.MsgTypeGetStats}
	resp, err := c.sendMessage(msg)
	if err != nil {
		return nil, err
	}

	if resp.Type == common.MsgTypeError {
		var errResp common.ErrorResponse
		json.Unmarshal(resp.Payload, &errResp)
		return nil, fmt.Errorf(errResp.Message)
	}

	var stats common.StatsResponse
	if err := json.Unmarshal(resp.Payload, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

func main() {
	host := flag.String("host", "localhost", "Server host")
	port := flag.Int("port", 8765, "Server TCP port")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(*host, *port)

	command := strings.ToLower(args[0])
	switch command {
	case "ping":
		if err := client.Ping(); err != nil {
			fmt.Printf("Ping failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Pong")

	case "list", "ls":
		tasks, err := client.ListTasks()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printTasks(tasks)

	case "trigger", "run":
		if len(args) < 2 {
			fmt.Println("Usage: cron-client trigger <task-name>")
			os.Exit(1)
		}
		result, err := client.TriggerTask(args[1])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printTriggerResult(result)

	case "status":
		if len(args) < 2 {
			fmt.Println("Usage: cron-client status <task-name>")
			os.Exit(1)
		}
		status, err := client.GetStatus(args[1])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printTaskStatus(status)

	case "executions", "execs":
		taskName := ""
		limit := 20
		if len(args) >= 2 {
			taskName = args[1]
		}
		execs, err := client.GetExecutions(taskName, limit)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printExecutions(execs)

	case "stats":
		stats, err := client.GetStats()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printStats(stats)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Cron Executor Client")
	fmt.Println()
	fmt.Println("Usage: cron-client [flags] <command> [args]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -host string   Server host (default \"localhost\")")
	fmt.Println("  -port int      Server TCP port (default 8765)")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  ping                     Test server connection")
	fmt.Println("  list, ls                 List all tasks")
	fmt.Println("  trigger, run <name>      Manually trigger a task")
	fmt.Println("  status <name>            Get task status")
	fmt.Println("  executions, execs [name] List recent executions")
	fmt.Println("  stats                    Get execution statistics")
	fmt.Println()
}

func printTasks(tasks *common.TaskListResponse) {
	if len(tasks.Tasks) == 0 {
		fmt.Println("No tasks configured")
		return
	}

	fmt.Printf("%-20s %-15s %-12s %s\n", "NAME", "STATUS", "NEXT RUN", "CRON")
	fmt.Println(strings.Repeat("-", 80))

	for _, t := range tasks.Tasks {
		nextRun := "-"
		if t.NextRun != "" {
			nextRun = t.NextRun[11:19]
		}
		fmt.Printf("%-20s %-15s %-12s %s\n", t.Name, t.Status, nextRun, t.CronExpr)
	}
}

func printTriggerResult(result *common.TriggerResultResponse) {
	if result.Success {
		fmt.Printf("Success: %s\n", result.Message)
		if result.ExecID != "" {
			fmt.Printf("Execution ID: %s\n", result.ExecID)
		}
	} else {
		fmt.Printf("Failed: %s\n", result.Message)
	}
}

func printTaskStatus(status *common.TaskStatusResponse) {
	fmt.Printf("Name:         %s\n", status.Name)
	fmt.Printf("Status:       %s\n", status.Status)
	fmt.Printf("Cron:         %s\n", status.CronExpr)
	fmt.Printf("Command:      %s\n", status.Command)
	fmt.Printf("Timeout:      %s\n", status.Timeout)
	fmt.Printf("Max Retries:  %d\n", status.MaxRetries)
	if status.LastRun != "" {
		fmt.Printf("Last Run:     %s\n", status.LastRun)
	}
	if status.NextRun != "" {
		fmt.Printf("Next Run:     %s\n", status.NextRun)
	}
	if status.ActiveExecID != "" {
		fmt.Printf("Active Exec:  %s\n", status.ActiveExecID)
	}
	if len(status.Dependencies) > 0 {
		fmt.Printf("Dependencies: %v\n", status.Dependencies)
	}
}

func printExecutions(execs *common.ExecutionListResponse) {
	if len(execs.Executions) == 0 {
		fmt.Println("No executions found")
		return
	}

	fmt.Printf("%-20s %-15s %-8s %-10s %s\n", "TASK", "STATUS", "RETRY", "DURATION", "START TIME")
	fmt.Println(strings.Repeat("-", 90))

	for _, e := range execs.Executions {
		duration := "-"
		if e.Duration != "" {
			duration = e.Duration
		}
		startTime := e.StartTime
		if len(startTime) > 19 {
			startTime = startTime[:19]
		}
		fmt.Printf("%-20s %-15s %-8d %-10s %s\n",
			e.TaskName, e.Status, e.RetryCount, duration, startTime)
	}
}

func printStats(stats *common.StatsResponse) {
	fmt.Println("=== Statistics ===")
	fmt.Printf("Total Tasks:      %d\n", stats.TotalTasks)
	fmt.Printf("Running Tasks:    %d\n", stats.RunningTasks)
	fmt.Printf("Total Executions: %d\n", stats.TotalExecutions)
	fmt.Printf("Success Rate:     %.2f%%\n", stats.SuccessRate)
	fmt.Printf("Avg Duration:     %s\n", stats.AvgDuration)

	if len(stats.RecentFailures) > 0 {
		fmt.Println()
		fmt.Println("=== Recent Failures ===")
		fmt.Printf("%-20s %-15s %s\n", "TASK", "STATUS", "START TIME")
		fmt.Println(strings.Repeat("-", 60))
		for _, f := range stats.RecentFailures {
			startTime := f.StartTime
			if len(startTime) > 19 {
				startTime = startTime[:19]
			}
			fmt.Printf("%-20s %-15s %s\n", f.TaskName, f.Status, startTime)
		}
	}
}
