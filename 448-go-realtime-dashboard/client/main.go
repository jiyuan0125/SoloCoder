package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"realtime-dashboard/common"
)

var (
	serverURL = flag.String("server", "http://localhost:8080", "Server URL")
	clientID  = flag.String("client-id", "", "Client ID (auto-generated if not provided)")
)

func main() {
	flag.Parse()

	if *clientID == "" {
		*clientID = common.GenerateClientID()
	}

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	switch command {
	case "status":
		cmdStatus()
	case "metrics":
		cmdMetrics()
	case "create-metric":
		cmdCreateMetric(args[1:])
	case "update-metric":
		cmdUpdateMetric(args[1:])
	case "get-metric":
		cmdGetMetric(args[1:])
	case "set-threshold":
		cmdSetThreshold(args[1:])
	case "history":
		cmdHistory(args[1:])
	case "trend":
		cmdTrend(args[1:])
	case "comparison":
		cmdComparison(args[1:])
	case "subscribe":
		cmdSubscribe(args[1:])
	case "alerts":
		cmdAlerts()
	case "delay-stats":
		cmdDelayStats()
	case "save-layout":
		cmdSaveLayout(args[1:])
	case "get-layout":
		cmdGetLayout(args[1:])
	case "cleanup":
		cmdCleanup()
	case "interactive":
		cmdInteractive()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`
Usage: client [options] <command> [args]

Options:
  -server URL    Server URL (default: http://localhost:8080)
  -client-id ID  Client ID (auto-generated if not provided)

Commands:
  status [--details]           Get server status
  metrics                      List all metrics
  create-metric <key> <name> [unit] [description] [formula]
                               Create a new metric
  update-metric <key> <value> [timestamp]
                               Update metric value
  get-metric <key>             Get metric details
  set-threshold <key> [--max=<value>] [--min=<value>]
                               Set alert threshold for metric
  history <key> [--start=<time>] [--end=<time>]
                               Get metric history
  trend <key>                  Get metric trend (last hour, 5min intervals)
  comparison <key>             Get YoY/MoM comparison
  subscribe <keys...> [--rate=<seconds>]
                               Subscribe to metrics and listen for updates
  alerts                       Get pending alerts
  delay-stats                  Get push delay statistics
  save-layout <metrics...>     Save dashboard layout
  get-layout                   Get dashboard layout
  cleanup                      Trigger old data cleanup
  interactive                  Start interactive mode
`)
}

func cmdStatus() {
	api := NewAPIClient(*serverURL)

	includeDetails := false
	args := flag.Args()
	if len(args) > 1 && args[1] == "--details" {
		includeDetails = true
	}

	status, err := api.GetServerStatus(includeDetails)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Server Status:\n")
	fmt.Printf("  Online Clients: %d/%d\n", status.OnlineClients, status.MaxClients)
	fmt.Printf("  Queue Size: %d\n", status.QueueSize)
	fmt.Printf("  Metrics Count: %d\n", status.MetricsCount)

	if includeDetails && status.LatestValues != nil {
		fmt.Printf("\nLatest Values:\n")
		for _, mv := range status.LatestValues {
			expired := ""
			if mv.IsExpired {
				expired = " [EXPIRED]"
			}
			fmt.Printf("  %s (%s): %.2f %s%s\n",
				mv.Key, mv.Name, mv.Value, mv.Unit, expired)
			fmt.Printf("    Updated: %s\n", mv.Timestamp.Format(time.RFC3339))
		}
	}

	if includeDetails && status.Clients != nil {
		fmt.Printf("\nOnline Clients:\n")
		for _, c := range status.Clients {
			fmt.Printf("  Client: %s\n", c.ClientID)
			fmt.Printf("    Connected: %s\n", c.ConnectedAt.Format(time.RFC3339))
			fmt.Printf("    Subscribed: %v\n", c.SubscribedMetrics)
			fmt.Printf("    Refresh Rate: %ds\n", c.RefreshRate)
		}
	}
}

func cmdMetrics() {
	api := NewAPIClient(*serverURL)

	metrics, err := api.GetAllMetrics()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(metrics) == 0 {
		fmt.Println("No metrics found")
		return
	}

	fmt.Printf("Metrics (%d):\n", len(metrics))
	for _, m := range metrics {
		expired := ""
		if m.IsExpired {
			expired = " [EXPIRED]"
		}
		fmt.Printf("\n  %s: %s%s\n", m.Key, m.Metadata.Name, expired)
		fmt.Printf("    Value: %.2f %s\n", m.Value, m.Metadata.Unit)
		fmt.Printf("    Updated: %s\n", m.Timestamp.Format(time.RFC3339))
		if m.Metadata.Description != "" {
			fmt.Printf("    Description: %s\n", m.Metadata.Description)
		}
		if m.Metadata.Formula != "" {
			fmt.Printf("    Formula: %s\n", m.Metadata.Formula)
		}
	}
}

func cmdCreateMetric(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: create-metric <key> <name> [unit] [description] [formula]")
		os.Exit(1)
	}

	key := args[0]
	name := args[1]
	unit := ""
	description := ""
	formula := ""

	if len(args) > 2 {
		unit = args[2]
	}
	if len(args) > 3 {
		description = args[3]
	}
	if len(args) > 4 {
		formula = args[4]
	}

	api := NewAPIClient(*serverURL)

	metadata := common.MetricMetadata{
		Name:        name,
		Unit:        unit,
		Description: description,
		Formula:     formula,
	}

	err := api.CreateMetric(key, metadata)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Metric created: %s\n", key)
}

func cmdUpdateMetric(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: update-metric <key> <value> [timestamp]")
		os.Exit(1)
	}

	key := args[0]
	value, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		fmt.Printf("Invalid value: %v\n", err)
		os.Exit(1)
	}

	var timestamp *time.Time
	if len(args) > 2 {
		t, err := time.Parse(time.RFC3339, args[2])
		if err != nil {
			fmt.Printf("Invalid timestamp: %v\n", err)
			os.Exit(1)
		}
		timestamp = &t
	}

	api := NewAPIClient(*serverURL)

	err = api.UpdateMetricValue(key, value, timestamp)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Metric updated")
}

func cmdGetMetric(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: get-metric <key>")
		os.Exit(1)
	}

	key := args[0]
	api := NewAPIClient(*serverURL)

	metric, err := api.GetMetric(key)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	expired := ""
	if metric.IsExpired {
		expired = " [EXPIRED]"
	}

	fmt.Printf("Metric: %s%s\n", metric.Key, expired)
	fmt.Printf("  Name: %s\n", metric.Metadata.Name)
	fmt.Printf("  Value: %.2f %s\n", metric.Value, metric.Metadata.Unit)
	fmt.Printf("  Updated: %s\n", metric.Timestamp.Format(time.RFC3339))
	if metric.Metadata.Description != "" {
		fmt.Printf("  Description: %s\n", metric.Metadata.Description)
	}
	if metric.Metadata.Formula != "" {
		fmt.Printf("  Formula: %s\n", metric.Metadata.Formula)
	}
	if metric.AlertThreshold != nil {
		fmt.Printf("  Alert Thresholds:\n")
		if metric.AlertThreshold.MaxValue != nil {
			fmt.Printf("    Max: %.2f\n", *metric.AlertThreshold.MaxValue)
		}
		if metric.AlertThreshold.MinValue != nil {
			fmt.Printf("    Min: %.2f\n", *metric.AlertThreshold.MinValue)
		}
	}
}

func cmdSetThreshold(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: set-threshold <key> [--max=<value>] [--min=<value>]")
		os.Exit(1)
	}

	key := args[0]
	threshold := &common.AlertThreshold{}

	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--max=") {
			valStr := strings.TrimPrefix(arg, "--max=")
			val, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				fmt.Printf("Invalid max value: %v\n", err)
				os.Exit(1)
			}
			threshold.MaxValue = &val
		} else if strings.HasPrefix(arg, "--min=") {
			valStr := strings.TrimPrefix(arg, "--min=")
			val, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				fmt.Printf("Invalid min value: %v\n", err)
				os.Exit(1)
			}
			threshold.MinValue = &val
		}
	}

	if threshold.MaxValue == nil && threshold.MinValue == nil {
		fmt.Println("At least one threshold (--max or --min) is required")
		os.Exit(1)
	}

	api := NewAPIClient(*serverURL)

	err := api.SetMetricThreshold(key, threshold)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Threshold set")
}

func cmdHistory(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: history <key> [--start=<time>] [--end=<time>]")
		os.Exit(1)
	}

	key := args[0]
	var startTime, endTime *time.Time

	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--start=") {
			valStr := strings.TrimPrefix(arg, "--start=")
			t, err := time.Parse(time.RFC3339, valStr)
			if err != nil {
				fmt.Printf("Invalid start time: %v\n", err)
				os.Exit(1)
			}
			startTime = &t
		} else if strings.HasPrefix(arg, "--end=") {
			valStr := strings.TrimPrefix(arg, "--end=")
			t, err := time.Parse(time.RFC3339, valStr)
			if err != nil {
				fmt.Printf("Invalid end time: %v\n", err)
				os.Exit(1)
			}
			endTime = &t
		}
	}

	api := NewAPIClient(*serverURL)

	history, err := api.GetMetricHistory(key, startTime, endTime)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(history) == 0 {
		fmt.Println("No history data found")
		return
	}

	fmt.Printf("History for %s (%d points):\n", key, len(history))
	for _, dp := range history {
		fmt.Printf("  %s: %.2f\n", dp.Timestamp.Format(time.RFC3339), dp.Value)
	}
}

func cmdTrend(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: trend <key>")
		os.Exit(1)
	}

	key := args[0]
	api := NewAPIClient(*serverURL)

	trend, err := api.GetMetricTrend(key)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if trend == nil || len(trend) == 0 {
		fmt.Println("No trend data available")
		return
	}

	fmt.Printf("Trend for %s (last hour, 5min intervals):\n", key)
	for _, tp := range trend {
		fmt.Printf("  %s: %.2f\n", tp.Period, tp.Value)
	}
}

func cmdComparison(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: comparison <key>")
		os.Exit(1)
	}

	key := args[0]
	api := NewAPIClient(*serverURL)

	comparison, err := api.GetMetricComparison(key)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Comparison for %s:\n", key)
	fmt.Printf("  Current Value: %.2f\n", comparison.CurrentValue)
	fmt.Printf("  Compare Value: %.2f\n", comparison.CompareValue)
	fmt.Printf("  YoY: %.2f%%\n", comparison.YoYPercent)
	fmt.Printf("  MoM: %.2f%%\n", comparison.MoMPercent)
}

func cmdSubscribe(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: subscribe <keys...> [--rate=<seconds>]")
		os.Exit(1)
	}

	keys := make([]string, 0)
	refreshRate := int(common.MinRefreshRate.Seconds())

	for _, arg := range args {
		if strings.HasPrefix(arg, "--rate=") {
			valStr := strings.TrimPrefix(arg, "--rate=")
			val, err := strconv.Atoi(valStr)
			if err != nil {
				fmt.Printf("Invalid rate: %v\n", err)
				os.Exit(1)
			}
			if val < int(common.MinRefreshRate.Seconds()) {
				fmt.Printf("Refresh rate must be at least %d seconds\n", common.MinRefreshRate.Seconds())
				os.Exit(1)
			}
			refreshRate = val
		} else {
			keys = append(keys, arg)
		}
	}

	if len(keys) == 0 {
		fmt.Println("At least one metric key is required")
		os.Exit(1)
	}

	wsURL := strings.Replace(*serverURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)

	wsClient := NewWSClient(wsURL)
	wsClient.SetClientID(*clientID)

	wsClient.SetMetricHandler(func(m *common.Metric) {
		expired := ""
		if m.IsExpired {
			expired = " [EXPIRED]"
		}
		fmt.Printf("[%s] %s: %.2f %s%s\n",
			m.Timestamp.Format("15:04:05"),
			m.Key, m.Value, m.Metadata.Unit, expired)
	})

	wsClient.SetAlertHandler(func(a *common.Alert) {
		fmt.Printf("\n[ALERT] %s: %s\n", a.MetricKey, a.Message)
		fmt.Printf("  Threshold: %.2f, Actual: %.2f\n", a.Threshold, a.ActualValue)
		fmt.Printf("  Time: %s\n", a.Timestamp.Format(time.RFC3339))
	})

	wsClient.SetErrorHandler(func(err string) {
		fmt.Printf("[ERROR] %s\n", err)
	})

	fmt.Printf("Connecting to %s...\n", wsURL)
	if err := wsClient.Connect(); err != nil {
		fmt.Printf("Connection error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connected")

	fmt.Printf("Subscribing to: %v (rate: %ds)\n", keys, refreshRate)
	if err := wsClient.Subscribe(keys, refreshRate); err != nil {
		fmt.Printf("Subscribe error: %v\n", err)
		wsClient.Disconnect()
		os.Exit(1)
	}
	fmt.Println("Subscribed")
	fmt.Println("Listening for updates... (Ctrl+C to exit)")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	fmt.Println("\nDisconnecting...")
	wsClient.Disconnect()
	fmt.Println("Done")
}

func cmdAlerts() {
	api := NewAPIClient(*serverURL)

	alerts, err := api.GetAlerts()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(alerts) == 0 {
		fmt.Println("No pending alerts")
		return
	}

	fmt.Printf("Pending Alerts (%d):\n", len(alerts))
	for _, a := range alerts {
		fmt.Printf("\n  ID: %s\n", a.ID)
		fmt.Printf("  Metric: %s\n", a.MetricKey)
		fmt.Printf("  Type: %s\n", a.AlertType)
		fmt.Printf("  Message: %s\n", a.Message)
		fmt.Printf("  Threshold: %.2f, Actual: %.2f\n", a.Threshold, a.ActualValue)
		fmt.Printf("  Time: %s\n", a.Timestamp.Format(time.RFC3339))
	}
}

func cmdDelayStats() {
	api := NewAPIClient(*serverURL)

	stats, err := api.GetDelayStats()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Push Delay Statistics:")
	fmt.Printf("  Current: %.2f ms\n", stats["current_ms"])
	fmt.Printf("  Average: %.2f ms\n", stats["average_ms"])
	fmt.Printf("  Max: %.2f ms\n", stats["max_ms"])
	fmt.Printf("  P99: %.2f ms\n", stats["p99_ms"])
	fmt.Printf("  Warning: %v, Critical: %v\n", stats["warning"], stats["critical"])
	fmt.Printf("  Thresholds: Warning=%vms, Critical=%vms\n",
		stats["warning_threshold_ms"], stats["critical_threshold_ms"])
}

func cmdSaveLayout(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: save-layout <metrics...>")
		os.Exit(1)
	}

	api := NewAPIClient(*serverURL)

	order := make([]int, len(args))
	for i := range order {
		order[i] = i
	}

	err := api.SaveDashboardLayout(*clientID, args, order)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Layout saved")
}

func cmdGetLayout(args []string) {
	api := NewAPIClient(*serverURL)

	layout, err := api.GetDashboardLayout(*clientID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Dashboard Layout for client %s:\n", *clientID)
	fmt.Printf("  Metrics: %v\n", layout.Metrics)
	fmt.Printf("  Order: %v\n", layout.Order)
}

func cmdCleanup() {
	api := NewAPIClient(*serverURL)

	err := api.CleanupOldData()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Old data cleanup triggered")
}

func cmdInteractive() {
	fmt.Println("Interactive Mode")
	fmt.Println("----------------")
	fmt.Println("Commands:")
	fmt.Println("  status          - Get server status")
	fmt.Println("  metrics         - List metrics")
	fmt.Println("  subscribe <keys> - Subscribe to metrics")
	fmt.Println("  unsubscribe <keys> - Unsubscribe from metrics")
	fmt.Println("  update <key> <value> - Update metric value")
	fmt.Println("  alerts          - Show alerts")
	fmt.Println("  help            - Show this help")
	fmt.Println("  quit            - Exit")

	api := NewAPIClient(*serverURL)

	wsURL := strings.Replace(*serverURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)

	wsClient := NewWSClient(wsURL)
	wsClient.SetClientID(*clientID)

	wsClient.SetMetricHandler(func(m *common.Metric) {
		expired := ""
		if m.IsExpired {
			expired = " [EXPIRED]"
		}
		fmt.Printf("\n[UPDATE] %s: %.2f %s%s\n", m.Key, m.Value, m.Metadata.Unit, expired)
		fmt.Print("> ")
	})

	wsClient.SetAlertHandler(func(a *common.Alert) {
		fmt.Printf("\n[ALERT] %s: %s\n", a.MetricKey, a.Message)
		fmt.Print("> ")
	})

	wsClient.SetErrorHandler(func(err string) {
		fmt.Printf("\n[ERROR] %s\n", err)
		fmt.Print("> ")
	})

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]
		args := parts[1:]

		switch cmd {
		case "status":
			status, err := api.GetServerStatus(true)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Online: %d/%d, Queue: %d, Metrics: %d\n",
					status.OnlineClients, status.MaxClients, status.QueueSize, status.MetricsCount)
			}

		case "metrics":
			metrics, err := api.GetAllMetrics()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				for _, m := range metrics {
					expired := ""
					if m.IsExpired {
						expired = " [EXPIRED]"
					}
					fmt.Printf("  %s: %.2f %s%s\n", m.Key, m.Value, m.Metadata.Unit, expired)
				}
			}

		case "subscribe":
			if len(args) < 1 {
				fmt.Println("Usage: subscribe <keys...>")
				continue
			}

			if !wsClient.IsConnected() {
				if err := wsClient.Connect(); err != nil {
					fmt.Printf("Connection error: %v\n", err)
					continue
				}
				fmt.Println("Connected")
			}

			if err := wsClient.Subscribe(args, int(common.MinRefreshRate.Seconds())); err != nil {
				fmt.Printf("Subscribe error: %v\n", err)
			} else {
				fmt.Printf("Subscribed to: %v\n", args)
			}

		case "unsubscribe":
			if len(args) < 1 {
				fmt.Println("Usage: unsubscribe <keys...>")
				continue
			}

			if err := wsClient.Unsubscribe(args); err != nil {
				fmt.Printf("Unsubscribe error: %v\n", err)
			} else {
				fmt.Printf("Unsubscribed from: %v\n", args)
			}

		case "update":
			if len(args) < 2 {
				fmt.Println("Usage: update <key> <value>")
				continue
			}

			key := args[0]
			value, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				fmt.Printf("Invalid value: %v\n", err)
				continue
			}

			if err := api.UpdateMetricValue(key, value, nil); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Updated")
			}

		case "alerts":
			alerts, err := api.GetAlerts()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else if len(alerts) == 0 {
				fmt.Println("No alerts")
			} else {
				for _, a := range alerts {
					fmt.Printf("  %s: %s\n", a.MetricKey, a.Message)
				}
			}

		case "help":
			fmt.Println("Interactive Mode Commands:")
			fmt.Println("  status          - Get server status")
			fmt.Println("  metrics         - List metrics")
			fmt.Println("  subscribe <keys> - Subscribe to metrics")
			fmt.Println("  unsubscribe <keys> - Unsubscribe from metrics")
			fmt.Println("  update <key> <value> - Update metric value")
			fmt.Println("  alerts          - Show alerts")
			fmt.Println("  help            - Show this help")
			fmt.Println("  quit            - Exit")

		case "quit", "exit":
			if wsClient.IsConnected() {
				wsClient.Disconnect()
			}
			fmt.Println("Bye!")
			return

		default:
			fmt.Printf("Unknown command: %s\n", cmd)
			fmt.Println("Type 'help' for available commands")
		}
	}
}
