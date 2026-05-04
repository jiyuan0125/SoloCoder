package main

import (
	"bufio"
	"fmt"
	"os"
	"return-exchange/internal/client"
	"return-exchange/pkg/common"
	"return-exchange/pkg/models"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	apiClient  *client.APIClient
	serverURL  string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "rex",
		Short: "Return & Exchange Command Line Client",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			apiClient = client.NewAPIClient(serverURL)
		},
	}

	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8080", "Server URL")

	rootCmd.AddCommand(createMockOrderCmd())
	rootCmd.AddCommand(userCmd())
	rootCmd.AddCommand(adminCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createMockOrderCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mock-order",
		Short: "Create a mock order for testing",
		Run: func(cmd *cobra.Command, args []string) {
			orderID, err := apiClient.CreateMockOrder()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("Mock order created successfully!\n")
			fmt.Printf("Order ID: %s\n", orderID)
		},
	}
}

func userCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "User operations",
	}

	cmd.AddCommand(userSubmitReturnCmd())
	cmd.AddCommand(userSubmitExchangeCmd())
	cmd.AddCommand(userListCmd())
	cmd.AddCommand(userStatusCmd())

	return cmd
}

func userSubmitReturnCmd() *cobra.Command {
	var orderID, userID, reason string
	var evidenceImages []string

	cmd := &cobra.Command{
		Use:   "return",
		Short: "Submit a return application",
		Run: func(cmd *cobra.Command, args []string) {
			if orderID == "" || userID == "" || reason == "" {
				orderID, userID, reason, evidenceImages = interactiveReturnInput()
			}

			req := common.SubmitReturnRequest{
				OrderID:        orderID,
				UserID:         userID,
				Reason:         models.ReturnReason(reason),
				EvidenceImages: evidenceImages,
			}

			appID, err := apiClient.SubmitReturn(req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("Return application submitted successfully!\n")
			fmt.Printf("Application ID: %s\n", appID)
		},
	}

	cmd.Flags().StringVar(&orderID, "order", "", "Order ID")
	cmd.Flags().StringVar(&userID, "user", "", "User ID")
	cmd.Flags().StringVar(&reason, "reason", "", "Return reason (quality_issue, damaged, wrong_size, not_wanted, wrong_order, description_mismatch)")
	cmd.Flags().StringSliceVar(&evidenceImages, "evidence", []string{}, "Evidence images (base64 encoded, can be specified multiple times)")

	return cmd
}

func userSubmitExchangeCmd() *cobra.Command {
	var orderID, userID, reason, newSKU, newSpec string
	var newSKUPrice float64
	var evidenceImages []string

	cmd := &cobra.Command{
		Use:   "exchange",
		Short: "Submit an exchange application",
		Run: func(cmd *cobra.Command, args []string) {
			if orderID == "" || userID == "" || reason == "" || newSKU == "" {
				orderID, userID, reason, newSKU, newSpec, newSKUPrice, evidenceImages = interactiveExchangeInput()
			}

			req := common.SubmitExchangeRequest{
				OrderID:          orderID,
				UserID:           userID,
				Reason:           models.ReturnReason(reason),
				EvidenceImages:   evidenceImages,
				NewSKU:           newSKU,
				NewSpecification: newSpec,
				NewSKUPrice:      newSKUPrice,
			}

			appID, err := apiClient.SubmitExchange(req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("Exchange application submitted successfully!\n")
			fmt.Printf("Application ID: %s\n", appID)
		},
	}

	cmd.Flags().StringVar(&orderID, "order", "", "Order ID")
	cmd.Flags().StringVar(&userID, "user", "", "User ID")
	cmd.Flags().StringVar(&reason, "reason", "", "Return reason")
	cmd.Flags().StringVar(&newSKU, "new-sku", "", "New SKU for exchange")
	cmd.Flags().StringVar(&newSpec, "new-spec", "", "New specification")
	cmd.Flags().Float64Var(&newSKUPrice, "new-price", 0, "Price of new SKU")
	cmd.Flags().StringSliceVar(&evidenceImages, "evidence", []string{}, "Evidence images")

	return cmd
}

func userListCmd() *cobra.Command {
	var userID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all applications for a user",
		Run: func(cmd *cobra.Command, args []string) {
			if userID == "" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter User ID: ")
				userID, _ = reader.ReadString('\n')
				userID = strings.TrimSpace(userID)
			}

			apps, err := apiClient.ListUserApplications(userID)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			if len(apps) == 0 {
				fmt.Println("No applications found.")
				return
			}

			fmt.Printf("\nApplications for user %s:\n", userID)
			fmt.Println("========================================")
			for _, app := range apps {
				printApplicationSummary(&app)
				fmt.Println("----------------------------------------")
			}
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "User ID")
	return cmd
}

func userStatusCmd() *cobra.Command {
	var appID string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check application status",
		Run: func(cmd *cobra.Command, args []string) {
			if appID == "" && len(args) > 0 {
				appID = args[0]
			}
			if appID == "" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter Application ID: ")
				appID, _ = reader.ReadString('\n')
				appID = strings.TrimSpace(appID)
			}

			app, err := apiClient.GetApplication(appID)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			printApplicationDetail(app)
		},
	}

	cmd.Flags().StringVar(&appID, "id", "", "Application ID")
	return cmd
}

func adminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Admin operations",
	}

	cmd.AddCommand(adminListCmd())
	cmd.AddCommand(adminReviewCmd())
	cmd.AddCommand(adminProcessRefundCmd())
	cmd.AddCommand(adminHandlePriceDiffCmd())
	cmd.AddCommand(adminCreateShippingCmd())
	cmd.AddCommand(adminCompleteCmd())

	return cmd
}

func adminHandlePriceDiffCmd() *cobra.Command {
	var appID string

	cmd := &cobra.Command{
		Use:   "handle-price-diff",
		Short: "Handle price difference for exchange applications",
		Long: `Handle price difference for exchange applications.
- If new SKU is cheaper: refund the difference to user
- If new SKU is more expensive: require user to pay the difference first`,
		Run: func(cmd *cobra.Command, args []string) {
			if appID == "" && len(args) > 0 {
				appID = args[0]
			}
			if appID == "" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter Application ID: ")
				appID, _ = reader.ReadString('\n')
				appID = strings.TrimSpace(appID)
			}

			app, err := apiClient.GetApplication(appID)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			if app.Type != models.ExchangeType {
				fmt.Println("Error: only exchange applications need price difference handling")
				return
			}

			if app.Status != models.StatusApproved {
				fmt.Printf("Error: application status must be 'approved' to handle price difference, current status: %s\n", app.Status)
				return
			}

			if app.PriceDifferenceHandled {
				fmt.Println("Price difference already handled!")
				if app.PriceDifferenceRefundID != "" {
					fmt.Printf("Refund ID: %s\n", app.PriceDifferenceRefundID)
				}
				return
			}

			if app.PriceDifference == 0 {
				fmt.Println("No price difference to handle (prices are equal)")
			} else if app.PriceDifference > 0 {
				fmt.Printf("New SKU is cheaper by %.2f. Will refund this difference to user.\n", app.PriceDifference)
			} else {
				fmt.Printf("New SKU is more expensive by %.2f. User needs to pay this difference first.\n", -app.PriceDifference)
			}

			if app.PriceDifference != 0 {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Continue? (yes/no): ")
				confirm, _ := reader.ReadString('\n')
				confirm = strings.TrimSpace(strings.ToLower(confirm))
				if confirm != "yes" && confirm != "y" {
					fmt.Println("Operation cancelled.")
					return
				}
			}

			req := common.HandlePriceDifferenceRequest{
				ApplicationID: appID,
			}

			result, err := apiClient.HandlePriceDifference(req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("\nPrice difference handled successfully!")
			switch result.Action {
			case "refund_difference":
				fmt.Printf("Action: Refund difference to user\n")
				fmt.Printf("Refund Amount: %.2f\n", result.Amount)
				fmt.Printf("Refund ID: %s\n", result.RefundID)
			case "require_payment":
				fmt.Printf("Action: Require user to pay difference\n")
				fmt.Printf("Amount to pay: %.2f\n", result.Amount)
				fmt.Println("Please collect payment from user before creating shipping order.")
			case "no_action":
				fmt.Printf("Action: No action needed (prices are equal)\n")
			}
		},
	}

	cmd.Flags().StringVar(&appID, "id", "", "Application ID")
	return cmd
}

func adminListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all applications",
		Run: func(cmd *cobra.Command, args []string) {
			apps, err := apiClient.ListAllApplications()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			if len(apps) == 0 {
				fmt.Println("No applications found.")
				return
			}

			fmt.Println("\nAll Applications:")
			fmt.Println("========================================")
			for _, app := range apps {
				printApplicationSummary(&app)
				fmt.Println("----------------------------------------")
			}
		},
	}
}

func adminReviewCmd() *cobra.Command {
	var appID string
	var approved bool
	var shippingFee float64

	cmd := &cobra.Command{
		Use:   "review",
		Short: "Review an application",
		Run: func(cmd *cobra.Command, args []string) {
			if appID == "" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter Application ID: ")
				appID, _ = reader.ReadString('\n')
				appID = strings.TrimSpace(appID)

				fmt.Print("Approve? (yes/no): ")
				approveStr, _ := reader.ReadString('\n')
				approveStr = strings.TrimSpace(strings.ToLower(approveStr))
				approved = approveStr == "yes" || approveStr == "y"

				fmt.Print("Shipping Fee (enter 0 if none): ")
				feeStr, _ := reader.ReadString('\n')
				feeStr = strings.TrimSpace(feeStr)
				if feeStr != "" {
					shippingFee, _ = strconv.ParseFloat(feeStr, 64)
				}
			}

			req := common.ReviewRequest{
				ApplicationID: appID,
				Approved:      approved,
				ShippingFee:   shippingFee,
			}

			err := apiClient.ReviewApplication(req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			if approved {
				fmt.Println("Application approved successfully!")
			} else {
				fmt.Println("Application rejected successfully!")
			}
		},
	}

	cmd.Flags().StringVar(&appID, "id", "", "Application ID")
	cmd.Flags().BoolVar(&approved, "approve", false, "Approve the application")
	cmd.Flags().Float64Var(&shippingFee, "shipping-fee", 0, "Shipping fee amount")

	return cmd
}

func adminProcessRefundCmd() *cobra.Command {
	var appID string

	cmd := &cobra.Command{
		Use:   "refund",
		Short: "Process refund for a return application",
		Run: func(cmd *cobra.Command, args []string) {
			if appID == "" && len(args) > 0 {
				appID = args[0]
			}
			if appID == "" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter Application ID: ")
				appID, _ = reader.ReadString('\n')
				appID = strings.TrimSpace(appID)
			}

			req := common.ProcessRefundRequest{
				ApplicationID: appID,
			}

			refundID, totalRefund, err := apiClient.ProcessRefund(req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("Refund processed successfully!")
			fmt.Printf("Refund ID: %s\n", refundID)
			fmt.Printf("Total Refund Amount: %.2f\n", totalRefund)
		},
	}

	cmd.Flags().StringVar(&appID, "id", "", "Application ID")
	return cmd
}

func adminCreateShippingCmd() *cobra.Command {
	var appID string

	cmd := &cobra.Command{
		Use:   "shipping",
		Short: "Create shipping order for an exchange application",
		Run: func(cmd *cobra.Command, args []string) {
			if appID == "" && len(args) > 0 {
				appID = args[0]
			}
			if appID == "" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter Application ID: ")
				appID, _ = reader.ReadString('\n')
				appID = strings.TrimSpace(appID)
			}

			req := common.CreateShippingOrderRequest{
				ApplicationID: appID,
			}

			shippingID, err := apiClient.CreateShippingOrder(req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("Shipping order created successfully!")
			fmt.Printf("Shipping Order ID: %s\n", shippingID)
		},
	}

	cmd.Flags().StringVar(&appID, "id", "", "Application ID")
	return cmd
}

func adminCompleteCmd() *cobra.Command {
	var appID string

	cmd := &cobra.Command{
		Use:   "complete",
		Short: "Mark an application as completed",
		Run: func(cmd *cobra.Command, args []string) {
			if appID == "" && len(args) > 0 {
				appID = args[0]
			}
			if appID == "" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter Application ID: ")
				appID, _ = reader.ReadString('\n')
				appID = strings.TrimSpace(appID)
			}

			err := apiClient.CompleteApplication(appID)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("Application marked as completed successfully!")
		},
	}

	cmd.Flags().StringVar(&appID, "id", "", "Application ID")
	return cmd
}

func interactiveReturnInput() (orderID, userID, reason string, evidence []string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Order ID: ")
	orderID, _ = reader.ReadString('\n')
	orderID = strings.TrimSpace(orderID)

	fmt.Print("Enter User ID: ")
	userID, _ = reader.ReadString('\n')
	userID = strings.TrimSpace(userID)

	fmt.Println("\nAvailable reasons:")
	fmt.Println("  quality_issue   - 质量问题")
	fmt.Println("  damaged         - 商品破损")
	fmt.Println("  wrong_size      - 尺寸不合适")
	fmt.Println("  not_wanted      - 不想要了")
	fmt.Println("  wrong_order     - 拍错")
	fmt.Println("  description_mismatch - 与描述不符")
	fmt.Print("\nEnter return reason: ")
	reason, _ = reader.ReadString('\n')
	reason = strings.TrimSpace(reason)

	if reason == "quality_issue" || reason == "damaged" {
		fmt.Print("Enter evidence image (base64, leave empty to skip): ")
		img, _ := reader.ReadString('\n')
		img = strings.TrimSpace(img)
		if img != "" {
			evidence = append(evidence, img)
		}
	}

	return
}

func interactiveExchangeInput() (orderID, userID, reason, newSKU, newSpec string, newPrice float64, evidence []string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Order ID: ")
	orderID, _ = reader.ReadString('\n')
	orderID = strings.TrimSpace(orderID)

	fmt.Print("Enter User ID: ")
	userID, _ = reader.ReadString('\n')
	userID = strings.TrimSpace(userID)

	fmt.Println("\nAvailable reasons:")
	fmt.Println("  quality_issue   - 质量问题")
	fmt.Println("  damaged         - 商品破损")
	fmt.Println("  wrong_size      - 尺寸不合适")
	fmt.Println("  not_wanted      - 不想要了")
	fmt.Println("  wrong_order     - 拍错")
	fmt.Println("  description_mismatch - 与描述不符")
	fmt.Print("\nEnter return reason: ")
	reason, _ = reader.ReadString('\n')
	reason = strings.TrimSpace(reason)

	if reason == "quality_issue" || reason == "damaged" {
		fmt.Print("Enter evidence image (base64, leave empty to skip): ")
		img, _ := reader.ReadString('\n')
		img = strings.TrimSpace(img)
		if img != "" {
			evidence = append(evidence, img)
		}
	}

	fmt.Print("Enter new SKU: ")
	newSKU, _ = reader.ReadString('\n')
	newSKU = strings.TrimSpace(newSKU)

	fmt.Print("Enter new specification: ")
	newSpec, _ = reader.ReadString('\n')
	newSpec = strings.TrimSpace(newSpec)

	fmt.Print("Enter new SKU price: ")
	priceStr, _ := reader.ReadString('\n')
	priceStr = strings.TrimSpace(priceStr)
	newPrice, _ = strconv.ParseFloat(priceStr, 64)

	return
}

func printApplicationSummary(app *models.ReturnExchangeApplication) {
	fmt.Printf("Application ID: %s\n", app.ID)
	fmt.Printf("  Type: %s\n", app.Type)
	fmt.Printf("  Order ID: %s\n", app.OrderID)
	fmt.Printf("  Status: %s\n", app.Status)
	fmt.Printf("  Reason: %s\n", app.Reason)
	fmt.Printf("  Created: %s\n", app.CreateTime.Format("2006-01-02 15:04:05"))
}

func printApplicationDetail(app *models.ReturnExchangeApplication) {
	fmt.Println("\nApplication Details:")
	fmt.Println("========================================")
	fmt.Printf("Application ID: %s\n", app.ID)
	fmt.Printf("Order ID: %s\n", app.OrderID)
	fmt.Printf("User ID: %s\n", app.UserID)
	fmt.Printf("Type: %s\n", app.Type)
	fmt.Printf("Status: %s\n", app.Status)
	fmt.Printf("Reason: %s\n", app.Reason)

	if app.Type == models.ExchangeType {
		fmt.Printf("New SKU: %s\n", app.NewSKU)
		fmt.Printf("New Specification: %s\n", app.NewSpecification)
		fmt.Printf("New SKU Price: %.2f\n", app.NewSKUPrice)
		fmt.Printf("Price Difference: %.2f\n", app.PriceDifference)
		fmt.Printf("Price Difference Handled: %v\n", app.PriceDifferenceHandled)
		if app.PriceDifferenceRefundID != "" {
			fmt.Printf("Price Difference Refund ID: %s\n", app.PriceDifferenceRefundID)
		}
	}

	if app.RefundID != "" {
		fmt.Printf("Refund ID: %s\n", app.RefundID)
		fmt.Printf("Refund Amount: %.2f\n", app.RefundAmount)
	}

	if app.ShippingOrderID != "" {
		fmt.Printf("Shipping Order ID: %s\n", app.ShippingOrderID)
	}

	fmt.Printf("Shipping Fee: %.2f\n", app.ShippingFee)
	fmt.Printf("Shipping Fee Paid By: %s\n", app.ShippingFeePaidBy)

	if len(app.EvidenceImages) > 0 {
		fmt.Printf("Evidence Images: %d images\n", len(app.EvidenceImages))
	}

	fmt.Printf("Created: %s\n", app.CreateTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", app.UpdateTime.Format("2006-01-02 15:04:05"))
}
