package main

import (
	"fmt"
	"os"

	"csv-merger/client/cli"
	"csv-merger/client/network"
)

func main() {
	config, err := cli.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		cli.Parse()
		os.Exit(1)
	}

	client := network.NewClient(config.ServerHost, config.ServerPort)

	response, err := client.SendMergeRequest(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if response.Success {
		fmt.Printf("Success: %s\n", response.Message)
		fmt.Printf("Records merged: %d\n", response.RecordsMerged)
		os.Exit(0)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", response.Error)
		os.Exit(1)
	}
}
