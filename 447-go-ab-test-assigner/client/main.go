package main

import (
	"abtest/api"
	"abtest/client/cmd"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

const (
	defaultServerURL = "http://localhost:8080"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := os.Getenv("ABTEST_SERVER")
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	client := cmd.NewClient(serverURL)

	command := os.Args[1]
	args := os.Args[2:]

	var err error
	switch command {
	case "create":
		err = handleCreate(client, args)
	case "start":
		err = handleStart(client, args)
	case "end":
		err = handleEnd(client, args)
	case "archive":
		err = handleArchive(client, args)
	case "list":
		err = handleList(client, args)
	case "get":
		err = handleGet(client, args)
	case "assign":
		err = handleAssign(client, args)
	case "metrics":
		err = handleMetrics(client, args)
	case "stats":
		err = handleStats(client, args)
	case "user":
		err = handleUserAssignments(client, args)
	case "traffic":
		err = handleTraffic(client, args)
	case "health":
		err = handleHealth(client)
	case "help":
		printUsage()
	default:
		printUsage()
		fmt.Fprintf(os.Stderr, "\nUnknown command: %s\n", command)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleCreate(client *cmd.Client, args []string) error {
	flagSet := flag.NewFlagSet("create", flag.ExitOnError)
	name := flagSet.String("name", "", "Experiment name (required)")
	expType := flagSet.String("type", "normal", "Experiment type: normal or gray")
	control := flagSet.Int("control", 50, "Control group percentage (0-100)")
	treatment := flagSet.Int("treatment", 50, "Treatment group percentage (0-100)")

	flagSet.Parse(args)

	if *name == "" {
		return fmt.Errorf("experiment name is required, use -name flag")
	}

	etype := api.ExperimentType(*expType)
	if etype != api.ExperimentTypeNormal && etype != api.ExperimentTypeGray {
		return fmt.Errorf("invalid experiment type: %s, must be 'normal' or 'gray'", *expType)
	}

	exp, err := client.CreateExperiment(*name, etype, *control, *treatment)
	if err != nil {
		return err
	}

	printExperiment(exp)
	return nil
}

func handleStart(client *cmd.Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: abtest start <experiment-id>")
	}

	exp, err := client.StartExperiment(args[0])
	if err != nil {
		return err
	}

	printExperiment(exp)
	return nil
}

func handleEnd(client *cmd.Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: abtest end <experiment-id>")
	}

	exp, err := client.EndExperiment(args[0])
	if err != nil {
		return err
	}

	printExperiment(exp)
	return nil
}

func handleArchive(client *cmd.Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: abtest archive <experiment-id>")
	}

	exp, err := client.ArchiveExperiment(args[0])
	if err != nil {
		return err
	}

	printExperiment(exp)
	return nil
}

func handleList(client *cmd.Client, args []string) error {
	var status *api.ExperimentStatus
	if len(args) > 0 {
		s := api.ExperimentStatus(args[0])
		status = &s
	}

	experiments, err := client.ListExperiments(status)
	if err != nil {
		return err
	}

	printExperimentList(experiments)
	return nil
}

func handleGet(client *cmd.Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: abtest get <experiment-id>")
	}

	exp, err := client.GetExperiment(args[0])
	if err != nil {
		return err
	}

	printExperiment(exp)
	return nil
}

func handleAssign(client *cmd.Client, args []string) error {
	flagSet := flag.NewFlagSet("assign", flag.ExitOnError)
	experimentID := flagSet.String("exp", "", "Experiment ID (required)")
	userID := flagSet.String("user", "", "User ID (required)")

	flagSet.Parse(args)

	if *experimentID == "" || *userID == "" {
		return fmt.Errorf("both -exp and -user flags are required")
	}

	group, err := client.AssignUser(*experimentID, *userID)
	if err != nil {
		return err
	}

	fmt.Printf("User %s assigned to %s group in experiment %s\n", *userID, group, *experimentID)
	return nil
}

func handleMetrics(client *cmd.Client, args []string) error {
	flagSet := flag.NewFlagSet("metrics", flag.ExitOnError)
	experimentID := flagSet.String("exp", "", "Experiment ID (required)")
	userID := flagSet.String("user", "", "User ID (required)")
	converted := flagSet.Bool("converted", false, "Whether user converted")
	duration := flagSet.Float64("duration", 0, "Stay duration in seconds")

	flagSet.Parse(args)

	if *experimentID == "" || *userID == "" {
		return fmt.Errorf("both -exp and -user flags are required")
	}

	err := client.RecordMetrics(*experimentID, *userID, *converted, *duration)
	if err != nil {
		return err
	}

	fmt.Printf("Metrics recorded for user %s in experiment %s\n", *userID, *experimentID)
	return nil
}

func handleStats(client *cmd.Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: abtest stats <experiment-id>")
	}

	stats, err := client.GetStats(args[0])
	if err != nil {
		return err
	}

	printStats(stats)
	return nil
}

func handleUserAssignments(client *cmd.Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: abtest user <user-id>")
	}

	assignments, err := client.GetUserAssignments(args[0])
	if err != nil {
		return err
	}

	printUserAssignments(args[0], assignments)
	return nil
}

func handleTraffic(client *cmd.Client, args []string) error {
	flagSet := flag.NewFlagSet("traffic", flag.ExitOnError)
	experimentID := flagSet.String("exp", "", "Experiment ID (required)")
	control := flagSet.Int("control", 50, "Control group percentage (0-100)")
	treatment := flagSet.Int("treatment", 50, "Treatment group percentage (0-100)")

	flagSet.Parse(args)

	if *experimentID == "" {
		return fmt.Errorf("-exp flag is required")
	}

	exp, err := client.UpdateTraffic(*experimentID, *control, *treatment)
	if err != nil {
		return err
	}

	printExperiment(exp)
	return nil
}

func handleHealth(client *cmd.Client) error {
	ok, err := client.HealthCheck()
	if err != nil {
		return err
	}

	if ok {
		fmt.Println("Server status: OK")
	} else {
		fmt.Println("Server status: UNHEALTHY")
	}
	return nil
}

func printExperiment(exp *api.Experiment) {
	fmt.Println("Experiment:")
	fmt.Printf("  ID:          %s\n", exp.ID)
	fmt.Printf("  Name:        %s\n", exp.Name)
	fmt.Printf("  Type:        %s\n", exp.Type)
	fmt.Printf("  Status:      %s\n", exp.Status)
	fmt.Printf("  Control:     %d%%\n", exp.ControlPercentage)
	fmt.Printf("  Treatment:   %d%%\n", exp.TreatmentPercentage)
	fmt.Printf("  Created:     %s\n", exp.CreatedAt.Format("2006-01-02 15:04:05"))
	if exp.StartedAt != nil {
		fmt.Printf("  Started:     %s\n", exp.StartedAt.Format("2006-01-02 15:04:05"))
	}
	if exp.EndedAt != nil {
		fmt.Printf("  Ended:       %s\n", exp.EndedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("  Min Duration:%d days\n", exp.MinDurationDays)
}

func printExperimentList(experiments []api.Experiment) {
	if len(experiments) == 0 {
		fmt.Println("No experiments found")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTYPE\tSTATUS\tCONTROL\tTREATMENT\tCREATED")
	for _, exp := range experiments {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d%%\t%d%%\t%s\n",
			shortID(exp.ID),
			exp.Name,
			exp.Type,
			exp.Status,
			exp.ControlPercentage,
			exp.TreatmentPercentage,
			exp.CreatedAt.Format("2006-01-02"),
		)
	}
	w.Flush()
}

func printStats(stats *api.ExperimentStatsResponse) {
	fmt.Println("Experiment Statistics:")
	fmt.Printf("  Name:        %s\n", stats.Experiment.Name)
	fmt.Printf("  Status:      %s\n", stats.Experiment.Status)

	if !stats.SampleSufficient {
		fmt.Println("\n  ⚠️  Sample size insufficient - experiment needs to run at least 7 days")
		fmt.Println("      Results are for reference only")
	}

	fmt.Println("\nControl Group:")
	fmt.Printf("  Total Users:     %d\n", stats.ControlStats.TotalUsers)
	fmt.Printf("  Converted:       %d\n", stats.ControlStats.ConvertedUsers)
	fmt.Printf("  Conversion Rate: %.2f%%\n", stats.ControlStats.ConversionRate*100)
	fmt.Printf("  Avg Duration:    %.2fs\n", stats.ControlStats.AvgStayDuration)

	fmt.Println("\nTreatment Group:")
	fmt.Printf("  Total Users:     %d\n", stats.TreatmentStats.TotalUsers)
	fmt.Printf("  Converted:       %d\n", stats.TreatmentStats.ConvertedUsers)
	fmt.Printf("  Conversion Rate: %.2f%%\n", stats.TreatmentStats.ConversionRate*100)
	fmt.Printf("  Avg Duration:    %.2fs\n", stats.TreatmentStats.AvgStayDuration)

	if len(stats.Significance) > 0 {
		fmt.Println("\nStatistical Significance (95% confidence):")
		for _, sig := range stats.Significance {
			symbol := "✗"
			if sig.IsSignificant {
				symbol = "✓"
			}
			fmt.Printf("  %s: %s (p-value: %.4f)\n", sig.Metric, symbol, sig.PValue)
		}
	}
}

func printUserAssignments(userID string, assignments []api.UserAssignment) {
	if len(assignments) == 0 {
		fmt.Printf("No assignments found for user %s\n", userID)
		return
	}

	fmt.Printf("Assignments for user %s:\n", userID)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "EXPERIMENT\tNAME\tGROUP")
	for _, a := range assignments {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			shortID(a.ExperimentID),
			a.ExperimentName,
			a.Group,
		)
	}
	w.Flush()
}

func shortID(id string) string {
	parts := strings.Split(id, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return id
}

func printUsage() {
	fmt.Println("AB Test CLI Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  abtest <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create      Create a new experiment")
	fmt.Println("  start       Start an experiment")
	fmt.Println("  end         End an experiment")
	fmt.Println("  archive     Archive an experiment")
	fmt.Println("  list        List experiments")
	fmt.Println("  get         Get experiment details")
	fmt.Println("  assign      Assign user to a group")
	fmt.Println("  metrics     Record user metrics")
	fmt.Println("  stats       Get experiment statistics")
	fmt.Println("  user        Get user's assignments")
	fmt.Println("  traffic     Update traffic distribution (gray release only)")
	fmt.Println("  health      Check server health")
	fmt.Println("  help        Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Create a normal AB test (50/50 split)")
	fmt.Println("  abtest create -name \"Button Test\" -type normal -control 50 -treatment 50")
	fmt.Println()
	fmt.Println("  # Create a gray release (defaults to 95/5 split)")
	fmt.Println("  abtest create -name \"New Feature Rollout\" -type gray")
	fmt.Println()
	fmt.Println("  # Start an experiment")
	fmt.Println("  abtest start <experiment-id>")
	fmt.Println()
	fmt.Println("  # Assign user to a group")
	fmt.Println("  abtest assign -exp <experiment-id> -user user-123")
	fmt.Println()
	fmt.Println("  # Record metrics")
	fmt.Println("  abtest metrics -exp <experiment-id> -user user-123 -converted -duration 45.5")
	fmt.Println()
	fmt.Println("  # Get statistics")
	fmt.Println("  abtest stats <experiment-id>")
	fmt.Println()
	fmt.Println("  # Update traffic (gray release only, after 24h observation)")
	fmt.Println("  abtest traffic -exp <experiment-id> -control 90 -treatment 10")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  ABTEST_SERVER  Server URL (default: http://localhost:8080)")
}
