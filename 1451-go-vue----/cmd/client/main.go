package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	client := NewClient(serverURL)
	module := os.Args[1]

	if module == "help" {
		printUsage()
		return
	}

	switch module {
	case "access":
		handleAccessCommands(client, os.Args[2:])
	case "visitor":
		handleVisitorCommands(client, os.Args[2:])
	case "patrol":
		handlePatrolCommands(client, os.Args[2:])
	default:
		fmt.Printf("Unknown module: %s\n", module)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Smart Park Management System CLI")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  client <module> <command> [args]")
	fmt.Println("")
	fmt.Println("Modules:")
	fmt.Println("  access   - Access control management")
	fmt.Println("  visitor  - Visitor management")
	fmt.Println("  patrol   - Patrol management")
	fmt.Println("  help     - Show this help")
	fmt.Println("")
	fmt.Println("Access Commands:")
	fmt.Println("  add-point <id> <name> <area_id> <building_id>")
	fmt.Println("  list-points")
	fmt.Println("  add-rule <point_id> <dept1,dept2> <start> <end>")
	fmt.Println("  batch-rules <area_id> <dept1,dept2> <start> <end>")
	fmt.Println("  verify <employee_id> <point_id> <type>")
	fmt.Println("  list-records")
	fmt.Println("  fix-point <point_id>")
	fmt.Println("")
	fmt.Println("Visitor Commands:")
	fmt.Println("  reserve <employee_id> <name> <phone> <purpose> <arrival_time>")
	fmt.Println("  review <reservation_id> <approve/reject>")
	fmt.Println("  checkin <phone>")
	fmt.Println("  list")
	fmt.Println("")
	fmt.Println("Patrol Commands:")
	fmt.Println("  add-point <id> <name>")
	fmt.Println("  add-route <name> <point1,point2,...>")
	fmt.Println("  list-routes")
	fmt.Println("  add-task <route_id> <assignee> <frequency> <start_time>")
	fmt.Println("  list-tasks")
	fmt.Println("  start <task_id>")
	fmt.Println("  checkin <execution_id> <point_id> [has_anomaly] [desc] [related_ap]")
	fmt.Println("  list-checkins <execution_id>")
	fmt.Println("")
	fmt.Println("Environment:")
	fmt.Println("  SERVER_URL - Server base URL (default: http://localhost:8080)")
}
