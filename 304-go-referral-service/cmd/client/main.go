package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	clientapi "referral-service/internal/client/api"
)

const defaultServerURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	client := clientapi.NewClient(serverURL)

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "generate-code":
		handleGenerateCode(client, args)
	case "bind":
		handleBind(client, args)
	case "stats":
		handleUserStats(client, args)
	case "set-reward":
		handleSetReward(client, args)
	case "get-reward":
		handleGetReward(client, args)
	case "admin-stats":
		handleAdminStats(client, args)
	case "complete-order":
		handleCompleteOrder(client, args)
	case "refund-order":
		handleRefundOrder(client, args)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleGenerateCode(client *clientapi.Client, args []string) {
	fs := flag.NewFlagSet("generate-code", flag.ExitOnError)
	userID := fs.String("user", "", "User ID")
	fs.Parse(args)

	if *userID == "" {
		fmt.Println("Error: -user flag is required")
		os.Exit(1)
	}

	code, err := client.GenerateReferralCode(*userID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Referral code for user %s: %s\n", *userID, code)
}

func handleBind(client *clientapi.Client, args []string) {
	fs := flag.NewFlagSet("bind", flag.ExitOnError)
	newUserID := fs.String("new-user", "", "New user ID")
	code := fs.String("code", "", "Referral code")
	fs.Parse(args)

	if *newUserID == "" || *code == "" {
		fmt.Println("Error: -new-user and -code flags are required")
		os.Exit(1)
	}

	resp, err := client.BindReferral(*newUserID, *code)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleUserStats(client *clientapi.Client, args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	userID := fs.String("user", "", "User ID")
	fs.Parse(args)

	if *userID == "" {
		fmt.Println("Error: -user flag is required")
		os.Exit(1)
	}

	stats, err := client.GetUserStats(*userID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== User Statistics ===")
	fmt.Printf("User ID: %s\n", stats.UserID)
	if stats.ReferralCode != "" {
		fmt.Printf("Referral Code: %s\n", stats.ReferralCode)
	}
	fmt.Printf("Total Referrals: %d\n", stats.TotalReferrals)
	fmt.Printf("Completed First Orders: %d\n", stats.CompletedFirstOrders)
	fmt.Printf("Total Points: %d\n", stats.TotalPoints)
}

func handleSetReward(client *clientapi.Client, args []string) {
	fs := flag.NewFlagSet("set-reward", flag.ExitOnError)
	fs.Parse(args)

	if len(fs.Args()) < 1 {
		fmt.Println("Error: points value is required")
		os.Exit(1)
	}

	points, err := strconv.ParseInt(fs.Args()[0], 10, 64)
	if err != nil {
		fmt.Printf("Error: invalid points value: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.SetRewardPoints(points)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleGetReward(client *clientapi.Client, args []string) {
	points, err := client.GetRewardPoints()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Current reward points per first order: %d\n", points)
}

func handleAdminStats(client *clientapi.Client, args []string) {
	stats, err := client.GetStats()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== System Statistics ===")
	fmt.Printf("Total Referrals: %d\n", stats.TotalReferrals)
	fmt.Printf("Completed First Orders: %d\n", stats.CompletedFirstOrders)
	fmt.Printf("First Order Rate: %.2f%%\n", stats.FirstOrderRate)
	fmt.Printf("Total Points Issued: %d\n", stats.TotalPointsIssued)
}

func handleCompleteOrder(client *clientapi.Client, args []string) {
	fs := flag.NewFlagSet("complete-order", flag.ExitOnError)
	userID := fs.String("user", "", "User ID")
	orderID := fs.String("order", "", "Order ID")
	fs.Parse(args)

	if *userID == "" || *orderID == "" {
		fmt.Println("Error: -user and -order flags are required")
		os.Exit(1)
	}

	resp, err := client.CompleteFirstOrder(*userID, *orderID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
		if resp.PointsAwarded > 0 {
			fmt.Printf("Points awarded: %d\n", resp.PointsAwarded)
		}
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleRefundOrder(client *clientapi.Client, args []string) {
	fs := flag.NewFlagSet("refund-order", flag.ExitOnError)
	userID := fs.String("user", "", "User ID")
	orderID := fs.String("order", "", "Order ID")
	fs.Parse(args)

	if *userID == "" || *orderID == "" {
		fmt.Println("Error: -user and -order flags are required")
		os.Exit(1)
	}

	resp, err := client.RefundFirstOrder(*userID, *orderID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
		if resp.PointsDeducted > 0 {
			fmt.Printf("Points deducted: %d\n", resp.PointsDeducted)
		}
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Referral Service Client")
	fmt.Println("")
	fmt.Println("Usage: client <command> [options]")
	fmt.Println("")
	fmt.Println("User Commands:")
	fmt.Println("  generate-code -user <user_id>    Generate or get referral code for a user")
	fmt.Println("  bind -new-user <user_id> -code <code>  Bind new user to a referrer")
	fmt.Println("  stats -user <user_id>             Get user's referral statistics")
	fmt.Println("")
	fmt.Println("Admin Commands:")
	fmt.Println("  set-reward <points>               Set reward points per first order")
	fmt.Println("  get-reward                         Get current reward points")
	fmt.Println("  admin-stats                        Get system-wide statistics")
	fmt.Println("")
	fmt.Println("System Commands (for testing):")
	fmt.Println("  complete-order -user <user_id> -order <order_id>  Mark first order as complete")
	fmt.Println("  refund-order -user <user_id> -order <order_id>    Refund a completed first order")
	fmt.Println("")
	fmt.Println("Environment Variables:")
	fmt.Println("  SERVER_URL                         Server URL (default: http://localhost:8080)")
}
