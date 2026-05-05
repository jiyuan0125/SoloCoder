package main

import (
	"fmt"
	"os"

	"go-feedback-handler/internal/client"
)

const (
	EnvServerURL = "FEEDBACK_SERVER"
)

func main() {
	baseURL := os.Getenv(EnvServerURL)
	if baseURL == "" {
		baseURL = client.DefaultBaseURL
	}

	cli := client.NewCLI(baseURL)

	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
