package main

import (
	"fmt"
	"os"
)

const defaultServerURL = "http://localhost:8909"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	client := NewAPIClient(serverURL)
	commander := NewCommander(client)

	command := os.Args[1]
	args := os.Args[2:]

	var err error

	switch command {
	case "register-vehicle":
		err = commander.RegisterVehicle(args)
	case "vehicle-info":
		err = commander.GetVehicleInfo(args)
	case "list-vehicles":
		err = commander.ListVehicles(args)
	case "create-station":
		err = commander.CreateStation(args)
	case "add-schedule":
		err = commander.AddSchedule(args)
	case "list-stations":
		err = commander.ListStations(args)
	case "create-appointment":
		err = commander.CreateAppointment(args)
	case "list-appointments":
		err = commander.ListAppointments(args)
	case "start-inspection":
		err = commander.StartInspection(args)
	case "complete-step":
		err = commander.CompleteStep(args)
	case "start-recheck":
		err = commander.StartRecheck(args)
	case "list-processes":
		err = commander.ListProcesses(args)
	case "help", "-h", "--help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Vehicle Inspection System Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client <command> [arguments]")
	fmt.Println()
	fmt.Println("Vehicle Commands:")
	fmt.Println("  register-vehicle <plate> <type> <register-date> [last-inspection-date]")
	fmt.Println("  vehicle-info <plate>")
	fmt.Println("  list-vehicles")
	fmt.Println()
	fmt.Println("Station Commands:")
	fmt.Println("  create-station <name>")
	fmt.Println("  add-schedule <station-id> <date> <time-slot> <max-vehicles>")
	fmt.Println("  list-stations")
	fmt.Println()
	fmt.Println("Appointment Commands:")
	fmt.Println("  create-appointment <plate> <station-id> <date> <time-slot>")
	fmt.Println("  list-appointments")
	fmt.Println()
	fmt.Println("Inspection Commands:")
	fmt.Println("  start-inspection <appointment-id>")
	fmt.Println("  complete-step <process-id> <result> [fail-items-json]")
	fmt.Println("  start-recheck <process-id>")
	fmt.Println("  list-processes")
	fmt.Println()
	fmt.Println("Vehicle Types: 小型轿车, SUV, MPV, 货车, 客车")
	fmt.Println("Results: 合格, 不合格, 需复检")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  SERVER_URL   Server base URL (default: http://localhost:8909)")
}
