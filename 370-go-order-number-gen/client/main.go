package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	serverAddr = flag.String("server", "http://localhost:8080", "Server address")
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(*serverAddr)

	command := strings.ToLower(args[0])
	switch command {
	case "generate":
		handleGenerate(client, args[1:])
	case "parse":
		handleParse(client, args[1:])
	case "prefix":
		handlePrefix(client, args[1:])
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleGenerate(client *Client, args []string) {
	var prefix string
	if len(args) > 0 {
		prefix = args[0]
	}

	orderNo, err := client.Generate(prefix)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(orderNo)
}

func handleParse(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Error: Order number is required")
		os.Exit(1)
	}

	orderNo := args[0]
	info, err := client.Parse(orderNo)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Order No: %s\n", info.OrderNo)
	if info.Prefix != "" {
		fmt.Printf("Prefix: %s\n", info.Prefix)
	}
	fmt.Printf("Timestamp: %s\n", info.Timestamp)
	fmt.Printf("Serial Num: %d\n", info.SerialNum)
	fmt.Printf("Random Num: %d\n", info.RandomNum)
}

func handlePrefix(client *Client, args []string) {
	if len(args) == 0 {
		prefix, err := client.GetPrefix()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if prefix == "" {
			fmt.Println("(no prefix set)")
		} else {
			fmt.Println(prefix)
		}
	} else {
		prefix := args[0]
		if err := client.SetPrefix(prefix); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Prefix set to: %s\n", prefix)
	}
}

func printUsage() {
	fmt.Println("Order Number Generator Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client generate [prefix]   Generate a new order number")
	fmt.Println("  client parse <order_no>    Parse an existing order number")
	fmt.Println("  client prefix               Get current prefix")
	fmt.Println("  client prefix <prefix>      Set default prefix")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -server string   Server address (default \"http://localhost:8080\")")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client generate")
	fmt.Println("  client generate DD")
	fmt.Println("  client parse DD202605041230000000011234")
	fmt.Println("  client prefix")
	fmt.Println("  client prefix TK")
}
