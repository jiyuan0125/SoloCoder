package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	baseURL := "http://localhost:8080"
	var command string
	var cmdArgs []string

	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "--url" {
			if i+1 >= len(args) {
				fmt.Println("Error: --url requires an address")
				printUsage()
				os.Exit(1)
			}
			baseURL = args[i+1]
			i += 2
		} else if command == "" {
			command = arg
			cmdArgs = args[i+1:]
			break
		} else {
			cmdArgs = append(cmdArgs, arg)
			i++
		}
	}

	if command == "" {
		printUsage()
		os.Exit(1)
	}

	c := newClient(baseURL)

	var err error
	switch command {
	case "status":
		err = c.Status()

	case "graceful":
		err = c.Graceful()

	case "force":
		err = c.Force()

	case "register":
		if len(cmdArgs) < 2 {
			fmt.Println("Error: register requires <name> and <order>")
			printUsage()
			os.Exit(1)
		}
		name := cmdArgs[0]
		order, parseErr := strconv.Atoi(cmdArgs[1])
		if parseErr != nil {
			fmt.Printf("Error: invalid order '%s': %v\n", cmdArgs[1], parseErr)
			os.Exit(1)
		}
		err = c.Register(name, order)

	default:
		fmt.Printf("Error: unknown command '%s'\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
