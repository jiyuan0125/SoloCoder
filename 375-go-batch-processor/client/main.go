package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	flag.Usage = printUsage

	serverURL := flag.String("server", "http://localhost:8080", "Server URL")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]

	client := NewClient(*serverURL)

	var err error
	switch cmd {
	case "submit":
		if len(args) < 2 {
			fmt.Println("Error: submit requires at least one data item")
			fmt.Println("Usage: client submit <data1> [data2] [data3]...")
			os.Exit(1)
		}
		data := args[1:]
		err = client.Submit(data)

	case "flush":
		err = client.Flush()

	case "stats":
		err = client.Stats()

	case "shutdown":
		err = client.Shutdown()

	default:
		fmt.Printf("Error: unknown command '%s'\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	usage := `Batch Processor Client

Usage:
  client [options] <command> [arguments]

Commands:
  submit <data1> [data2]...  Submit one or more data items
  flush                       Immediately process all buffered data
  stats                       Get current buffer size
  shutdown                    Shutdown the server

Options:
  -server string    Server URL (default "http://localhost:8080")
  -help             Show this help message

Examples:
  client submit "hello" "world"
  client -server http://localhost:9999 stats
  client flush
  client shutdown
`
	fmt.Println(strings.TrimSpace(usage))
}
