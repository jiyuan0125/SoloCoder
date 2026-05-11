package main

import (
	"encoding/json"
	"os"
	"strings"

	"safetymanager/internal/api"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		printUsage()
	}

	baseURL := os.Getenv("SAFETY_MANAGER_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := NewClient(baseURL)

	var err error
	switch args[0] {
	case "zone":
		err = handleZoneCommands(client, args[1:])
	case "plan":
		err = handlePlanCommands(client, args[1:])
	case "task":
		err = handleTaskCommands(client, args[1:])
	case "inspection":
		err = handleInspectionCommands(client, args[1:])
	case "hazard":
		err = handleHazardCommands(client, args[1:])
	case "audit":
		err = handleAuditCommands(client, args[1:])
	default:
		printUsage()
	}

	if err != nil {
		os.Stderr.WriteString("Error: " + err.Error() + "\n")
		os.Exit(1)
	}
}

func handleZoneCommands(client *Client, args []string) error {
	if len(args) < 1 {
		printUsage()
	}

	switch args[0] {
	case "create":
		if len(args) < 2 {
			printUsage()
		}
		var parentID *string
		if len(args) >= 3 {
			parentID = &args[2]
		}
		return client.CreateZone(args[1], parentID)
	case "list":
		return client.ListZones()
	default:
		printUsage()
		return nil
	}
}

func handlePlanCommands(client *Client, args []string) error {
	if len(args) < 1 {
		printUsage()
	}

	switch args[0] {
	case "create":
		if len(args) < 6 {
			printUsage()
		}
		inspectors := strings.Split(args[3], ",")
		items := strings.Split(args[4], ",")
		freq := api.Frequency(args[5])
		return client.CreatePlan(args[1], args[2], inspectors, items, freq)
	case "list":
		return client.ListPlans()
	default:
		printUsage()
		return nil
	}
}

func handleTaskCommands(client *Client, args []string) error {
	if len(args) < 1 {
		printUsage()
	}

	switch args[0] {
	case "generate":
		return client.GenerateTasks()
	case "list":
		return client.ListTasks()
	default:
		printUsage()
		return nil
	}
}

func handleInspectionCommands(client *Client, args []string) error {
	if len(args) < 1 {
		printUsage()
	}

	switch args[0] {
	case "submit":
		if len(args) < 3 {
			printUsage()
		}
		var items []api.SubmitInspectionItem
		if err := json.Unmarshal([]byte(args[2]), &items); err != nil {
			return err
		}
		return client.SubmitInspection(args[1], items)
	default:
		printUsage()
		return nil
	}
}

func handleHazardCommands(client *Client, args []string) error {
	if len(args) < 1 {
		printUsage()
	}

	switch args[0] {
	case "list":
		return client.ListHazards()
	case "remediate":
		if len(args) < 3 {
			printUsage()
		}
		return client.SubmitRemediation(args[1], args[2])
	case "review":
		if len(args) < 3 {
			printUsage()
		}
		approved := strings.ToLower(args[2]) == "true" || args[2] == "1"
		return client.ReviewRemediation(args[1], approved)
	case "level-change":
		if len(args) < 2 {
			printUsage()
		}
		switch args[1] {
		case "request":
			if len(args) < 5 {
				printUsage()
			}
			level := api.HazardLevel(args[3])
			return client.RequestLevelChange(args[2], level, args[4])
		case "review":
			if len(args) < 4 {
				printUsage()
			}
			approved := strings.ToLower(args[3]) == "true" || args[3] == "1"
			return client.ReviewLevelChange(args[2], approved)
		default:
			printUsage()
			return nil
		}
	case "escalations":
		if len(args) < 2 || args[1] != "check" {
			printUsage()
		}
		return client.CheckEscalations()
	default:
		printUsage()
		return nil
	}
}

func handleAuditCommands(client *Client, args []string) error {
	if len(args) < 1 {
		printUsage()
	}

	switch args[0] {
	case "list":
		return client.ListAuditLogs()
	case "export-critical":
		return client.ExportCriticalLogs()
	default:
		printUsage()
		return nil
	}
}
