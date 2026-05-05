package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"env-switcher/protocol"
)

const (
	defaultPort = "45678"
	portEnvVar  = "ENV_SWITCHER_PORT"
)

func getPort() string {
	if port := os.Getenv(portEnvVar); port != "" {
		return port
	}
	return defaultPort
}

func printUsage() {
	fmt.Println("Usage: env-switcher <command> [arguments]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  switch <env>     Switch to specified environment")
	fmt.Println("  list             List all available environments")
	fmt.Println("  show             Show current environment variables")
	fmt.Println("  export           Export current environment as shell commands")
	fmt.Println("  add <env>        Add a new environment")
	fmt.Println("  del <env>        Delete an existing environment")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  env-switcher switch dev")
	fmt.Println("  env-switcher list")
	fmt.Println("  env-switcher show")
	fmt.Println("  env-switcher export")
}

func ensureServerRunning() {
	port := getPort()
	
	client := NewClient("localhost:" + port)
	_, err := client.SendRequest(&protocol.Request{Cmd: protocol.CmdList})
	if err == nil {
		return
	}
	
	serverCmd := exec.Command("env-switcher-server")
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	err = serverCmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	
	command := os.Args[1]
	
	ensureServerRunning()
	
	port := getPort()
	client := NewClient("localhost:" + port)
	
	var req *protocol.Request
	
	switch command {
	case "switch":
		if len(os.Args) < 3 {
			fmt.Println("Usage: env-switcher switch <environment>")
			os.Exit(1)
		}
		req = &protocol.Request{
			Cmd: protocol.CmdSwitch,
			Env: os.Args[2],
		}
		
	case "list":
		req = &protocol.Request{Cmd: protocol.CmdList}
		
	case "show":
		req = &protocol.Request{Cmd: protocol.CmdShow}
		
	case "export":
		req = &protocol.Request{Cmd: protocol.CmdExport}
		
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: env-switcher add <environment>")
			os.Exit(1)
		}
		vars := make(map[string]string)
		if len(os.Args) > 3 {
			for _, arg := range os.Args[3:] {
				parts := strings.SplitN(arg, "=", 2)
				if len(parts) == 2 {
					vars[parts[0]] = parts[1]
				}
			}
		}
		req = &protocol.Request{
			Cmd:  protocol.CmdAdd,
			Env:  os.Args[2],
			Vars: vars,
		}
		
	case "del":
		if len(os.Args) < 3 {
			fmt.Println("Usage: env-switcher del <environment>")
			os.Exit(1)
		}
		req = &protocol.Request{
			Cmd: protocol.CmdDel,
			Env: os.Args[2],
		}
		
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
	
	resp, err := client.SendRequest(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	printer := NewResponsePrinter()
	printer.Print(command, resp)
}
