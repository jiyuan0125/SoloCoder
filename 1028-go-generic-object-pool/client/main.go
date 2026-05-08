package main

import (
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Print(`Usage: client <command> [options]

Commands:
  create   Create a new object pool
  get      Get an object from pool
  put      Return an object to pool
  stats    Get pool statistics
  close    Close a pool

Options:
  --server      Server URL (default: http://localhost:8080)
  --pool-id     Pool identifier (required for most commands)
  --max-size    Max pool size (for create command)
  --idle-timeout Idle timeout duration (for create command, e.g., 5m)
  --data        Data to set in object (for put command)
`)
}

func main() {
	server := flag.String("server", "http://localhost:8080", "Server URL")
	poolID := flag.String("pool-id", "", "Pool identifier")
	maxSize := flag.Int("max-size", 100, "Max pool size")
	idleTimeout := flag.String("idle-timeout", "", "Idle timeout duration")
	data := flag.String("data", "", "Object data for put")

	flag.Usage = usage

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	flag.CommandLine.Parse(os.Args[2:])

	c := &Client{BaseURL: *server}

	var err error
	switch cmd {
	case "create":
		if *poolID == "" {
			fmt.Println("error: --pool-id is required")
			os.Exit(1)
		}
		err = c.CreatePool(*poolID, *maxSize, *idleTimeout)

	case "get":
		if *poolID == "" {
			fmt.Println("error: --pool-id is required")
			os.Exit(1)
		}
		err = c.GetObject(*poolID)

	case "put":
		if *poolID == "" {
			fmt.Println("error: --pool-id is required")
			os.Exit(1)
		}
		err = c.PutObject(*poolID, *data)

	case "stats":
		if *poolID == "" {
			fmt.Println("error: --pool-id is required")
			os.Exit(1)
		}
		err = c.GetStats(*poolID)

	case "close":
		if *poolID == "" {
			fmt.Println("error: --pool-id is required")
			os.Exit(1)
		}
		err = c.ClosePool(*poolID)

	default:
		fmt.Printf("error: unknown command '%s'\n", cmd)
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
