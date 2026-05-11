package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := flag.String("server", "http://localhost:8080", "server URL")
	flag.CommandLine.Parse(os.Args[2:])

	client := NewClient(*serverURL)

	command := strings.ToLower(os.Args[1])
	switch command {
	case "register-gas":
		cmdRegisterGas(client)
	case "submit-gas":
		cmdSubmitGas(client)
	case "query-gas":
		cmdQueryGas(client)
	case "register-wastewater":
		cmdRegisterWastewater(client)
	case "submit-wastewater":
		cmdSubmitWastewater(client)
	case "query-wastewater":
		cmdQueryWastewater(client)
	case "daily-report":
		cmdDailyReport(client)
	case "add-solid-waste":
		cmdAddSolidWaste(client)
	case "query-solid-waste":
		cmdQuerySolidWaste(client)
	case "query-alarms":
		cmdQueryAlarms(client)
	case "resolve-alarm":
		cmdResolveAlarm(client)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Environment Monitor Client")
	fmt.Println("\nUsage:")
	fmt.Println("  client <command> [flags]")
	fmt.Println("\nCommands:")
	fmt.Println("  register-gas        Register a gas outlet")
	fmt.Println("  submit-gas          Submit gas monitoring data")
	fmt.Println("  query-gas           Query gas reports")
	fmt.Println("  register-wastewater Register a wastewater outlet")
	fmt.Println("  submit-wastewater   Submit wastewater monitoring data")
	fmt.Println("  query-wastewater    Query wastewater reports")
	fmt.Println("  daily-report        Get wastewater daily report")
	fmt.Println("  add-solid-waste     Add solid waste record")
	fmt.Println("  query-solid-waste   Query solid waste records")
	fmt.Println("  query-alarms        Query alarms")
	fmt.Println("  resolve-alarm       Resolve an alarm")
	fmt.Println("\nFlags:")
	fmt.Println("  -server string    Server URL (default \"http://localhost:8080\")")
}
