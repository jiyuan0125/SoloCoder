package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"circuit-monitor/common"
)

type Command interface {
	Name() string
	Description() string
	Usage() string
	Execute(args []string, client *APIClient) error
}

type ListCommand struct{}

func (c *ListCommand) Name() string        { return "list" }
func (c *ListCommand) Description() string { return "List all circuit breakers" }
func (c *ListCommand) Usage() string        { return "list" }

func (c *ListCommand) Execute(args []string, client *APIClient) error {
	resp, err := client.ListCircuits()
	if err != nil {
		return err
	}

	fmt.Println("Circuit breakers:")
	for _, name := range resp.Circuits {
		fmt.Printf("  - %s\n", name)
	}
	if len(resp.Circuits) == 0 {
		fmt.Println("  (none)")
	}
	return nil
}

type GetCommand struct{}

func (c *GetCommand) Name() string        { return "get" }
func (c *GetCommand) Description() string { return "Get circuit breaker information" }
func (c *GetCommand) Usage() string        { return "get <name>" }

func (c *GetCommand) Execute(args []string, client *APIClient) error {
	if len(args) < 1 {
		return fmt.Errorf("circuit name is required")
	}

	resp, err := client.GetCircuit(args[0])
	if err != nil {
		return err
	}

	info := resp.CircuitInfo
	fmt.Printf("Circuit: %s\n", info.Name)
	fmt.Printf("  State: %s\n", info.State)
	fmt.Printf("  Total Requests: %d\n", info.TotalRequests)
	fmt.Printf("  Success Count: %d\n", info.SuccessCount)
	fmt.Printf("  Failure Count: %d\n", info.FailureCount)
	fmt.Printf("  Recent Failures: %d\n", info.RecentFailures)
	fmt.Printf("  Recent Successes: %d\n", info.RecentSuccesses)
	if info.OpenAt != nil {
		fmt.Printf("  Open Since: %s\n", info.OpenAt.Format(time.RFC3339))
	}
	if info.LastStateChange != nil {
		fmt.Printf("  Last State Change: %s\n", info.LastStateChange.Format(time.RFC3339))
	}
	return nil
}

type CreateCommand struct {
	windowSize          int
	failureThreshold    int
	openTimeout         int
	halfOpenMaxRequests int
	successThreshold    int
}

func (c *CreateCommand) Name() string        { return "create" }
func (c *CreateCommand) Description() string { return "Create a new circuit breaker" }
func (c *CreateCommand) Usage() string {
	return `create <name> [options]
  Options:
    -window-size N        Sliding window size (default: 10)
    -failure-threshold N  Number of failures to trigger open (default: 6)
    -open-timeout N       Seconds to stay open before half-open (default: 30)
    -half-open-requests N Number of test requests in half-open (default: 3)
    -success-threshold N  Consecutive successes to reset window (default: 10)`
}

func (c *CreateCommand) Execute(args []string, client *APIClient) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	fs.IntVar(&c.windowSize, "window-size", 10, "Sliding window size")
	fs.IntVar(&c.failureThreshold, "failure-threshold", 6, "Failure threshold")
	fs.IntVar(&c.openTimeout, "open-timeout", 30, "Open timeout in seconds")
	fs.IntVar(&c.halfOpenMaxRequests, "half-open-requests", 3, "Half-open test requests")
	fs.IntVar(&c.successThreshold, "success-threshold", 10, "Success threshold")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("circuit name is required")
	}

	name := fs.Arg(0)
	var config *common.CircuitConfig

	if c.windowSize != 10 || c.failureThreshold != 6 || c.openTimeout != 30 ||
		c.halfOpenMaxRequests != 3 || c.successThreshold != 10 {
		config = &common.CircuitConfig{
			WindowSize:          c.windowSize,
			FailureThreshold:    c.failureThreshold,
			OpenTimeout:         time.Duration(c.openTimeout) * time.Second,
			HalfOpenMaxRequests: c.halfOpenMaxRequests,
			SuccessThreshold:    c.successThreshold,
		}
	}

	if err := client.CreateCircuit(name, config); err != nil {
		return err
	}

	fmt.Printf("Circuit breaker '%s' created successfully\n", name)
	return nil
}

type ResetCommand struct{}

func (c *ResetCommand) Name() string        { return "reset" }
func (c *ResetCommand) Description() string { return "Reset a circuit breaker to closed state" }
func (c *ResetCommand) Usage() string        { return "reset <name>" }

func (c *ResetCommand) Execute(args []string, client *APIClient) error {
	if len(args) < 1 {
		return fmt.Errorf("circuit name is required")
	}

	if err := client.ResetCircuit(args[0]); err != nil {
		return err
	}

	fmt.Printf("Circuit breaker '%s' reset successfully\n", args[0])
	return nil
}

type ForceStateCommand struct{}

func (c *ForceStateCommand) Name() string        { return "force-state" }
func (c *ForceStateCommand) Description() string { return "Force a circuit breaker to a specific state" }
func (c *ForceStateCommand) Usage() string        { return "force-state <name> <closed|open|half-open>" }

func (c *ForceStateCommand) Execute(args []string, client *APIClient) error {
	if len(args) < 2 {
		return fmt.Errorf("circuit name and state are required")
	}

	name := args[0]
	state := common.CircuitState(args[1])

	switch state {
	case common.StateClosed, common.StateOpen, common.StateHalfOpen:
	default:
		return fmt.Errorf("invalid state: %s (must be closed, open, or half-open)", state)
	}

	if err := client.ForceState(name, state); err != nil {
		return err
	}

	fmt.Printf("Circuit breaker '%s' forced to state '%s'\n", name, state)
	return nil
}

type GetConfigCommand struct{}

func (c *GetConfigCommand) Name() string        { return "config" }
func (c *GetConfigCommand) Description() string { return "Get circuit breaker configuration" }
func (c *GetConfigCommand) Usage() string        { return "config <name>" }

func (c *GetConfigCommand) Execute(args []string, client *APIClient) error {
	if len(args) < 1 {
		return fmt.Errorf("circuit name is required")
	}

	resp, err := client.GetConfig(args[0])
	if err != nil {
		return err
	}

	cfg := resp.Config
	fmt.Printf("Config for '%s':\n", args[0])
	fmt.Printf("  Window Size: %d\n", cfg.WindowSize)
	fmt.Printf("  Failure Threshold: %d\n", cfg.FailureThreshold)
	fmt.Printf("  Open Timeout: %v\n", cfg.OpenTimeout)
	fmt.Printf("  Half-open Max Requests: %d\n", cfg.HalfOpenMaxRequests)
	fmt.Printf("  Success Threshold: %d\n", cfg.SuccessThreshold)
	return nil
}

type SaveCommand struct{}

func (c *SaveCommand) Name() string        { return "save" }
func (c *SaveCommand) Description() string { return "Save circuit breaker state to file" }
func (c *SaveCommand) Usage() string        { return "save <name>" }

func (c *SaveCommand) Execute(args []string, client *APIClient) error {
	if len(args) < 1 {
		return fmt.Errorf("circuit name is required")
	}

	if err := client.SaveState(args[0]); err != nil {
		return err
	}

	fmt.Printf("Circuit breaker '%s' state saved\n", args[0])
	return nil
}

type LoadCommand struct{}

func (c *LoadCommand) Name() string        { return "load" }
func (c *LoadCommand) Description() string { return "Load circuit breaker state from file" }
func (c *LoadCommand) Usage() string        { return "load <name>" }

func (c *LoadCommand) Execute(args []string, client *APIClient) error {
	if len(args) < 1 {
		return fmt.Errorf("circuit name is required")
	}

	if err := client.LoadState(args[0]); err != nil {
		return err
	}

	fmt.Printf("Circuit breaker '%s' state loaded\n", args[0])
	return nil
}

type HealthCommand struct{}

func (c *HealthCommand) Name() string        { return "health" }
func (c *HealthCommand) Description() string { return "Check server health" }
func (c *HealthCommand) Usage() string        { return "health" }

func (c *HealthCommand) Execute(args []string, client *APIClient) error {
	if err := client.HealthCheck(); err != nil {
		return err
	}
	fmt.Println("Server is healthy")
	return nil
}

type HelpCommand struct {
	commands []Command
}

func (c *HelpCommand) Name() string        { return "help" }
func (c *HelpCommand) Description() string { return "Show help information" }
func (c *HelpCommand) Usage() string        { return "help [command]" }

func (c *HelpCommand) Execute(args []string, client *APIClient) error {
	if len(args) == 0 {
		c.printMainHelp()
		return nil
	}

	cmdName := args[0]
	for _, cmd := range c.commands {
		if cmd.Name() == cmdName {
			fmt.Printf("Usage: %s\n\n", cmd.Usage())
			fmt.Printf("%s\n", cmd.Description())
			return nil
		}
	}

	return fmt.Errorf("unknown command: %s", cmdName)
}

func (c *HelpCommand) printMainHelp() {
	fmt.Println("Circuit Breaker Client")
	fmt.Println()
	fmt.Println("Usage: client [global-options] <command> [args]")
	fmt.Println()
	fmt.Println("Global Options:")
	fmt.Println("  -server URL    Server URL (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("Commands:")
	for _, cmd := range c.commands {
		fmt.Printf("  %-12s %s\n", cmd.Name(), cmd.Description())
	}
	fmt.Println()
	fmt.Println("Use 'client help <command>' for more information about a command")
}

func AllCommands() []Command {
	commands := []Command{
		&ListCommand{},
		&GetCommand{},
		&CreateCommand{},
		&ResetCommand{},
		&ForceStateCommand{},
		&GetConfigCommand{},
		&SaveCommand{},
		&LoadCommand{},
		&HealthCommand{},
	}

	helpCmd := &HelpCommand{commands: commands}
	return append(commands, helpCmd)
}

func Run() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		helpCmd := &HelpCommand{commands: AllCommands()}
		helpCmd.Execute(nil, nil)
		os.Exit(0)
	}

	cmdName := args[0]
	commands := AllCommands()

	for _, cmd := range commands {
		if cmd.Name() == cmdName {
			client := NewAPIClient(*serverURL)
			if err := cmd.Execute(args[1:], client); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", cmdName)
	fmt.Fprintln(os.Stderr, "Use 'client help' to see available commands")
	os.Exit(1)
}
