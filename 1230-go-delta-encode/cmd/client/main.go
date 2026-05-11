package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"delta-encode/internal/client"
)

const DefaultServerURL = "http://localhost:8211"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := os.Getenv("DELTA_SERVER_URL")
	if serverURL == "" {
		serverURL = DefaultServerURL
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "encode":
		runEncode(serverURL)
	case "decode":
		runDecode(serverURL)
	case "stats":
		runStats(serverURL)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func runEncode(serverURL string) {
	fs := flag.NewFlagSet("encode", flag.ExitOnError)
	dataType := fs.String("type", "int", "Data type: 'int' or 'float'")
	dataFlag := fs.String("data", "", "Comma-separated values (e.g., '1,2,3' or '1.5,2.5')")
	serverFlag := fs.String("server", serverURL, "Server URL")

	fs.Parse(os.Args[2:])

	if *dataFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: -data flag is required")
		printEncodeUsage()
		os.Exit(1)
	}

	values := strings.Split(*dataFlag, ",")
	err := client.Encode(*serverFlag, *dataType, values)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runDecode(serverURL string) {
	fs := flag.NewFlagSet("decode", flag.ExitOnError)
	dataType := fs.String("type", "int", "Data type: 'int' or 'float'")
	baseFlag := fs.String("base", "", "Base value (e.g., '10' or '3.14')")
	deltasFlag := fs.String("deltas", "", "Comma-separated deltas (e.g., '1,2,-1' or '0.1,-0.2')")
	serverFlag := fs.String("server", serverURL, "Server URL")

	fs.Parse(os.Args[2:])

	if *baseFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: -base flag is required")
		printDecodeUsage()
		os.Exit(1)
	}

	var deltas []string
	if *deltasFlag != "" {
		deltas = strings.Split(*deltasFlag, ",")
	}

	err := client.Decode(*serverFlag, *dataType, *baseFlag, deltas)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runStats(serverURL string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	dataType := fs.String("type", "int", "Data type: 'int' or 'float'")
	dataFlag := fs.String("data", "", "Comma-separated values (e.g., '1,2,3' or '1.5,2.5')")
	serverFlag := fs.String("server", serverURL, "Server URL")

	fs.Parse(os.Args[2:])

	if *dataFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: -data flag is required")
		printStatsUsage()
		os.Exit(1)
	}

	values := strings.Split(*dataFlag, ",")
	err := client.Stats(*serverFlag, *dataType, values)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Delta Encode CLI

Usage:
  delta-client <subcommand> [options]

Subcommands:
  encode    Encode a data array using Delta encoding
  decode    Decode Delta encoded data back to original
  stats     Get compression statistics for data
  help      Show this help message

Environment Variables:
  DELTA_SERVER_URL   Server URL (default: http://localhost:8211)

Use "delta-client <subcommand> --help" for more information.`)
}

func printEncodeUsage() {
	fmt.Println(`Usage: delta-client encode [options]

Options:
  -type string    Data type: 'int' or 'float' (default "int")
  -data string    Comma-separated values (e.g., '1,2,3' or '1.5,2.5')
  -server string  Server URL (default from DELTA_SERVER_URL env var)`)
}

func printDecodeUsage() {
	fmt.Println(`Usage: delta-client decode [options]

Options:
  -type string    Data type: 'int' or 'float' (default "int")
  -base string    Base value (e.g., '10' or '3.14')
  -deltas string  Comma-separated deltas (e.g., '1,2,-1' or '0.1,-0.2')
  -server string  Server URL (default from DELTA_SERVER_URL env var)`)
}

func printStatsUsage() {
	fmt.Println(`Usage: delta-client stats [options]

Options:
  -type string    Data type: 'int' or 'float' (default "int")
  -data string    Comma-separated values (e.g., '1,2,3' or '1.5,2.5')
  -server string  Server URL (default from DELTA_SERVER_URL env var)`)
}
