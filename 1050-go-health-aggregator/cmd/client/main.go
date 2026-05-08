package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

const usage = `Health Aggregator Client

Usage:
  hac [command] [options]

Commands:
  status [service]   Show status (all services or specific one)
  history <name> [n] Show probe history for a service (last n records)
  add <name> <http|tcp> <target> [interval] [timeout] [failure_threshold] [recovery_threshold] [group]
                     Add a new service to monitor
  remove <name>      Remove a service from monitoring
  sync               Sync services from config file to server
  dashboard          Show real-time dashboard

Options:
  -server URL        Server URL (default: http://localhost:8080)
  -config PATH       Config file path (for sync command)
  -interval DURATION Dashboard refresh interval (default: 5s)
  -help              Show this help
`

func main() {
	fs := flag.NewFlagSet("hac", flag.ContinueOnError)
	serverURL := fs.String("server", "http://localhost:8080", "Server URL")
	configPath := fs.String("config", "", "Config file path")
	intervalStr := fs.String("interval", "5s", "Dashboard refresh interval")
	help := fs.Bool("help", false, "Show help")

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Print(usage)
		os.Exit(0)
	}

	cmd := args[0]
	remaining := args[1:]

	if err := fs.Parse(remaining); err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(usage)
		os.Exit(0)
	}

	if *help {
		fmt.Print(usage)
		os.Exit(0)
	}

	remaining = fs.Args()

	interval, err := time.ParseDuration(*intervalStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid interval: %v\n", err)
		os.Exit(1)
	}

	apiClient := NewAPIClient(*serverURL)

	switch cmd {
	case "help", "-help", "--help":
		fmt.Print(usage)

	case "status":
		serviceName := ""
		if len(remaining) > 0 {
			serviceName = remaining[0]
		}
		if err := CmdStatus(apiClient, serviceName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "history":
		if len(remaining) < 1 {
			fmt.Fprintln(os.Stderr, "usage: history <name> [n]")
			os.Exit(1)
		}
		name := remaining[0]
		limit := 0
		if len(remaining) > 1 {
			fmt.Sscanf(remaining[1], "%d", &limit)
		}
		if err := CmdHistory(apiClient, name, limit); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "add":
		if len(remaining) < 3 {
			fmt.Fprintln(os.Stderr, "usage: add <name> <http|tcp> <target> [interval] [timeout] [failure_threshold] [recovery_threshold] [group]")
			os.Exit(1)
		}
		if err := CmdAdd(apiClient, remaining); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "remove":
		if len(remaining) < 1 {
			fmt.Fprintln(os.Stderr, "usage: remove <name>")
			os.Exit(1)
		}
		if err := CmdRemove(apiClient, remaining[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "sync":
		if *configPath == "" {
			fmt.Fprintln(os.Stderr, "config file required: use -config flag")
			os.Exit(1)
		}
		cfg, err := LoadClientConfig(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
			os.Exit(1)
		}
		if err := CmdSync(apiClient, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "dashboard":
		if err := CmdDashboard(apiClient, interval); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		fmt.Print(usage)
		os.Exit(1)
	}
}
