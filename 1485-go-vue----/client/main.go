package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	defaultServerURL = "http://localhost:8080"
)

func main() {
	serverURL := flag.String("server", defaultServerURL, "Server URL")
	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	client := NewAPIClient(*serverURL)
	runner := NewCommandRunner(client)

	cmd := args[0]
	cmdArgs := args[1:]

	var err error
	switch cmd {
	case "create-task":
		err = runner.RunCreateTask(cmdArgs)
	case "list-tasks":
		err = runner.RunListTasks(cmdArgs)
	case "get-task":
		err = runner.RunGetTask(cmdArgs)
	case "end-task":
		err = runner.RunEndTask(cmdArgs)
	case "submit-data":
		err = runner.RunSubmitSensorData(cmdArgs)
	case "get-data":
		err = runner.RunGetSensorData(cmdArgs)
	case "get-alerts":
		err = runner.RunGetAlerts(cmdArgs)
	case "get-stats":
		err = runner.RunGetStats(cmdArgs)
	case "get-report":
		err = runner.RunGetReport(cmdArgs)
	case "export":
		err = runner.RunExportCSV(cmdArgs)
	case "help", "-h", "--help":
		printUsage()
		return
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("ColdChain CLI - Cold Chain Logistics Monitoring System")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client [--server <url>] <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create-task <device_id> <cargo_type> <start_warehouse> <end_warehouse>")
	fmt.Println("      Create a new transport task")
	fmt.Println("      Cargo types: frozen, refrigerated, medicine, normal")
	fmt.Println()
	fmt.Println("  list-tasks")
	fmt.Println("      List all transport tasks")
	fmt.Println()
	fmt.Println("  get-task <task_id>")
	fmt.Println("      Get details of a specific task")
	fmt.Println()
	fmt.Println("  end-task <task_id>")
	fmt.Println("      End a transport task")
	fmt.Println()
	fmt.Println("  submit-data <device_id> <temp> <humidity> <lat> <lon> <speed> [timestamp]")
	fmt.Println("      Submit sensor data (timestamp in RFC3339 format, optional)")
	fmt.Println()
	fmt.Println("  get-data <task_id>")
	fmt.Println("      Get all sensor data for a task")
	fmt.Println()
	fmt.Println("  get-alerts <task_id>")
	fmt.Println("      Get all alerts for a task")
	fmt.Println()
	fmt.Println("  get-stats <task_id>")
	fmt.Println("      Get statistics for a task")
	fmt.Println()
	fmt.Println("  get-report <task_id>")
	fmt.Println("      Get full transport report")
	fmt.Println()
	fmt.Println("  export <task_id> <type> <output_file>")
	fmt.Println("      Export data to CSV")
	fmt.Println("      Types: sensor, stats, alerts")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --server <url>  Server URL (default: http://localhost:8080)")
}
