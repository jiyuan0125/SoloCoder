package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const defaultServerURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := getServerURL()
	client := NewClient(serverURL)

	command := os.Args[1]
	args := os.Args[2:]

	var err error
	switch command {
	case "estimate":
		err = cmdEstimate(client, args)
	case "book":
		err = cmdBook(client, args)
	case "list":
		err = cmdList(client)
	case "get":
		err = cmdGet(client, args)
	case "cancel":
		err = cmdCancel(client, args)
	case "complete":
		err = cmdComplete(client, args)
	case "settle":
		err = cmdSettle(client, args)
	case "review":
		err = cmdReview(client, args)
	case "availability":
		err = cmdAvailability(client, args)
	default:
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func getServerURL() string {
	url := defaultServerURL
	if envURL := os.Getenv("SERVER_URL"); envURL != "" {
		url = envURL
	}

	flagURL := flag.String("server", url, "Server URL")
	flag.Parse()

	if *flagURL != "" {
		url = *flagURL
	}

	return url
}

func printUsage() {
	fmt.Println("Moving Platform Client")
	fmt.Println("Usage:")
	fmt.Println("  moving-client <command> [arguments]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  estimate      - Get price estimate")
	fmt.Println("  book          - Create a new booking")
	fmt.Println("  list          - List all bookings")
	fmt.Println("  get <id>      - Get booking details")
	fmt.Println("  cancel <id>   - Cancel a booking")
	fmt.Println("  complete <id> - Mark booking as completed")
	fmt.Println("  settle <id>   - Settle a booking with final details")
	fmt.Println("  review <id>   - Add review to a booking")
	fmt.Println("  availability  - Check available slots for a date")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  -server <url> - Server URL (default: http://localhost:8080)")
}

func formatCents(cents int64) string {
	yuan := cents / 100
	remainder := cents % 100
	if remainder == 0 {
		return fmt.Sprintf("%d元", yuan)
	}
	return fmt.Sprintf("%d.%02d元", yuan, remainder)
}

func joinStrings(s []string, sep string) string {
	return strings.Join(s, sep)
}
