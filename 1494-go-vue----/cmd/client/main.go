package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type APIClient struct {
	BaseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	if baseURL == "" {
		baseURL = "http://localhost:9014"
	}
	return &APIClient{BaseURL: baseURL}
}

func (c *APIClient) request(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func printJSON(data []byte) {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err == nil {
		fmt.Println(pretty.String())
	} else {
		fmt.Println(string(data))
	}
}

func main() {
	baseURL := flag.String("server", "http://localhost:9014", "Server base URL")
	flag.Parse()

	if len(flag.Args()) < 1 {
		printUsage()
		os.Exit(1)
	}

	client := NewAPIClient(*baseURL)
	command := flag.Args()[0]

	var err error
	switch command {
	case "zone":
		err = handleZone(client, flag.Args()[1:])
	case "plant":
		err = handlePlant(client, flag.Args()[1:])
	case "worker":
		err = handleWorker(client, flag.Args()[1:])
	case "plan":
		err = handlePlan(client, flag.Args()[1:])
	case "task":
		err = handleTask(client, flag.Args()[1:])
	case "report":
		err = handleReport(client, flag.Args()[1:])
	default:
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: green-care [options] <command> [subcommand] [args]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -server string    Server base URL (default: http://localhost:9014)")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  zone list|create|get <id>")
	fmt.Println("  plant list|create|get <id>|update <id>|delete <id>")
	fmt.Println("  worker list|create|get <id>|update <id>")
	fmt.Println("  plan list|create|get <id>")
	fmt.Println("  task list|create|get <id>|generate|execute|review")
	fmt.Println("  report <year> <month>")
}

func handleZone(c *APIClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("zone command requires subcommand: list|create|get")
	}

	sub := args[0]
	switch sub {
	case "list":
		resp, err := c.request("GET", "/api/zones", nil)
		if err != nil {
			return err
		}
		printJSON(resp)
	case "create":
		if len(args) < 2 {
			return fmt.Errorf("usage: zone create <name> [description]")
		}
		name := args[1]
		desc := ""
		if len(args) > 2 {
			desc = strings.Join(args[2:], " ")
		}
		body := map[string]string{"name": name, "description": desc}
		resp, err := c.request("POST", "/api/zones", body)
		if err != nil {
			return err
		}
		printJSON(resp)
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("usage: zone get <id>")
		}
		resp, err := c.request("GET", "/api/zones/"+args[1], nil)
		if err != nil {
			return err
		}
		printJSON(resp)
	default:
		return fmt.Errorf("unknown zone subcommand: %s", sub)
	}
	return nil
}

func handlePlant(c *APIClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("plant command requires subcommand")
	}

	sub := args[0]
	switch sub {
	case "list":
		path := "/api/plants"
		if len(args) > 1 {
			path += "?zone_id=" + args[1]
		}
		resp, err := c.request("GET", path, nil)
		if err != nil {
			return err
		}
		printJSON(resp)
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("usage: plant get <id>")
		}
		resp, err := c.request("GET", "/api/plants/"+args[1], nil)
		if err != nil {
			return err
		}
		printJSON(resp)
	case "delete":
		if len(args) < 2 {
			return fmt.Errorf("usage: plant delete <id>")
		}
		resp, err := c.request("DELETE", "/api/plants/"+args[1], nil)
		if err != nil {
			return err
		}
		printJSON(resp)
	default:
		return fmt.Errorf("plant subcommands: list, get <id>, delete <id>")
	}
	return nil
}

func handleWorker(c *APIClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("worker command requires subcommand")
	}

	sub := args[0]
	switch sub {
	case "list":
		path := "/api/workers"
		if len(args) > 1 {
			path += "?zone_id=" + args[1]
		}
		resp, err := c.request("GET", path, nil)
		if err != nil {
			return err
		}
		printJSON(resp)
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("usage: worker get <id>")
		}
		resp, err := c.request("GET", "/api/workers/"+args[1], nil)
		if err != nil {
			return err
		}
		printJSON(resp)
	default:
		return fmt.Errorf("worker subcommands: list, get <id>")
	}
	return nil
}

func handlePlan(c *APIClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("plan command requires subcommand")
	}

	sub := args[0]
	switch sub {
	case "list":
		resp, err := c.request("GET", "/api/maintenance-plans", nil)
		if err != nil {
			return err
		}
		printJSON(resp)
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("usage: plan get <id>")
		}
		resp, err := c.request("GET", "/api/maintenance-plans/"+args[1], nil)
		if err != nil {
			return err
		}
		printJSON(resp)
	default:
		return fmt.Errorf("plan subcommands: list, get <id>")
	}
	return nil
}

func handleTask(c *APIClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("task command requires subcommand")
	}

	sub := args[0]
	switch sub {
	case "list":
		path := "/api/tasks"
		if len(args) > 1 {
			if strings.HasPrefix(args[1], "zone=") {
				path += "?zone_id=" + strings.TrimPrefix(args[1], "zone=")
			} else if strings.HasPrefix(args[1], "worker=") {
				path += "?worker_id=" + strings.TrimPrefix(args[1], "worker=")
			}
		}
		resp, err := c.request("GET", path, nil)
		if err != nil {
			return err
		}
		printJSON(resp)
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("usage: task get <id>")
		}
		resp, err := c.request("GET", "/api/tasks/"+args[1], nil)
		if err != nil {
			return err
		}
		printJSON(resp)
	case "generate":
		body := map[string]interface{}{}
		resp, err := c.request("POST", "/api/tasks/generate", body)
		if err != nil {
			return err
		}
		printJSON(resp)
	default:
		return fmt.Errorf("task subcommands: list, get <id>, generate")
	}
	return nil
}

func handleReport(c *APIClient, args []string) error {
	var year, month int
	if len(args) >= 2 {
		fmt.Sscanf(args[0], "%d", &year)
		fmt.Sscanf(args[1], "%d", &month)
	}
	body := map[string]int{"year": year, "month": month}
	resp, err := c.request("POST", "/api/cost-report", body)
	if err != nil {
		return err
	}
	printJSON(resp)
	return nil
}
