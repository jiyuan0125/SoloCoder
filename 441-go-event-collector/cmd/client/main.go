package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"event-collector/client"
	"event-collector/common"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverAddr := flag.String("server", fmt.Sprintf("http://localhost:%d", common.DefaultServerPort), "Server address")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]
	commandArgs := args[1:]

	apiClient := client.NewAPIClient(*serverAddr)

	switch command {
	case "track":
		handleTrack(apiClient, commandArgs)
	case "query":
		handleQuery(apiClient, commandArgs)
	case "aggregate":
		handleAggregate(apiClient, commandArgs)
	case "reports":
		handleReports(apiClient, commandArgs)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleTrack(apiClient *client.APIClient, args []string) {
	if len(args) < 5 {
		fmt.Println("Usage: track <event_name> <timestamp> <properties_json> <user_id> <device_id>")
		os.Exit(1)
	}

	eventName := args[0]
	timestamp, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		timestamp = time.Now().Unix()
	}

	propertiesJSON := args[2]
	userID := args[3]
	deviceID := args[4]

	event := common.Event{
		Name:      eventName,
		Timestamp: timestamp,
		UserID:    userID,
		DeviceID:  deviceID,
	}

	var properties map[string]interface{}
	if propertiesJSON != "" && propertiesJSON != "{}" {
		if err := client.ParseJSON(propertiesJSON, &properties); err != nil {
			fmt.Printf("Error parsing properties JSON: %v\n", err)
			os.Exit(1)
		}
	}
	event.Properties = properties

	fmt.Printf("Tracking event: %s\n", eventName)
	response, err := apiClient.Track(&event)
	if err != nil {
		fmt.Printf("Error tracking event: %v\n", err)
		os.Exit(1)
	}

	client.PrettyPrintJSON(response)
}

func handleQuery(apiClient *client.APIClient, args []string) {
	req := common.QueryRequest{
		Page:     1,
		PageSize: common.DefaultPageSize,
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--event":
			if i+1 < len(args) {
				req.EventName = args[i+1]
				i++
			}
		case "--user":
			if i+1 < len(args) {
				req.UserID = args[i+1]
				i++
			}
		case "--start":
			if i+1 < len(args) {
				startTime, err := strconv.ParseInt(args[i+1], 10, 64)
				if err == nil {
					req.StartTime = startTime
				}
				i++
			}
		case "--end":
			if i+1 < len(args) {
				endTime, err := strconv.ParseInt(args[i+1], 10, 64)
				if err == nil {
					req.EndTime = endTime
				}
				i++
			}
		case "--page":
			if i+1 < len(args) {
				page, err := strconv.Atoi(args[i+1])
				if err == nil {
					req.Page = page
				}
				i++
			}
		case "--page-size":
			if i+1 < len(args) {
				pageSize, err := strconv.Atoi(args[i+1])
				if err == nil {
					req.PageSize = pageSize
				}
				i++
			}
		}
	}

	fmt.Println("Querying events...")
	response, err := apiClient.Query(&req)
	if err != nil {
		fmt.Printf("Error querying events: %v\n", err)
		os.Exit(1)
	}

	client.PrettyPrintJSON(response)
}

func handleAggregate(apiClient *client.APIClient, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: aggregate --events <event1,event2,...> --start <timestamp> --end <timestamp>")
		os.Exit(1)
	}

	req := common.AggregationRequest{
		EventNames: make([]string, 0),
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--events":
			if i+1 < len(args) {
				eventNames := strings.Split(args[i+1], ",")
				for _, name := range eventNames {
					name = strings.TrimSpace(name)
					if name != "" {
						req.EventNames = append(req.EventNames, name)
					}
				}
				i++
			}
		case "--start":
			if i+1 < len(args) {
				startTime, err := strconv.ParseInt(args[i+1], 10, 64)
				if err == nil {
					req.StartTime = startTime
				}
				i++
			}
		case "--end":
			if i+1 < len(args) {
				endTime, err := strconv.ParseInt(args[i+1], 10, 64)
				if err == nil {
					req.EndTime = endTime
				}
				i++
			}
		}
	}

	if len(req.EventNames) == 0 {
		fmt.Println("Error: at least one event name is required")
		os.Exit(1)
	}

	fmt.Println("Aggregating events...")
	response, err := apiClient.Aggregate(&req)
	if err != nil {
		fmt.Printf("Error aggregating events: %v\n", err)
		os.Exit(1)
	}

	client.PrettyPrintJSON(response)
}

func handleReports(apiClient *client.APIClient, args []string) {
	fmt.Println("Fetching daily reports...")
	response, err := apiClient.GetReports()
	if err != nil {
		fmt.Printf("Error fetching reports: %v\n", err)
		os.Exit(1)
	}

	client.PrettyPrintJSON(response)
}

func printUsage() {
	fmt.Println("Event Collector Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client [flags] <command> [arguments]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -server string    Server address (default \"http://localhost:8080\")")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  track     Track an event")
	fmt.Println("  query     Query events")
	fmt.Println("  aggregate Aggregate events for funnel analysis")
	fmt.Println("  reports   Get daily reports")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client track page_view 1234567890 '{\"page\":\"home\"}' user_001 device_001")
	fmt.Println("  client query --event page_view --page 1 --page-size 20")
	fmt.Println("  client aggregate --events view_cart,checkout,purchase --start 1234567890 --end 1234567999")
	fmt.Println("  client reports")
}
