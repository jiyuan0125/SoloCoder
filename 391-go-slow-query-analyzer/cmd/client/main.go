package main

import (
	"flag"
	"fmt"
	"os"

	"slowquery/internal/client"
	"slowquery/protocol"
)

const defaultServerAddr = "localhost:" + protocol.DefaultPort

func main() {
	flag.Usage = usage
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "start":
		handleStart()
	case "stop":
		handleStop()
	case "status":
		handleStatus()
	case "result":
		handleResult()
	case "show":
		handleShow()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Slow Query Log Analyzer - Client

Usage:
  slowquery-client start [options] <log-file>
  slowquery-client stop
  slowquery-client status
  slowquery-client result [options]
  slowquery-client show <rank>

Commands:
  start      Start analyzing a log file
  stop       Stop the current analysis
  status     Check server status
  result     Get and display analysis results
  show       Show full details of a specific template by rank

Options for 'start':
  --format <format>     Log format: "delimiter:col1,col2,col3,col4"
                       Default: "\t:sql,exec_time,scan_rows,lock_wait"
                       Columns: sql, exec_time, scan_rows, lock_wait
  --min-time <ms>    Minimum execution time in milliseconds to include (default: 0)
  --server <addr>    Server address (default: localhost:9876)

Options for 'result':
  --top <n>          Number of top results to show (default: 20)
  --sort-by <type>   Sort by: avg_exec_time or total_exec_time (default: avg_exec_time)
  --server <addr>    Server address (default: localhost:9876)

Examples:
  slowquery-client start slow.log
  slowquery-client start --format "|:sql,exec_time,scan_rows" slow.log
  slowquery-client start --min-time 100 slow.log
  slowquery-client result
  slowquery-client result --top 50 --sort-by total_exec_time
  slowquery-client show 1
`)
}

func handleStart() {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	formatFlag := fs.String("format", "", "Log format")
	minTimeFlag := fs.Float64("min-time", 0, "Minimum execution time")
	serverFlag := fs.String("server", defaultServerAddr, "Server address")

	fs.Parse(os.Args[2:])

	args := fs.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: log file path is required")
		os.Exit(1)
	}

	logFilePath := args[0]

	c := client.NewClient(*serverFlag)
	if err := c.StartAnalysis(logFilePath, *formatFlag, *minTimeFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleStop() {
	fs := flag.NewFlagSet("stop", flag.ExitOnError)
	serverFlag := fs.String("server", defaultServerAddr, "Server address")
	fs.Parse(os.Args[2:])

	c := client.NewClient(*serverFlag)
	if err := c.StopAnalysis(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleStatus() {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	serverFlag := fs.String("server", defaultServerAddr, "Server address")
	fs.Parse(os.Args[2:])

	c := client.NewClient(*serverFlag)
	status, err := c.GetStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Server status: %s\n", status)
}

func handleResult() {
	fs := flag.NewFlagSet("result", flag.ExitOnError)
	topFlag := fs.Int("top", 20, "Number of top results")
	sortFlag := fs.String("sort-by", "avg_exec_time", "Sort by: avg_exec_time or total_exec_time")
	serverFlag := fs.String("server", defaultServerAddr, "Server address")
	fs.Parse(os.Args[2:])

	var sortBy protocol.SortType
	switch *sortFlag {
	case "total_exec_time":
		sortBy = protocol.SortByTotalExecTime
	case "avg_exec_time":
		fallthrough
	default:
		sortBy = protocol.SortByAvgExecTime
	}

	c := client.NewClient(*serverFlag)
	stats, err := c.GetStats(*topFlag, sortBy)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printer := client.NewTablePrinter(os.Stdout)
	printer.PrintSummary(stats)
	printer.PrintTable(stats)
}

func handleShow() {
	fs := flag.NewFlagSet("show", flag.ExitOnError)
	serverFlag := fs.String("server", defaultServerAddr, "Server address")
	fs.Parse(os.Args[2:])

	args := fs.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: rank number is required")
		os.Exit(1)
	}

	var rank int
	fmt.Sscanf(args[0], "%d", &rank)
	if rank <= 0 {
		fmt.Fprintln(os.Stderr, "Error: rank must be a positive number")
		os.Exit(1)
	}

	c := client.NewClient(*serverFlag)
	stats, err := c.GetStats(1000, protocol.SortByAvgExecTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printer := client.NewTablePrinter(os.Stdout)
	printer.PrintFullTemplate(stats, rank)
}
