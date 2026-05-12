package main

import (
	"fmt"
	"os"

	"health-archive/internal/cli"
)

func main() {
	apiBase := os.Getenv("API_BASE")

	c := cli.New(apiBase)
	if err := c.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
