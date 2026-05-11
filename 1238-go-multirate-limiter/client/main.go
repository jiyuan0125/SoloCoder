package main

import (
	"flag"
	"fmt"
	"os"
)

const defaultServerURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmdName := os.Args[1]
	if cmdName == "-h" || cmdName == "--help" || cmdName == "help" {
		printUsage()
		return
	}

	cmd := findCommand(cmdName)
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmdName)
		printUsage()
		os.Exit(1)
	}

	args := os.Args[2:]
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			printCommandHelp(cmd)
			return
		}
	}

	serverURL := defaultServerURL
	if envURL := os.Getenv("LIMITER_SERVER"); envURL != "" {
		serverURL = envURL
	}

	urlFlag := flag.NewFlagSet("url", flag.ContinueOnError)
	urlFlag.StringVar(&serverURL, "server", serverURL, "server URL")
	urlFlag.Parse(args)

	if err := cmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		printCommandHelp(cmd)
		os.Exit(1)
	}

	httpClient := NewHTTPClient(serverURL)
	if err := cmd.Run(httpClient); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
