package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: barcode-client [flags] <input-string>")
		fmt.Println("Flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	input := flag.Arg(0)

	client := NewClient(*serverURL)
	response, err := client.Encode(input)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !response.Success {
		fmt.Printf("Error: %s\n", response.Error)
		os.Exit(1)
	}

	fmt.Println("Barcode encoding result:")
	fmt.Printf("  Input: %s\n", response.Input)
	fmt.Printf("  Pattern: %s\n", response.Pattern)
	fmt.Printf("  Total modules: %d\n", response.TotalModules)
	fmt.Printf("  Width sequence: %v\n", response.WidthSequence)
}
