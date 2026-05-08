package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "server URL")
	expr := flag.String("expr", "", "expression to evaluate")
	filePath := flag.String("file", "", "file containing expressions (one per line)")
	setVar := flag.String("set", "", "set variable (format: name=value)")
	getVar := flag.String("get", "", "get variable value")
	listVars := flag.Bool("list", false, "list all variables")
	interactive := flag.Bool("i", false, "interactive mode")
	flag.Parse()

	client := NewRPCClient(*serverURL)

	if *setVar != "" {
		parts := strings.SplitN(*setVar, "=", 2)
		if len(parts) != 2 {
			fmt.Fprintln(os.Stderr, "invalid variable format, use: name=value")
			os.Exit(1)
		}
		name := parts[0]
		value, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid number: %s\n", parts[1])
			os.Exit(1)
		}
		resp, err := client.SetVariable(name, value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "error: %s\n", resp.Error)
			os.Exit(1)
		}
		fmt.Printf("Set variable: %s = %v\n", name, value)
		return
	}

	if *getVar != "" {
		resp, err := client.GetVariable(*getVar)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "error: %s\n", resp.Error)
			os.Exit(1)
		}
		fmt.Printf("%s = %v\n", *getVar, resp.Value)
		return
	}

	if *listVars {
		resp, err := client.ListVariables()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "error: %s\n", resp.Error)
			os.Exit(1)
		}
		if len(resp.Variables) == 0 {
			fmt.Println("No variables defined.")
		} else {
			fmt.Println("Variables:")
			for k, v := range resp.Variables {
				fmt.Printf("  %s = %v\n", k, v)
			}
		}
		return
	}

	if *expr != "" {
		if err := client.ProcessExpression(*expr); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *filePath != "" {
		file, err := os.Open(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		hasError := false
		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			fmt.Printf("--- Line %d ---\n", lineNum)
			if err := client.ProcessExpression(line); err != nil {
				fmt.Fprintf(os.Stderr, "  error: %v\n", err)
				hasError = true
			}
			fmt.Println()
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
			os.Exit(1)
		}
		if hasError {
			os.Exit(1)
		}
		return
	}

	if *interactive || flag.NArg() == 0 {
		fmt.Println("RPN Evaluator Client - Interactive Mode")
		fmt.Println("Enter expressions to evaluate, or:")
		fmt.Println("  :set name=value  - set variable")
		fmt.Println("  :get name        - get variable")
		fmt.Println("  :list            - list all variables")
		fmt.Println("  :quit or :q      - exit")
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("> ")
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, ":") {
				cmd := strings.TrimPrefix(line, ":")
				parts := strings.Fields(cmd)
				if len(parts) == 0 {
					continue
				}
				switch parts[0] {
				case "quit", "q":
					return
				case "set":
					if len(parts) < 2 {
						fmt.Println("usage: :set name=value")
						continue
					}
					kv := strings.SplitN(parts[1], "=", 2)
					if len(kv) != 2 {
						fmt.Println("invalid format, use: name=value")
						continue
					}
					value, err := strconv.ParseFloat(kv[1], 64)
					if err != nil {
						fmt.Printf("invalid number: %s\n", kv[1])
						continue
					}
					resp, err := client.SetVariable(kv[0], value)
					if err != nil {
						fmt.Printf("error: %v\n", err)
						continue
					}
					if !resp.Success {
						fmt.Printf("error: %s\n", resp.Error)
						continue
					}
					fmt.Printf("Set: %s = %v\n", kv[0], value)
				case "get":
					if len(parts) < 2 {
						fmt.Println("usage: :get name")
						continue
					}
					resp, err := client.GetVariable(parts[1])
					if err != nil {
						fmt.Printf("error: %v\n", err)
						continue
					}
					if !resp.Success {
						fmt.Printf("error: %s\n", resp.Error)
						continue
					}
					fmt.Printf("%s = %v\n", parts[1], resp.Value)
				case "list":
					resp, err := client.ListVariables()
					if err != nil {
						fmt.Printf("error: %v\n", err)
						continue
					}
					if !resp.Success {
						fmt.Printf("error: %s\n", resp.Error)
						continue
					}
					if len(resp.Variables) == 0 {
						fmt.Println("No variables defined.")
					} else {
						fmt.Println("Variables:")
						for k, v := range resp.Variables {
							fmt.Printf("  %s = %v\n", k, v)
						}
					}
				default:
					fmt.Printf("unknown command: %s\n", parts[0])
				}
				continue
			}

			if err := client.ProcessExpression(line); err != nil {
				fmt.Printf("error: %v\n", err)
			}
		}
		return
	}

	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  client -expr \"3 + 4 * 2\"")
	fmt.Fprintln(os.Stderr, "  client -file expressions.txt")
	fmt.Fprintln(os.Stderr, "  client -set \"x=5\"")
	fmt.Fprintln(os.Stderr, "  client -get x")
	fmt.Fprintln(os.Stderr, "  client -list")
	fmt.Fprintln(os.Stderr, "  client -i (interactive mode)")
	os.Exit(1)
}
