package main

import (
	"flag"
	"fmt"
	"go-report-scheduler/internal/client"
	"go-report-scheduler/pkg/common"
	"os"
	"strings"
)

var (
	hostFlag = flag.String("host", "localhost", "Server host")
	portFlag = flag.Int("port", 8080, "Server port")
)

func printUsage() {
	fmt.Println(`Report Scheduler Client

Usage:
  rs-cli [global-options] command [command-options] [arguments]

Global Options:
  -host string    Server host (default "localhost")
  -port int       Server port (default 8080)

Commands:
  create      Create a new task
  get         Get task details
  list        List all tasks
  update      Update a task
  delete      Delete a task
  pause       Pause a task
  resume      Resume a task
  trigger     Manually trigger a task
  exec-get    Get execution details
  exec-list   List all executions
  help        Show this help message

Use "rs-cli command -h" for more information about a command.`)
}

func printCreateUsage() {
	fmt.Println(`Create a new task

Usage:
  rs-cli create -name <name> -cron <cron-expression> [options]

Options:
  -name string         Task name (required)
  -cron string         Cron expression (required, 6 fields: second minute hour day month week)
  -params string       Parameters in format: key1=value1,key2=value2
  -depends string      Dependent task IDs, comma-separated

Examples:
  rs-cli create -name "销售日报" -cron "0 0 9 * * 1-5"
  rs-cli create -name "财务月报" -cron "0 0 8 1 * *" -params "start_date=2024-01-01,end_date=2024-12-31"
  rs-cli create -name "汇总报表" -cron "0 30 9 * * *" -depends "task_abc123,task_def456"`)
}

func printGetUsage() {
	fmt.Println(`Get task details

Usage:
  rs-cli get <task-id>

Examples:
  rs-cli get task_abc123`)
}

func printListUsage() {
	fmt.Println(`List all tasks

Usage:
  rs-cli list [options]

Options:
  -status string    Filter by status (active, paused, error, deleted)

Examples:
  rs-cli list
  rs-cli list -status active`)
}

func printUpdateUsage() {
	fmt.Println(`Update a task

Usage:
  rs-cli update -id <task-id> [options]

Options:
  -id string           Task ID (required)
  -name string         New task name
  -cron string         New cron expression
  -params string       New parameters in format: key1=value1,key2=value2

Examples:
  rs-cli update -id task_abc123 -name "新名称"
  rs-cli update -id task_abc123 -cron "0 0 10 * * *"`)
}

func printDeleteUsage() {
	fmt.Println(`Delete a task

Usage:
  rs-cli delete <task-id>

Note: Execution history is preserved.

Examples:
  rs-cli delete task_abc123`)
}

func printPauseUsage() {
	fmt.Println(`Pause a task

Usage:
  rs-cli pause <task-id>

Note: Missed executions during pause will not be compensated.

Examples:
  rs-cli pause task_abc123`)
}

func printResumeUsage() {
	fmt.Println(`Resume a task

Usage:
  rs-cli resume <task-id>

Examples:
  rs-cli resume task_abc123`)
}

func printTriggerUsage() {
	fmt.Println(`Manually trigger a task

Usage:
  rs-cli trigger <task-id>

Note: Manual trigger does not affect the scheduled plan.

Examples:
  rs-cli trigger task_abc123`)
}

func printExecGetUsage() {
	fmt.Println(`Get execution details

Usage:
  rs-cli exec-get <execution-id>

Examples:
  rs-cli exec-get exec_abc123`)
}

func printExecListUsage() {
	fmt.Println(`List executions

Usage:
  rs-cli exec-list [options]

Options:
  -task string         Filter by task ID
  -status string       Filter by status (pending, running, success, failed, skipped)
  -type string         Filter by trigger type (scheduled, manual)

Examples:
  rs-cli exec-list
  rs-cli exec-list -task task_abc123
  rs-cli exec-list -status failed`)
}

func main() {
	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := strings.ToLower(args[0])
	apiClient := client.NewAPIClient(*hostFlag, *portFlag)

	switch command {
	case "create":
		handleCreate(apiClient, args[1:])
	case "get":
		handleGet(apiClient, args[1:])
	case "list":
		handleList(apiClient, args[1:])
	case "update":
		handleUpdate(apiClient, args[1:])
	case "delete":
		handleDelete(apiClient, args[1:])
	case "pause":
		handlePause(apiClient, args[1:])
	case "resume":
		handleResume(apiClient, args[1:])
	case "trigger":
		handleTrigger(apiClient, args[1:])
	case "exec-get":
		handleExecGet(apiClient, args[1:])
	case "exec-list":
		handleExecList(apiClient, args[1:])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleCreate(apiClient *client.APIClient, args []string) {
	createFlag := flag.NewFlagSet("create", flag.ExitOnError)
	createFlag.Usage = printCreateUsage
	
	namePtr := createFlag.String("name", "", "Task name")
	cronPtr := createFlag.String("cron", "", "Cron expression")
	paramsPtr := createFlag.String("params", "", "Parameters")
	dependsPtr := createFlag.String("depends", "", "Depends on")
	
	createFlag.Parse(args)

	if *namePtr == "" || *cronPtr == "" {
		printCreateUsage()
		os.Exit(1)
	}

	req := &common.CreateTaskRequest{
		Name:           *namePtr,
		CronExpression: *cronPtr,
		Parameters:     client.ParseParameters(*paramsPtr),
		DependsOn:      client.ParseDependsOn(*dependsPtr),
	}

	task, err := apiClient.CreateTask(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Task created successfully!")
	fmt.Println()
	fmt.Println(client.FormatTask(task))
}

func handleGet(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		printGetUsage()
		os.Exit(1)
	}

	taskID := args[0]
	task, err := apiClient.GetTask(taskID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(client.FormatTask(task))
}

func handleList(apiClient *client.APIClient, args []string) {
	listFlag := flag.NewFlagSet("list", flag.ExitOnError)
	listFlag.Usage = printListUsage
	
	statusPtr := listFlag.String("status", "", "Status filter")
	
	listFlag.Parse(args)

	var status *common.TaskStatus
	if *statusPtr != "" {
		ts := common.TaskStatus(*statusPtr)
		status = &ts
	}

	tasks, err := apiClient.ListTasks(status)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(client.FormatTaskList(tasks))
}

func handleUpdate(apiClient *client.APIClient, args []string) {
	updateFlag := flag.NewFlagSet("update", flag.ExitOnError)
	updateFlag.Usage = printUpdateUsage
	
	idPtr := updateFlag.String("id", "", "Task ID")
	namePtr := updateFlag.String("name", "", "New name")
	cronPtr := updateFlag.String("cron", "", "New cron expression")
	paramsPtr := updateFlag.String("params", "", "New parameters")
	
	updateFlag.Parse(args)

	if *idPtr == "" {
		printUpdateUsage()
		os.Exit(1)
	}

	req := &common.UpdateTaskRequest{
		ID:             *idPtr,
		Name:           client.StrPtr(*namePtr),
		CronExpression: client.StrPtr(*cronPtr),
	}

	if *paramsPtr != "" {
		req.Parameters = client.ParseParameters(*paramsPtr)
	}

	task, err := apiClient.UpdateTask(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Task updated successfully!")
	fmt.Println()
	fmt.Println(client.FormatTask(task))
}

func handleDelete(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		printDeleteUsage()
		os.Exit(1)
	}

	taskID := args[0]
	err := apiClient.DeleteTask(taskID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Task deleted successfully. Execution history is preserved.")
}

func handlePause(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		printPauseUsage()
		os.Exit(1)
	}

	taskID := args[0]
	task, err := apiClient.PauseTask(taskID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Task paused successfully.")
	fmt.Println()
	fmt.Println(client.FormatTask(task))
}

func handleResume(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		printResumeUsage()
		os.Exit(1)
	}

	taskID := args[0]
	task, err := apiClient.ResumeTask(taskID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Task resumed successfully.")
	fmt.Println()
	fmt.Println(client.FormatTask(task))
}

func handleTrigger(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		printTriggerUsage()
		os.Exit(1)
	}

	taskID := args[0]
	execID, err := apiClient.TriggerTask(taskID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task triggered successfully.\n")
	fmt.Printf("Execution ID: %s\n", execID)
}

func handleExecGet(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		printExecGetUsage()
		os.Exit(1)
	}

	execID := args[0]
	execution, err := apiClient.GetExecution(execID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(client.FormatExecution(execution))
}

func handleExecList(apiClient *client.APIClient, args []string) {
	execListFlag := flag.NewFlagSet("exec-list", flag.ExitOnError)
	execListFlag.Usage = printExecListUsage
	
	taskPtr := execListFlag.String("task", "", "Task ID filter")
	statusPtr := execListFlag.String("status", "", "Status filter")
	typePtr := execListFlag.String("type", "", "Trigger type filter")
	
	execListFlag.Parse(args)

	var taskID *string
	if *taskPtr != "" {
		taskID = taskPtr
	}

	var status *common.ExecutionStatus
	if *statusPtr != "" {
		es := common.ExecutionStatus(*statusPtr)
		status = &es
	}

	var triggerType *common.TriggerType
	if *typePtr != "" {
		tt := common.TriggerType(*typePtr)
		triggerType = &tt
	}

	executions, err := apiClient.ListExecutions(taskID, status, triggerType)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(client.FormatExecutionList(executions))
}
