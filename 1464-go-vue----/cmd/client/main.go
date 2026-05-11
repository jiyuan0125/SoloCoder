package main

import (
	"fmt"
	"os"
)

func printUsage() {
	fmt.Println(`Fire Management System CLI

Usage:
  fm [command] [subcommand] [args...]

Commands:
  device    - Manage fire devices
    create <code> <type> <location> <install_date> <expiry_date> <last_check_date>
    list [type]
    get <id>
    update-status <id> <status>

  point     - Manage inspection points
    create <name> <location>
    list

  route     - Manage inspection routes
    create <name> <point_id1> [point_id2...]
    list

  plan      - Manage inspection plans
    create <route_id> <inspector_id> <inspector_name> <frequency> <start_time>
    list

  task      - Manage inspection tasks
    list
    get <id>
    check <task_id> <point_id> <normal:true|false>
    summary <task_id>

  drill     - Manage fire drills
    create-plan <type> <name> <scheduled_time> <planned_attendees>
    list-plans
    complete <plan_id> <actual_attendees> <duration_minutes> <improvements...>
    list-records

  reminder  - Manage reminders
    list [unread|read]
    read <id> [true|false]

Environment:
  SERVER_URL  - Server URL (default: http://localhost:8080)
`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient("")

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "device":
		handleDevice(client, args)
	case "point":
		handlePoint(client, args)
	case "route":
		handleRoute(client, args)
	case "plan":
		handlePlan(client, args)
	case "task":
		handleTask(client, args)
	case "drill":
		handleDrill(client, args)
	case "reminder":
		handleReminder(client, args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleDevice(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: device <create|list|get|update-status> [args...]")
		os.Exit(1)
	}
	switch args[0] {
	case "create":
		cmdDeviceCreate(client, args[1:])
	case "list":
		cmdDeviceList(client, args[1:])
	case "get":
		cmdDeviceGet(client, args[1:])
	case "update-status":
		cmdDeviceUpdateStatus(client, args[1:])
	default:
		fmt.Printf("Unknown device subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func handlePoint(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: point <create|list> [args...]")
		os.Exit(1)
	}
	switch args[0] {
	case "create":
		cmdPointCreate(client, args[1:])
	case "list":
		cmdPointList(client, args[1:])
	default:
		fmt.Printf("Unknown point subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func handleRoute(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: route <create|list> [args...]")
		os.Exit(1)
	}
	switch args[0] {
	case "create":
		cmdRouteCreate(client, args[1:])
	case "list":
		cmdRouteList(client, args[1:])
	default:
		fmt.Printf("Unknown route subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func handlePlan(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: plan <create|list> [args...]")
		os.Exit(1)
	}
	switch args[0] {
	case "create":
		cmdPlanCreate(client, args[1:])
	case "list":
		cmdPlanList(client, args[1:])
	default:
		fmt.Printf("Unknown plan subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func handleTask(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: task <list|get|check|summary> [args...]")
		os.Exit(1)
	}
	switch args[0] {
	case "list":
		cmdTaskList(client, args[1:])
	case "get":
		cmdTaskGet(client, args[1:])
	case "check":
		cmdTaskCheck(client, args[1:])
	case "summary":
		cmdTaskSummary(client, args[1:])
	default:
		fmt.Printf("Unknown task subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func handleDrill(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: drill <create-plan|list-plans|complete|list-records> [args...]")
		os.Exit(1)
	}
	switch args[0] {
	case "create-plan":
		cmdDrillCreatePlan(client, args[1:])
	case "list-plans":
		cmdDrillListPlans(client, args[1:])
	case "complete":
		cmdDrillComplete(client, args[1:])
	case "list-records":
		cmdDrillListRecords(client, args[1:])
	default:
		fmt.Printf("Unknown drill subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func handleReminder(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: reminder <list|read> [args...]")
		os.Exit(1)
	}
	switch args[0] {
	case "list":
		cmdReminderList(client, args[1:])
	case "read":
		cmdReminderRead(client, args[1:])
	default:
		fmt.Printf("Unknown reminder subcommand: %s\n", args[0])
		os.Exit(1)
	}
}
