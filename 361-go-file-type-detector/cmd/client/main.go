package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var filePath string
	var help bool

	flag.StringVar(&filePath, "file", "", "Path to the file to detect")
	flag.BoolVar(&help, "help", false, "Show help information")
	flag.Parse()

	if help || filePath == "" {
		printUsage()
		os.Exit(0)
	}

	result, err := DetectFile(filePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("File: %s\n", filePath)
	fmt.Printf("MIME Type: %s\n", result.MimeType)
	if result.Error != "" {
		fmt.Printf("Error: %s\n", result.Error)
	}
}

func printUsage() {
	fmt.Println("File Type Detector Client")
	fmt.Println("Usage:")
	fmt.Println("  client -file <filepath>")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -file string   Path to the file to detect")
	fmt.Println("  -help          Show help information")
}
