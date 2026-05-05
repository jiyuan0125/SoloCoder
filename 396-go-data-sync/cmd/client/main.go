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
	
	command := os.Args[1]
	
	switch command {
	case "sync":
		handleSyncCommand(os.Args[2:])
	case "status":
		handleStatusCommand(os.Args[2:])
	case "history":
		handleHistoryCommand(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Go Data Sync Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client sync [options]     - Perform data synchronization")
	fmt.Println("  client status              - Check server status")
	fmt.Println("  client history             - Show synchronization history")
	fmt.Println()
	fmt.Println("Sync Options:")
	fmt.Println("  --source <file>    Source CSV file (required)")
	fmt.Println("  --target <file>    Target CSV file (required)")
	fmt.Println("  --key <column>     Key column name (default: first column)")
	fmt.Println("  --apply            Apply changes to target file")
	fmt.Println("  --host <host>      Server host (default: 127.0.0.1)")
	fmt.Println("  --port <port>      Server port (default: 8888)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client sync --source source.csv --target target.csv")
	fmt.Println("  client sync --source source.csv --target target.csv --key ID --apply")
	fmt.Println("  client status")
	fmt.Println("  client history")
}
