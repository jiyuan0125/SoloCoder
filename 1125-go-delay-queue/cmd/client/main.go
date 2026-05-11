package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"delayqueue/pkg/api"
)

const defaultEndpoint = "http://localhost:8080"

type Command interface {
	Name() string
	Run(args []string) error
	Usage() string
}

type SubmitCmd struct {
	endpoint string
}

func (c *SubmitCmd) Name() string { return "submit" }

func (c *SubmitCmd) Run(args []string) error {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.StringVar(&c.endpoint, "endpoint", defaultEndpoint, "server endpoint")
	id := fs.String("id", "", "task id")
	payload := fs.String("payload", "", "task payload")
	delay := fs.String("delay", "", "delay duration (e.g. 5s, 1m30s)")
	at := fs.String("at", "", "execute at RFC3339 time")
	priority := fs.String("priority", "normal", "priority: low|normal|high|urgent")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == "" {
		return fmt.Errorf("id is required")
	}

	var executeAt time.Time
	if *delay != "" {
		d, err := time.ParseDuration(*delay)
		if err != nil {
			return fmt.Errorf("invalid delay: %w", err)
		}
		executeAt = time.Now().Add(d)
	} else if *at != "" {
		var err error
		executeAt, err = time.Parse(time.RFC3339, *at)
		if err != nil {
			return fmt.Errorf("invalid at time: %w", err)
		}
	} else {
		return fmt.Errorf("delay or at is required")
	}

	p, err := parsePriority(*priority)
	if err != nil {
		return err
	}

	req := api.SubmitRequest{
		ID:        *id,
		Payload:   *payload,
		ExecuteAt: executeAt,
		Priority:  p,
	}

	var resp api.SubmitResponse
	if err := postJSON(c.endpoint+"/submit", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("submit failed: %s", resp.Error)
	}

	fmt.Printf("Submitted task %s, execute at %s (priority %s)\n", *id, executeAt.Format(time.RFC3339), *priority)
	return nil
}

func (c *SubmitCmd) Usage() string {
	return "submit -id <id> [-payload <data>] -delay <duration>|-at <rfc3339> [-priority low|normal|high|urgent]"
}

type ModifyCmd struct {
	endpoint string
}

func (c *ModifyCmd) Name() string { return "modify" }

func (c *ModifyCmd) Run(args []string) error {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.StringVar(&c.endpoint, "endpoint", defaultEndpoint, "server endpoint")
	id := fs.String("id", "", "task id")
	delay := fs.String("delay", "", "new delay duration")
	at := fs.String("at", "", "new execute at RFC3339 time")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == "" {
		return fmt.Errorf("id is required")
	}

	var executeAt time.Time
	if *delay != "" {
		d, err := time.ParseDuration(*delay)
		if err != nil {
			return fmt.Errorf("invalid delay: %w", err)
		}
		executeAt = time.Now().Add(d)
	} else if *at != "" {
		var err error
		executeAt, err = time.Parse(time.RFC3339, *at)
		if err != nil {
			return fmt.Errorf("invalid at time: %w", err)
		}
	} else {
		return fmt.Errorf("delay or at is required")
	}

	req := api.ModifyRequest{
		ID:        *id,
		ExecuteAt: executeAt,
	}

	var resp api.ModifyResponse
	if err := postJSON(c.endpoint+"/modify", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("modify failed: %s", resp.Error)
	}

	fmt.Printf("Modified task %s, new execute at %s\n", *id, executeAt.Format(time.RFC3339))
	return nil
}

func (c *ModifyCmd) Usage() string {
	return "modify -id <id> -delay <duration>|-at <rfc3339>"
}

type PromoteCmd struct {
	endpoint string
}

func (c *PromoteCmd) Name() string { return "promote" }

func (c *PromoteCmd) Run(args []string) error {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.StringVar(&c.endpoint, "endpoint", defaultEndpoint, "server endpoint")
	id := fs.String("id", "", "task id")
	priority := fs.String("priority", "", "new priority: low|normal|high|urgent")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == "" {
		return fmt.Errorf("id is required")
	}
	if *priority == "" {
		return fmt.Errorf("priority is required")
	}

	p, err := parsePriority(*priority)
	if err != nil {
		return err
	}

	req := api.PromoteRequest{
		ID:       *id,
		Priority: p,
	}

	var resp api.PromoteResponse
	if err := postJSON(c.endpoint+"/promote", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("promote failed: %s", resp.Error)
	}

	fmt.Printf("Promoted task %s to priority %s\n", *id, *priority)
	return nil
}

func (c *PromoteCmd) Usage() string {
	return "promote -id <id> -priority low|normal|high|urgent"
}

type PollCmd struct {
	endpoint string
}

func (c *PollCmd) Name() string { return "poll" }

func (c *PollCmd) Run(args []string) error {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.StringVar(&c.endpoint, "endpoint", defaultEndpoint, "server endpoint")

	if err := fs.Parse(args); err != nil {
		return err
	}

	var resp api.PollResponse
	if err := getJSON(c.endpoint+"/poll", &resp); err != nil {
		return err
	}

	if resp.Task == nil {
		fmt.Println("No expired tasks")
		return nil
	}

	printTask(resp.Task)
	return nil
}

func (c *PollCmd) Usage() string { return "poll" }

type PeekCmd struct {
	endpoint string
}

func (c *PeekCmd) Name() string { return "peek" }

func (c *PeekCmd) Run(args []string) error {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.StringVar(&c.endpoint, "endpoint", defaultEndpoint, "server endpoint")

	if err := fs.Parse(args); err != nil {
		return err
	}

	var resp api.PeekResponse
	if err := getJSON(c.endpoint+"/peek", &resp); err != nil {
		return err
	}

	if resp.Task == nil {
		fmt.Println("Queue is empty")
		return nil
	}

	fmt.Println("Peek (next task, not removed):")
	printTask(resp.Task)
	return nil
}

func (c *PeekCmd) Usage() string { return "peek" }

type DrainCmd struct {
	endpoint string
}

func (c *DrainCmd) Name() string { return "drain" }

func (c *DrainCmd) Run(args []string) error {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.StringVar(&c.endpoint, "endpoint", defaultEndpoint, "server endpoint")

	if err := fs.Parse(args); err != nil {
		return err
	}

	var resp api.DrainResponse
	if err := getJSON(c.endpoint+"/drain", &resp); err != nil {
		return err
	}

	if len(resp.Tasks) == 0 {
		fmt.Println("No expired tasks")
		return nil
	}

	fmt.Printf("Drained %d expired task(s):\n", len(resp.Tasks))
	for _, t := range resp.Tasks {
		printTask(t)
	}
	return nil
}

func (c *DrainCmd) Usage() string { return "drain" }

type StatusCmd struct {
	endpoint string
}

func (c *StatusCmd) Name() string { return "status" }

func (c *StatusCmd) Run(args []string) error {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.StringVar(&c.endpoint, "endpoint", defaultEndpoint, "server endpoint")

	if err := fs.Parse(args); err != nil {
		return err
	}

	var resp api.StatusResponse
	if err := getJSON(c.endpoint+"/status", &resp); err != nil {
		return err
	}

	fmt.Printf("Queued tasks: %d\n", resp.Queued)
	return nil
}

func (c *StatusCmd) Usage() string { return "status" }

func parsePriority(s string) (api.Priority, error) {
	switch strings.ToLower(s) {
	case "low":
		return api.PriorityLow, nil
	case "normal":
		return api.PriorityNormal, nil
	case "high":
		return api.PriorityHigh, nil
	case "urgent":
		return api.PriorityUrgent, nil
	}
	return 0, fmt.Errorf("invalid priority: %s", s)
}

func priorityString(p api.Priority) string {
	switch p {
	case api.PriorityLow:
		return "low"
	case api.PriorityNormal:
		return "normal"
	case api.PriorityHigh:
		return "high"
	case api.PriorityUrgent:
		return "urgent"
	}
	return strconv.Itoa(int(p))
}

func printTask(t *api.Task) {
	fmt.Printf("  ID:       %s\n", t.ID)
	fmt.Printf("  Payload:  %s\n", t.Payload)
	fmt.Printf("  ExecuteAt: %s\n", t.ExecuteAt.Format(time.RFC3339))
	fmt.Printf("  Priority: %s\n", priorityString(t.Priority))
}

func postJSON(url string, req, resp any) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpResp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	data, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode >= 400 {
		return fmt.Errorf("server error %d: %s", httpResp.StatusCode, string(data))
	}

	return json.Unmarshal(data, resp)
}

func getJSON(url string, resp any) error {
	httpResp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	data, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode >= 400 {
		return fmt.Errorf("server error %d: %s", httpResp.StatusCode, string(data))
	}

	return json.Unmarshal(data, resp)
}

func main() {
	commands := []Command{
		&SubmitCmd{},
		&ModifyCmd{},
		&PromoteCmd{},
		&PollCmd{},
		&PeekCmd{},
		&DrainCmd{},
		&StatusCmd{},
	}

	if len(os.Args) < 2 {
		printUsage(commands)
		os.Exit(1)
	}

	cmdName := os.Args[1]
	for _, cmd := range commands {
		if cmd.Name() == cmdName {
			if err := cmd.Run(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
				fmt.Fprintf(os.Stderr, "Usage: %s %s\n", os.Args[0], cmd.Usage())
				os.Exit(1)
			}
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmdName)
	printUsage(commands)
	os.Exit(1)
}

func printUsage(commands []Command) {
	fmt.Println("Delay Queue Client")
	fmt.Println("\nUsage:")
	fmt.Printf("  %s <command> [options]\n\n", os.Args[0])
	fmt.Println("Commands:")
	for _, cmd := range commands {
		fmt.Printf("  %s\n", cmd.Usage())
	}
}
