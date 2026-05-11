package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"multirate-limiter/common"
)

type Command interface {
	Name() string
	Parse(args []string) error
	Run(c *HTTPClient) error
	Help() string
}

type AllowCommand struct {
	path      string
	clientID  string
	noClient  bool
	anonymous bool
}

func (cmd *AllowCommand) Name() string { return "allow" }

func (cmd *AllowCommand) Parse(args []string) error {
	fs := flag.NewFlagSet("allow", flag.ContinueOnError)
	fs.StringVar(&cmd.path, "path", "", "API path to test")
	fs.StringVar(&cmd.clientID, "client", "", "client ID for per-user limiting")
	fs.BoolVar(&cmd.noClient, "no-client", false, "don't pass client ID (all requests share one bucket)")
	fs.BoolVar(&cmd.anonymous, "anonymous", false, "use empty string client ID (anonymous users share one bucket)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if cmd.path == "" {
		return fmt.Errorf("path is required")
	}

	return nil
}

func (cmd *AllowCommand) Run(c *HTTPClient) error {
	req := map[string]interface{}{
		"path": cmd.path,
	}

	if !cmd.noClient {
		if cmd.anonymous {
			req["client_id"] = ""
		} else if cmd.clientID != "" {
			req["client_id"] = cmd.clientID
		}
	}

	var result common.AllowResponse
	if err := c.do(http.MethodPost, "/allow", req, &result); err != nil {
		return err
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))
	return nil
}

func (cmd *AllowCommand) Help() string {
	return `allow - test rate limiting
Usage: allow --path <path> [--client <id>|--anonymous|--no-client]
Examples:
  allow --path /api/user --no-client
  allow --path /api/user --anonymous
  allow --path /api/user --client user123`
}

type ConfigCommand struct {
	path      string
	get       bool
	set       string
}

func (cmd *ConfigCommand) Name() string { return "config" }

func (cmd *ConfigCommand) Parse(args []string) error {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.StringVar(&cmd.path, "path", "", "API path to configure")
	fs.BoolVar(&cmd.get, "get", false, "get current rules (default)")
	fs.StringVar(&cmd.set, "set", "", "set rules as JSON, e.g. '[{\"granularity\":\"second\",\"quota\":10}]'")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if cmd.path == "" {
		return fmt.Errorf("path is required")
	}

	return nil
}

func (cmd *ConfigCommand) Run(c *HTTPClient) error {
	if cmd.set != "" {
		return cmd.runSet(c)
	}
	return cmd.runGet(c)
}

func (cmd *ConfigCommand) runGet(c *HTTPClient) error {
	req := common.ConfigGetRequest{Path: cmd.path}
	var result common.ConfigGetResponse
	if err := c.do(http.MethodGet, "/config", req, &result); err != nil {
		return err
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))
	return nil
}

func (cmd *ConfigCommand) runSet(c *HTTPClient) error {
	var rules []common.Rule
	if err := json.Unmarshal([]byte(cmd.set), &rules); err != nil {
		return fmt.Errorf("parse rules JSON: %w", err)
	}

	req := common.ConfigSetRequest{Path: cmd.path, Rules: rules}
	var result common.ConfigSetResponse
	if err := c.do(http.MethodPost, "/config", req, &result); err != nil {
		return err
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))
	return nil
}

func (cmd *ConfigCommand) Help() string {
	return `config - view or set rate limiting rules
Usage: config --path <path> [--get|--set <json>]
Examples:
  config --path /api/user --get
  config --path /api/user --set '[{"granularity":"second","quota":10},{"granularity":"minute","quota":100}]'`
}

type StatsCommand struct {
	path      string
	clientID  string
	noClient  bool
	anonymous bool
}

func (cmd *StatsCommand) Name() string { return "stats" }

func (cmd *StatsCommand) Parse(args []string) error {
	fs := flag.NewFlagSet("stats", flag.ContinueOnError)
	fs.StringVar(&cmd.path, "path", "", "API path to view stats")
	fs.StringVar(&cmd.clientID, "client", "", "client ID for per-user stats")
	fs.BoolVar(&cmd.noClient, "no-client", false, "don't pass client ID (all requests share one bucket)")
	fs.BoolVar(&cmd.anonymous, "anonymous", false, "use empty string client ID (anonymous users share one bucket)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if cmd.path == "" {
		return fmt.Errorf("path is required")
	}

	return nil
}

func (cmd *StatsCommand) Run(c *HTTPClient) error {
	req := map[string]interface{}{
		"path": cmd.path,
	}

	if !cmd.noClient {
		if cmd.anonymous {
			req["client_id"] = ""
		} else if cmd.clientID != "" {
			req["client_id"] = cmd.clientID
		}
	}

	var result common.StatsResponse
	if err := c.do(http.MethodGet, "/stats", req, &result); err != nil {
		return err
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))
	return nil
}

func (cmd *StatsCommand) Help() string {
	return `stats - view current usage stats
Usage: stats --path <path> [--client <id>|--anonymous|--no-client]
Examples:
  stats --path /api/user --no-client
  stats --path /api/user --anonymous
  stats --path /api/user --client user123`
}

var commands = []Command{
	&AllowCommand{},
	&ConfigCommand{},
	&StatsCommand{},
}

func findCommand(name string) Command {
	for _, cmd := range commands {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <command> [options]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Commands:\n")
	for _, cmd := range commands {
		lines := strings.Split(cmd.Help(), "\n")
		fmt.Fprintf(os.Stderr, "  %s\n", lines[0])
	}
	fmt.Fprintf(os.Stderr, "\nUse '%s <command> --help' for more info.\n", os.Args[0])
}

func printCommandHelp(cmd Command) {
	fmt.Println(cmd.Help())
}
