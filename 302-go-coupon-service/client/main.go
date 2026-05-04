package main

import (
	"coupon-service/pkg/api"
	"flag"
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := flag.String("server", "http://localhost:8080", "Coupon service server URL")
	flag.Parse()

	client := NewAPIClient(*serverURL)

	command := os.Args[1]
	switch command {
	case "create-batch":
		handleCreateBatch(client)
	case "update-batch":
		handleUpdateBatch(client)
	case "batch-stats":
		handleBatchStats(client)
	case "claim":
		handleClaim(client)
	case "redeem":
		handleRedeem(client)
	case "my-coupons":
		handleMyCoupons(client)
	case "return":
		handleReturn(client)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Coupon Service Client")
	fmt.Println("Usage:")
	fmt.Println("  Admin Commands:")
	fmt.Println("    create-batch -name=<name> -discount=<amount> -threshold=<amount> -total=<quantity> -start=<time> -end=<time> -per-user=<limit>")
	fmt.Println("    update-batch -id=<batch_id> [-name=<name>] [-discount=<amount>] [-threshold=<amount>] [-start=<time>] [-end=<time>] [-per-user=<limit>]")
	fmt.Println("    batch-stats -id=<batch_id>")
	fmt.Println("")
	fmt.Println("  User Commands:")
	fmt.Println("    claim -user=<user_id> -batch=<batch_id>")
	fmt.Println("    redeem -user=<user_id> -claim=<claim_id> -order=<amount>")
	fmt.Println("    my-coupons -user=<user_id>")
	fmt.Println("    return -user=<user_id> -claim=<claim_id>")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -server=<url>  Server URL (default: http://localhost:8080)")
}

func handleCreateBatch(client *APIClient) {
	flags := flag.NewFlagSet("create-batch", flag.ExitOnError)
	name := flags.String("name", "", "Coupon batch name")
	discount := flags.Float64("discount", 0, "Discount amount")
	threshold := flags.Float64("threshold", 0, "Minimum order amount threshold")
	total := flags.Int("total", 0, "Total quantity of coupons")
	start := flags.String("start", "", "Start time (format: 2006-01-02 or 2006-01-02 15:04:05)")
	end := flags.String("end", "", "End time (format: 2006-01-02 or 2006-01-02 15:04:05)")
	perUser := flags.Int("per-user", 1, "Maximum coupons per user")

	flags.Parse(os.Args[2:])

	if *name == "" || *discount <= 0 || *total <= 0 || *start == "" || *end == "" {
		fmt.Println("Error: name, discount, total, start, and end are required")
		flags.Usage()
		os.Exit(1)
	}

	req := &api.CreateBatchRequest{
		Name:          *name,
		Discount:      *discount,
		Threshold:     *threshold,
		TotalQuantity: *total,
		StartTime:     *start,
		EndTime:       *end,
		PerUserLimit:  *perUser,
	}

	resp, err := client.CreateBatch(req)
	if err != nil {
		fmt.Printf("Error creating batch: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
		fmt.Printf("Batch ID: %s\n", resp.Batch.ID)
		fmt.Printf("Name: %s\n", resp.Batch.Name)
		fmt.Printf("Discount: %.2f\n", resp.Batch.Discount)
		fmt.Printf("Threshold: %.2f\n", resp.Batch.Threshold)
		fmt.Printf("Total: %d\n", resp.Batch.TotalQuantity)
		fmt.Printf("Start: %s\n", resp.Batch.StartTime)
		fmt.Printf("End: %s\n", resp.Batch.EndTime)
		fmt.Printf("Per User Limit: %d\n", resp.Batch.PerUserLimit)
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleUpdateBatch(client *APIClient) {
	flags := flag.NewFlagSet("update-batch", flag.ExitOnError)
	id := flags.String("id", "", "Batch ID")
	name := flags.String("name", "", "Coupon batch name")
	discount := flags.Float64("discount", 0, "Discount amount")
	threshold := flags.Float64("threshold", -1, "Minimum order amount threshold")
	start := flags.String("start", "", "Start time (format: 2006-01-02 or 2006-01-02 15:04:05)")
	end := flags.String("end", "", "End time (format: 2006-01-02 or 2006-01-02 15:04:05)")
	perUser := flags.Int("per-user", 0, "Maximum coupons per user")

	flags.Parse(os.Args[2:])

	if *id == "" {
		fmt.Println("Error: batch ID is required")
		flags.Usage()
		os.Exit(1)
	}

	req := &api.UpdateBatchRequest{
		BatchID:      *id,
		Name:         *name,
		Discount:     *discount,
		Threshold:    *threshold,
		StartTime:    *start,
		EndTime:      *end,
		PerUserLimit: *perUser,
	}

	resp, err := client.UpdateBatch(req)
	if err != nil {
		fmt.Printf("Error updating batch: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
		fmt.Printf("Batch ID: %s\n", resp.Batch.ID)
		fmt.Printf("Name: %s\n", resp.Batch.Name)
		fmt.Printf("Discount: %.2f\n", resp.Batch.Discount)
		fmt.Printf("Threshold: %.2f\n", resp.Batch.Threshold)
		fmt.Printf("Total: %d\n", resp.Batch.TotalQuantity)
		fmt.Printf("Start: %s\n", resp.Batch.StartTime)
		fmt.Printf("End: %s\n", resp.Batch.EndTime)
		fmt.Printf("Per User Limit: %d\n", resp.Batch.PerUserLimit)
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleBatchStats(client *APIClient) {
	flags := flag.NewFlagSet("batch-stats", flag.ExitOnError)
	id := flags.String("id", "", "Batch ID")

	flags.Parse(os.Args[2:])

	if *id == "" {
		fmt.Println("Error: batch ID is required")
		flags.Usage()
		os.Exit(1)
	}

	resp, err := client.GetBatchStats(*id)
	if err != nil {
		fmt.Printf("Error getting batch stats: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
		fmt.Printf("Batch ID: %s\n", resp.Batch.ID)
		fmt.Printf("Name: %s\n", resp.Batch.Name)
		fmt.Printf("Discount: %.2f\n", resp.Batch.Discount)
		fmt.Printf("Threshold: %.2f\n", resp.Batch.Threshold)
		fmt.Printf("Total: %d\n", resp.Batch.TotalQuantity)
		fmt.Printf("Claimed: %d\n", resp.Claimed)
		fmt.Printf("Redeemed: %d\n", resp.Redeemed)
		fmt.Printf("Remaining: %d\n", resp.Remaining)
		fmt.Printf("Is All Claimed: %t\n", resp.Batch.IsAllClaimed)
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleClaim(client *APIClient) {
	flags := flag.NewFlagSet("claim", flag.ExitOnError)
	userID := flags.String("user", "", "User ID")
	batchID := flags.String("batch", "", "Batch ID")

	flags.Parse(os.Args[2:])

	if *userID == "" || *batchID == "" {
		fmt.Println("Error: user and batch are required")
		flags.Usage()
		os.Exit(1)
	}

	req := &api.ClaimCouponRequest{
		UserID:  *userID,
		BatchID: *batchID,
	}

	resp, err := client.ClaimCoupon(req)
	if err != nil {
		fmt.Printf("Error claiming coupon: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
		fmt.Printf("Claim ID: %s\n", resp.Claim.ID)
		fmt.Printf("User ID: %s\n", resp.Claim.UserID)
		fmt.Printf("Batch ID: %s\n", resp.Claim.BatchID)
		fmt.Printf("Claim Time: %s\n", resp.Claim.ClaimTime)
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleRedeem(client *APIClient) {
	flags := flag.NewFlagSet("redeem", flag.ExitOnError)
	userID := flags.String("user", "", "User ID")
	claimID := flags.String("claim", "", "Claim ID")
	orderAmount := flags.String("order", "", "Order amount")

	flags.Parse(os.Args[2:])

	if *userID == "" || *claimID == "" || *orderAmount == "" {
		fmt.Println("Error: user, claim, and order are required")
		flags.Usage()
		os.Exit(1)
	}

	amount, err := strconv.ParseFloat(*orderAmount, 64)
	if err != nil {
		fmt.Printf("Error: invalid order amount: %v\n", err)
		os.Exit(1)
	}

	req := &api.RedeemCouponRequest{
		UserID:      *userID,
		ClaimID:     *claimID,
		OrderAmount: amount,
	}

	resp, err := client.RedeemCoupon(req)
	if err != nil {
		fmt.Printf("Error redeeming coupon: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
		fmt.Printf("Redemption ID: %s\n", resp.Redemption.ID)
		fmt.Printf("Claim ID: %s\n", resp.Redemption.ClaimID)
		fmt.Printf("User ID: %s\n", resp.Redemption.UserID)
		fmt.Printf("Order Amount: %.2f\n", resp.Redemption.OrderAmount)
		fmt.Printf("Discount Used: %.2f\n", resp.Redemption.DiscountUsed)
		fmt.Printf("Redeem Time: %s\n", resp.Redemption.RedeemTime)
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleMyCoupons(client *APIClient) {
	flags := flag.NewFlagSet("my-coupons", flag.ExitOnError)
	userID := flags.String("user", "", "User ID")

	flags.Parse(os.Args[2:])

	if *userID == "" {
		fmt.Println("Error: user is required")
		flags.Usage()
		os.Exit(1)
	}

	resp, err := client.GetUserCoupons(*userID)
	if err != nil {
		fmt.Printf("Error getting user coupons: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
		fmt.Printf("Available coupons for user %s: %d\n", *userID, len(resp.Coupons))
		for i, coupon := range resp.Coupons {
			fmt.Printf("\nCoupon %d:\n", i+1)
			fmt.Printf("  Claim ID: %s\n", coupon.ID)
			fmt.Printf("  Batch ID: %s\n", coupon.BatchID)
			fmt.Printf("  Claim Time: %s\n", coupon.ClaimTime)
			fmt.Printf("  Is Redeemed: %t\n", coupon.IsRedeemed)
		}
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleReturn(client *APIClient) {
	flags := flag.NewFlagSet("return", flag.ExitOnError)
	userID := flags.String("user", "", "User ID")
	claimID := flags.String("claim", "", "Claim ID")

	flags.Parse(os.Args[2:])

	if *userID == "" || *claimID == "" {
		fmt.Println("Error: user and claim are required")
		flags.Usage()
		os.Exit(1)
	}

	req := &api.ReturnCouponRequest{
		UserID:  *userID,
		ClaimID: *claimID,
	}

	resp, err := client.ReturnCoupon(req)
	if err != nil {
		fmt.Printf("Error returning coupon: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}
}
