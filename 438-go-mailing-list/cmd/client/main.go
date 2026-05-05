package main

import (
	"os"

	"go-mailing-list/internal/client"
)

const defaultServerURL = "http://localhost:8080"

func main() {
	serverURL := os.Getenv("MLC_SERVER")
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	cli := client.NewCLI(serverURL)
	args := os.Args[1:]

	if len(args) == 0 {
		cli.Run([]string{"help"})
		return
	}

	cli.Run(args)
}
