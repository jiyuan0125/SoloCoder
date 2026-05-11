package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "import":
		handleImport(args)
	case "weight":
		handleWeight(args)
	case "mine":
		handleMine(args)
	case "rules":
		handleRules(args)
	case "items":
		handleItems(args)
	default:
		fmt.Fprintf(os.Stderr, "错误: 未知命令 '%s'\n", command)
		fmt.Println()
		printUsage()
	}
}
