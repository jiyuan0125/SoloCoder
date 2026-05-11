package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	serverURL := os.Getenv("CMS_SERVER")
	if serverURL == "" {
		serverURL = "http://localhost:8215"
	}

	fs := flag.NewFlagSet("client", flag.ContinueOnError)
	fs.StringVar(&serverURL, "server", serverURL, "server URL")
	fs.Parse(args)

	remaining := fs.Args()

	var err error
	switch cmd {
	case "create":
		err = cmdCreate(serverURL, remaining)
	case "add":
		err = cmdAdd(serverURL, remaining)
	case "query":
		err = cmdQuery(serverURL, remaining)
	case "merge":
		err = cmdMerge(serverURL, remaining)
	case "stats":
		err = cmdStats(serverURL, remaining)
	default:
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage: cms-client <command> [options]

Commands:
  create  -name <name> [-width <w>] [-depth <d>]
  add     -name <name> [-item <item>] [-count <n>] [-file <path>]
  query   -name <name> [-item <item>]
  merge   -source <name> -target <name>
  stats   [-name <name>]

Options:
  -server <url>  Server URL (default: http://localhost:8215)`)
}
