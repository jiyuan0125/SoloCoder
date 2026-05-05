package main

import (
	"flag"
	"fmt"
	"os"
)

func handleSyncCommand(args []string) {
	fs := flag.NewFlagSet("sync", flag.ExitOnError)
	
	sourceFile := fs.String("source", "", "Source CSV file")
	targetFile := fs.String("target", "", "Target CSV file")
	keyColumn := fs.String("key", "", "Key column name")
	apply := fs.Bool("apply", false, "Apply changes to target file")
	host := fs.String("host", "127.0.0.1", "Server host")
	port := fs.String("port", "8888", "Server port")
	
	fs.Parse(args)
	
	if *sourceFile == "" {
		fmt.Println("Error: --source is required")
		os.Exit(1)
	}
	
	if *targetFile == "" {
		fmt.Println("Error: --target is required")
		os.Exit(1)
	}
	
	client, err := NewClient(*host, *port)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()
	
	fmt.Println("Sending sync request...")
	fmt.Printf("  Source: %s\n", *sourceFile)
	fmt.Printf("  Target: %s\n", *targetFile)
	if *keyColumn != "" {
		fmt.Printf("  Key column: %s\n", *keyColumn)
	} else {
		fmt.Printf("  Key column: (first column)\n")
	}
	fmt.Printf("  Apply changes: %v\n", *apply)
	fmt.Println()
	
	resp, err := client.Sync(*sourceFile, *targetFile, *keyColumn, *apply)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	if resp.Success {
		fmt.Println("Success!")
		fmt.Println(resp.Message)
		
		if resp.Changes != nil && resp.Changes.Total > 0 {
			fmt.Println()
			fmt.Println("Summary:")
			fmt.Printf("  Total changes: %d\n", resp.Changes.Total)
			fmt.Printf("  Added:         %d\n", resp.Changes.Added)
			fmt.Printf("  Modified:      %d\n", resp.Changes.Modified)
			fmt.Printf("  Deleted:       %d\n", resp.Changes.Deleted)
		}
	} else {
		fmt.Printf("Error: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleStatusCommand(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	
	host := fs.String("host", "127.0.0.1", "Server host")
	port := fs.String("port", "8888", "Server port")
	
	fs.Parse(args)
	
	client, err := NewClient(*host, *port)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()
	
	resp, err := client.Status()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	if resp.Success {
		fmt.Println(resp.Message)
	} else {
		fmt.Printf("Error: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleHistoryCommand(args []string) {
	fs := flag.NewFlagSet("history", flag.ExitOnError)
	
	host := fs.String("host", "127.0.0.1", "Server host")
	port := fs.String("port", "8888", "Server port")
	
	fs.Parse(args)
	
	client, err := NewClient(*host, *port)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()
	
	resp, err := client.History()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	if resp.Success {
		fmt.Println(resp.Message)
		fmt.Println()
		
		if len(resp.History) > 0 {
			fmt.Println("History:")
			for i, h := range resp.History {
				fmt.Printf("\n[%d]\n", i+1)
				fmt.Printf("  ID:        %s\n", h.ID)
				fmt.Printf("  Timestamp: %s\n", h.Timestamp)
				fmt.Printf("  Source:    %s\n", h.SourceFile)
				fmt.Printf("  Target:    %s\n", h.TargetFile)
				fmt.Printf("  Changes:   %d total (%d added, %d modified, %d deleted)\n",
					h.Changes.Total, h.Changes.Added, h.Changes.Modified, h.Changes.Deleted)
				fmt.Printf("  Applied:   %v\n", h.Applied)
			}
		}
	} else {
		fmt.Printf("Error: %s\n", resp.Message)
		os.Exit(1)
	}
}
