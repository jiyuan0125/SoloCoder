package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	formatCmd := flag.NewFlagSet("format", flag.ExitOnError)
	validateCmd := flag.NewFlagSet("validate", flag.ExitOnError)
	extractCmd := flag.NewFlagSet("extract", flag.ExitOnError)

	formatPhone := formatCmd.String("phone", "", "Phone number to format")
	formatType := formatCmd.String("type", "domestic", "Format type: domestic, international, pure")

	validatePhone := validateCmd.String("phone", "", "Phone number to validate")

	extractText := extractCmd.String("text", "", "Text to extract phone numbers from")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'format', 'validate' or 'extract' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "format":
		formatCmd.Parse(os.Args[2:])
		if *formatPhone == "" {
			fmt.Println("Please provide a phone number with -phone flag")
			os.Exit(1)
		}
		result, err := FormatPhone(*formatPhone, *formatType)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)

	case "validate":
		validateCmd.Parse(os.Args[2:])
		if *validatePhone == "" {
			fmt.Println("Please provide a phone number with -phone flag")
			os.Exit(1)
		}
		valid, err := ValidatePhone(*validatePhone)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if valid {
			fmt.Println("Valid phone number")
		} else {
			fmt.Println("Invalid phone number")
		}

	case "extract":
		extractCmd.Parse(os.Args[2:])
		if *extractText == "" {
			fmt.Println("Please provide text with -text flag")
			os.Exit(1)
		}
		results, err := ExtractPhones(*extractText)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		for _, phone := range results {
			fmt.Println(phone)
		}

	default:
		fmt.Println("Expected 'format', 'validate' or 'extract' subcommands")
		os.Exit(1)
	}
}
