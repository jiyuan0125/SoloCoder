package main

import (
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

	switch cmd {
	case "mask":
		handleMask(args)
	case "rules":
		handleRules(args)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client mask <text> [options]  - Mask sensitive data in text")
	fmt.Println("  client rules [options]         - Get or update rules")
	fmt.Println("")
	fmt.Println("Global options:")
	fmt.Println("  -server <url>  Server URL (default: http://localhost:8080)")
	fmt.Println("")
	fmt.Println("Mask options:")
	fmt.Println("  -phone-prefix <n>    Phone prefix keep")
	fmt.Println("  -phone-suffix <n>    Phone suffix keep")
	fmt.Println("  -idcard-prefix <n>   ID card prefix keep")
	fmt.Println("  -idcard-suffix <n>   ID card suffix keep")
	fmt.Println("  -bankcard-prefix <n> Bank card prefix keep")
	fmt.Println("  -bankcard-suffix <n> Bank card suffix keep")
	fmt.Println("  -email-prefix <n>    Email prefix keep")
	fmt.Println("  -name-prefix <n>     Name prefix keep")
	fmt.Println("")
	fmt.Println("Rules options:")
	fmt.Println("  -update              Update rules instead of getting")
	fmt.Println("  -phone-prefix <n>    Phone prefix keep")
	fmt.Println("  -phone-suffix <n>    Phone suffix keep")
	fmt.Println("  -idcard-prefix <n>   ID card prefix keep")
	fmt.Println("  -idcard-suffix <n>   ID card suffix keep")
	fmt.Println("  -bankcard-prefix <n> Bank card prefix keep")
	fmt.Println("  -bankcard-suffix <n> Bank card suffix keep")
	fmt.Println("  -email-prefix <n>    Email prefix keep")
	fmt.Println("  -name-prefix <n>     Name prefix keep")
}
