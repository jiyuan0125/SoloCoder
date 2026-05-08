package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"

	"github.com/health-aggregator/pkg/common"
)

func CmdStatus(client *APIClient, serviceName string) error {
	if serviceName != "" {
		status, err := client.GetServiceStatus(serviceName)
		if err != nil {
			return err
		}
		printServiceStatus(status)
		return nil
	}

	agg, err := client.GetAggregateStatus()
	if err != nil {
		return err
	}
	printAggregateStatus(agg)
	return nil
}

func CmdHistory(client *APIClient, serviceName string, limit int) error {
	history, err := client.GetServiceHistory(serviceName)
	if err != nil {
		return err
	}

	fmt.Printf("History for %s (total: %d)\n", serviceName, history.Total)
	fmt.Println(strings.Repeat("-", 80))

	results := history.Results
	if limit > 0 && limit < len(results) {
		results = results[len(results)-limit:]
	}

	for i, r := range results {
		status := color.GreenString("SUCCESS")
		if !r.Success {
			status = color.RedString("FAILED")
		}
		latency := r.Latency.Round(time.Millisecond)
		fmt.Printf("%d. %s | %s | %s", i+1, r.Timestamp.Format("2006-01-02 15:04:05"), status, latency)
		if r.Error != "" {
			fmt.Printf(" | %s", r.Error)
		}
		fmt.Println()
	}
	return nil
}

func CmdAdd(client *APIClient, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: add <name> <http|tcp> <target> [interval] [timeout] [failure_threshold] [recovery_threshold] [group]")
	}

	name := args[0]
	probeTypeStr := strings.ToLower(args[1])
	target := args[2]

	var probeType common.ProbeType
	switch probeTypeStr {
	case "http":
		probeType = common.ProbeTypeHTTP
	case "tcp":
		probeType = common.ProbeTypeTCP
	default:
		return fmt.Errorf("invalid probe type: %s (must be http or tcp)", probeTypeStr)
	}

	cfg := &common.ServiceConfig{
		Name:      name,
		ProbeType: probeType,
		Target:    target,
	}

	if len(args) > 3 {
		interval, err := time.ParseDuration(args[3])
		if err != nil {
			return fmt.Errorf("invalid interval: %v", err)
		}
		cfg.Interval = interval
	}
	if len(args) > 4 {
		timeout, err := time.ParseDuration(args[4])
		if err != nil {
			return fmt.Errorf("invalid timeout: %v", err)
		}
		cfg.Timeout = timeout
	}
	if len(args) > 5 {
		ft, err := strconv.Atoi(args[5])
		if err != nil {
			return fmt.Errorf("invalid failure_threshold: %v", err)
		}
		cfg.FailureThreshold = ft
	}
	if len(args) > 6 {
		rt, err := strconv.Atoi(args[6])
		if err != nil {
			return fmt.Errorf("invalid recovery_threshold: %v", err)
		}
		cfg.RecoveryThreshold = rt
	}
	if len(args) > 7 {
		cfg.Group = args[7]
	}

	resp, err := client.AddService(cfg)
	if err != nil {
		return err
	}
	fmt.Printf("Success: %s\n", resp.Message)
	return nil
}

func CmdRemove(client *APIClient, serviceName string) error {
	resp, err := client.RemoveService(serviceName)
	if err != nil {
		return err
	}
	fmt.Printf("Success: %s\n", resp.Message)
	return nil
}

func CmdSync(client *APIClient, cfg *ClientConfig) error {
	if cfg == nil || len(cfg.Services) == 0 {
		return fmt.Errorf("no services defined in config file")
	}

	for _, svc := range cfg.Services {
		existing, _ := client.GetServiceStatus(svc.Name)
		if existing != nil {
			fmt.Printf("Service %s already exists, skipping\n", svc.Name)
			continue
		}

		resp, err := client.AddService(svc)
		if err != nil {
			fmt.Printf("Failed to add %s: %v\n", svc.Name, err)
			continue
		}
		fmt.Printf("Added %s: %s\n", svc.Name, resp.Message)
	}
	return nil
}

func CmdDashboard(client *APIClient, interval time.Duration) error {
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	render := func() {
		agg, err := client.GetAggregateStatus()
		if err != nil {
			fmt.Printf("\033[2J\033[HError: %v\n", err)
			return
		}
		fmt.Print("\033[2J\033[H")
		printDashboardHeader(agg)
		printDashboardServices(agg)
		fmt.Printf("\nLast updated: %s | Press Ctrl+C to exit\n", time.Now().Format("2006-01-02 15:04:05"))
	}

	render()
	for {
		select {
		case <-stopCh:
			fmt.Println("\nExiting dashboard...")
			return nil
		case <-ticker.C:
			render()
		}
	}
}

func printDashboardHeader(agg *common.AggregateStatus) {
	fmt.Println("=== Health Aggregator Dashboard ===")
	fmt.Println()

	var statusColor *color.Color
	switch agg.OverallStatus {
	case common.StatusHealthy:
		statusColor = color.New(color.FgGreen, color.Bold)
	case common.StatusDegraded:
		statusColor = color.New(color.FgYellow, color.Bold)
	case common.StatusUnhealthy:
		statusColor = color.New(color.FgRed, color.Bold)
	}

	fmt.Printf("Overall Status: ")
	statusColor.Printf("%s", strings.ToUpper(string(agg.OverallStatus)))
	fmt.Printf("  |  Healthy: %d/%d  |  Unhealthy: %d\n", agg.HealthyCount, agg.TotalServices, agg.UnhealthyCount)
	fmt.Println()

	if agg.GroupAggregates != nil && len(agg.GroupAggregates) > 0 {
		fmt.Println("Groups:")
		for name, ga := range agg.GroupAggregates {
			var gc *color.Color
			switch ga.OverallStatus {
			case common.StatusHealthy:
				gc = color.New(color.FgGreen)
			case common.StatusDegraded:
				gc = color.New(color.FgYellow)
			case common.StatusUnhealthy:
				gc = color.New(color.FgRed)
			}
			fmt.Printf("  [%s] ", name)
			gc.Printf("%s", ga.OverallStatus)
			fmt.Printf(" (%d/%d healthy)\n", ga.HealthyCount, ga.TotalServices)
		}
		fmt.Println()
	}
}

func printDashboardServices(agg *common.AggregateStatus) {
	fmt.Printf("%-20s %-10s %-25s %-10s %-15s %s\n", "NAME", "TYPE", "TARGET", "STATUS", "CONSECUTIVE", "LATENCY")
	fmt.Println(strings.Repeat("-", 90))

	for _, s := range agg.ServiceStatuses {
		statusStr := strings.ToUpper(string(s.CurrentStatus))
		var statusColor *color.Color
		switch s.CurrentStatus {
		case common.StatusHealthy:
			statusColor = color.New(color.FgGreen)
		case common.StatusDegraded:
			statusColor = color.New(color.FgYellow)
		case common.StatusUnhealthy:
			statusColor = color.New(color.FgRed)
		}

		consecutive := fmt.Sprintf("S:%d/F:%d", s.ConsecutiveSuccess, s.ConsecutiveFail)
		latency := "-"
		if s.LastProbe != nil {
			latency = s.LastProbe.Latency.Round(time.Millisecond).String()
		}

		fmt.Printf("%-20s %-10s %-25s ", s.Name, s.ProbeType, truncate(s.Target, 25))
		statusColor.Printf("%-10s", statusStr)
		fmt.Printf(" %-15s %s\n", consecutive, latency)
	}
}

func printAggregateStatus(agg *common.AggregateStatus) {
	fmt.Println("=== Aggregate Status ===")
	fmt.Printf("Overall: %s\n", agg.OverallStatus)
	fmt.Printf("Total Services: %d\n", agg.TotalServices)
	fmt.Printf("Healthy: %d\n", agg.HealthyCount)
	fmt.Printf("Unhealthy: %d\n", agg.UnhealthyCount)
	fmt.Println()

	if agg.GroupAggregates != nil && len(agg.GroupAggregates) > 0 {
		fmt.Println("Groups:")
		for name, ga := range agg.GroupAggregates {
			fmt.Printf("  [%s] %s (%d/%d healthy)\n", name, ga.OverallStatus, ga.HealthyCount, ga.TotalServices)
		}
		fmt.Println()
	}

	fmt.Println("Services:")
	for _, s := range agg.ServiceStatuses {
		printServiceStatus(s)
	}
}

func printServiceStatus(s *common.ServiceStatus) {
	var statusColor *color.Color
	switch s.CurrentStatus {
	case common.StatusHealthy:
		statusColor = color.New(color.FgGreen)
	case common.StatusDegraded:
		statusColor = color.New(color.FgYellow)
	case common.StatusUnhealthy:
		statusColor = color.New(color.FgRed)
	}

	fmt.Printf("  %s (", s.Name)
	statusColor.Printf("%s", s.CurrentStatus)
	fmt.Printf(")\n")
	fmt.Printf("    Type: %s | Target: %s\n", s.ProbeType, s.Target)
	if s.Group != "" {
		fmt.Printf("    Group: %s\n", s.Group)
	}
	fmt.Printf("    Consecutive Success: %d | Consecutive Fail: %d\n", s.ConsecutiveSuccess, s.ConsecutiveFail)
	if s.LastProbe != nil {
		fmt.Printf("    Last Probe: %s | Latency: %s", s.LastProbe.Timestamp.Format("2006-01-02 15:04:05"), s.LastProbe.Latency.Round(time.Millisecond))
		if s.LastProbe.Error != "" {
			fmt.Printf(" | Error: %s", s.LastProbe.Error)
		}
		fmt.Println()
	}
	fmt.Println()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
