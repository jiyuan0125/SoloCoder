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
	case "mask-type":
		handleMaskType(args)
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
	fmt.Println("  client mask <text> [options]               - Auto-detect and mask sensitive data in text")
	fmt.Println("  client mask-type <value> -type <type> [options] - Mask value with explicit data type")
	fmt.Println("  client rules [options]                      - Get or update global rules")
	fmt.Println("")
	fmt.Println("Global options:")
	fmt.Println("  -server <url>  Server URL (default: http://localhost:8102)")
	fmt.Println("")
	fmt.Println("mask-type options:")
	fmt.Println("  -type <type>       Data type (required): phone, idcard, bankcard, email, name")
	fmt.Println("  -prefix <n>        Prefix keep (optional override)")
	fmt.Println("  -suffix <n>        Suffix keep (optional override)")
	fmt.Println("")
	fmt.Println("mask options:")
	fmt.Println("  -phone-prefix <n>    Phone prefix keep")
	fmt.Println("  -phone-suffix <n>    Phone suffix keep")
	fmt.Println("  -idcard-prefix <n>   ID card prefix keep")
	fmt.Println("  -idcard-suffix <n>   ID card suffix keep")
	fmt.Println("  -bankcard-prefix <n> Bank card prefix keep")
	fmt.Println("  -bankcard-suffix <n> Bank card suffix keep")
	fmt.Println("  -email-prefix <n>    Email prefix keep")
	fmt.Println("  -name-prefix <n>     Name prefix keep")
	fmt.Println("")
	fmt.Println("rules options:")
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
