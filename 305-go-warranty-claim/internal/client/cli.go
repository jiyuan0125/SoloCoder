package client

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
	"warranty-claim/pkg/common"
)

type CLI struct {
	apiClient *APIClient
	reader    *bufio.Reader
}

func NewCLI(apiClient *APIClient) *CLI {
	return &CLI{
		apiClient: apiClient,
		reader:    bufio.NewReader(os.Stdin),
	}
}

func (c *CLI) prompt(prompt string) string {
	fmt.Print(prompt)
	input, _ := c.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (c *CLI) Run() {
	if len(os.Args) < 2 {
		c.printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "submit":
		c.cmdSubmit()
	case "status":
		c.cmdStatus()
	case "list":
		c.cmdList()
	case "appeal":
		c.cmdAppeal()
	case "admin-list":
		c.cmdAdminList()
	case "review":
		c.cmdReview()
	case "admin-appeals":
		c.cmdAdminAppeals()
	case "resolve-appeal":
		c.cmdResolveAppeal()
	case "stats":
		c.cmdStats()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		c.printUsage()
	}
}

func (c *CLI) printUsage() {
	fmt.Println("Warranty Claim Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client submit          - Submit a new warranty application")
	fmt.Println("  client status <id>     - Check status of an application")
	fmt.Println("  client list            - List all your applications")
	fmt.Println("  client appeal <id>     - Appeal a reviewed application")
	fmt.Println()
	fmt.Println("Admin Commands:")
	fmt.Println("  client admin-list      - List all pending applications")
	fmt.Println("  client review <id>     - Review an application")
	fmt.Println("  client admin-appeals   - List all pending appeals")
	fmt.Println("  client resolve-appeal  - Resolve an appeal")
	fmt.Println("  client stats           - Show statistics")
}

func (c *CLI) cmdSubmit() {
	serial := c.prompt("Product Serial Number (12 chars): ")
	purchaseDate := c.prompt("Purchase Date (YYYY-MM-DD): ")
	desc := c.prompt("Fault Description: ")

	req := common.SubmitApplicationRequest{
		UserID:       c.apiClient.userID,
		SerialNumber: serial,
		PurchaseDate: purchaseDate,
		Description:  desc,
	}

	resp, err := c.apiClient.SubmitApplication(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Println("Application submitted successfully!")
		fmt.Printf("Application ID: %s\n", resp.ApplicationID)
		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Printf("Message: %s\n", resp.Message)
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
	}
}

func (c *CLI) cmdStatus() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: client status <application-id>")
		return
	}

	appID := os.Args[2]
	app, err := c.apiClient.GetApplication(appID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printApplication(app)

	appeals, err := c.apiClient.GetAppealsByApplication(appID)
	if err == nil && len(appeals) > 0 {
		fmt.Println("\nAppeal History:")
		for _, aa := range appeals {
			printAppeal(&aa.Appeal)
		}
	}
}

func (c *CLI) cmdList() {
	apps, err := c.apiClient.GetMyApplications()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(apps) == 0 {
		fmt.Println("No applications found.")
		return
	}

	fmt.Printf("Found %d application(s):\n\n", len(apps))
	for _, app := range apps {
		printApplication(&app)
		fmt.Println("---")
	}
}

func (c *CLI) cmdAppeal() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: client appeal <application-id>")
		return
	}

	appID := os.Args[2]
	reason := c.prompt("Appeal Reason: ")

	req := common.SubmitAppealRequest{
		ApplicationID: appID,
		Reason:        reason,
	}

	success, msg, err := c.apiClient.SubmitAppeal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if success {
		fmt.Println("Appeal submitted successfully!")
	} else {
		fmt.Printf("Failed: %s\n", msg)
	}
}

func (c *CLI) cmdAdminList() {
	apps, err := c.apiClient.GetPendingApplications()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(apps) == 0 {
		fmt.Println("No pending applications.")
		return
	}

	fmt.Printf("Found %d pending application(s):\n\n", len(apps))
	for _, app := range apps {
		printApplication(&app)
		fmt.Println("---")
	}
}

func (c *CLI) cmdReview() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: client review <application-id>")
		return
	}

	appID := os.Args[2]
	
	app, err := c.apiClient.GetApplication(appID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printApplication(app)
	fmt.Println()

	action := c.prompt("Approve? (yes/no): ")
	approved := strings.ToLower(action) == "yes" || strings.ToLower(action) == "y"

	var reason string
	if !approved {
		reason = c.prompt("Rejection Reason: ")
	}

	req := common.ReviewApplicationRequest{
		ApplicationID: appID,
		Approved:      approved,
		Reason:        reason,
	}

	success, msg, err := c.apiClient.ReviewApplication(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if success {
		fmt.Println("Review completed successfully!")
	} else {
		fmt.Printf("Failed: %s\n", msg)
	}
}

func (c *CLI) cmdAdminAppeals() {
	appeals, err := c.apiClient.GetPendingAppeals()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(appeals) == 0 {
		fmt.Println("No pending appeals.")
		return
	}

	fmt.Printf("Found %d pending appeal(s):\n\n", len(appeals))
	for _, aa := range appeals {
		fmt.Println("=== Application ===")
		printApplication(&aa.Application)
		fmt.Println("=== Appeal ===")
		printAppeal(&aa.Appeal)
		fmt.Println("---")
	}
}

func (c *CLI) cmdResolveAppeal() {
	appealID := c.prompt("Appeal ID: ")
	action := c.prompt("Resolve in favor? (yes/no): ")
	resolved := strings.ToLower(action) == "yes" || strings.ToLower(action) == "y"
	note := c.prompt("Admin Note: ")

	success, msg, err := c.apiClient.ResolveAppeal(appealID, resolved, note)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if success {
		fmt.Println("Appeal resolved successfully!")
	} else {
		fmt.Printf("Failed: %s\n", msg)
	}
}

func (c *CLI) cmdStats() {
	stats, err := c.apiClient.GetStatistics()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("=== Warranty Claim Statistics ===")
	fmt.Printf("Total Applications: %d\n", stats.TotalApplications)
	fmt.Printf("Pending:            %d\n", stats.PendingApplications)
	fmt.Printf("Approved:           %d\n", stats.ApprovedApplications)
	fmt.Printf("Rejected:           %d\n", stats.RejectedApplications)
	fmt.Printf("Auto-Approved:      %d\n", stats.AutoApprovedApplications)
	fmt.Printf("Appealed:           %d\n", stats.AppealedApplications)
}

func printApplication(app *common.WarrantyApplication) {
	fmt.Printf("Application ID: %s\n", app.ID)
	fmt.Printf("User ID:        %s\n", app.UserID)
	fmt.Printf("Serial Number:  %s\n", app.SerialNumber)
	fmt.Printf("Purchase Date:  %s\n", app.PurchaseDate)
	fmt.Printf("Warranty Expiry:%s\n", app.WarrantyExpiry)
	fmt.Printf("Status:         %s\n", app.Status)
	if app.RejectReason != "" {
		fmt.Printf("Reject Reason:  %s\n", app.RejectReason)
	}
	fmt.Printf("Submitted At:   %s\n", time.Unix(app.SubmittedAt, 0).Format("2006-01-02 15:04:05"))
	if app.ApprovedAt > 0 {
		fmt.Printf("Approved At:    %s\n", time.Unix(app.ApprovedAt, 0).Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("Description:    %s\n", app.Description)
}

func printAppeal(appeal *common.Appeal) {
	fmt.Printf("Appeal ID:      %s\n", appeal.ID)
	fmt.Printf("Status:         %s\n", appeal.Status)
	fmt.Printf("Submitted At:   %s\n", time.Unix(appeal.SubmittedAt, 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("Reason:         %s\n", appeal.Reason)
	if appeal.AdminNote != "" {
		fmt.Printf("Admin Note:     %s\n", appeal.AdminNote)
	}
	if appeal.ResolvedAt > 0 {
		fmt.Printf("Resolved At:    %s\n", time.Unix(appeal.ResolvedAt, 0).Format("2006-01-02 15:04:05"))
	}
}
