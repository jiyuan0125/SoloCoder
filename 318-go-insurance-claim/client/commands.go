package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"insurance-claim/common"
)

type CLI struct {
	client *APIClient
	reader *bufio.Reader
}

func NewCLI(client *APIClient) *CLI {
	return &CLI{
		client: client,
		reader: bufio.NewReader(os.Stdin),
	}
}

func (c *CLI) readInput(prompt string) string {
	fmt.Print(prompt)
	input, _ := c.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (c *CLI) readDate(prompt string) (time.Time, error) {
	for {
		input := c.readInput(prompt)
		if input == "" {
			return time.Time{}, fmt.Errorf("date is required")
		}
		t, err := time.Parse("2006-01-02", input)
		if err == nil {
			return t, nil
		}
		fmt.Println("Invalid date format. Please use YYYY-MM-DD")
	}
}

func (c *CLI) readFloat(prompt string) (float64, error) {
	for {
		input := c.readInput(prompt)
		if input == "" {
			return 0, fmt.Errorf("input is required")
		}
		f, err := strconv.ParseFloat(input, 64)
		if err == nil {
			return f, nil
		}
		fmt.Println("Invalid number. Please try again.")
	}
}

func (c *CLI) printPolicy(p *common.Policy) {
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Policy Number: %s\n", p.PolicyNumber)
	fmt.Printf("Type: %s\n", p.PolicyType)
	fmt.Printf("Total Amount: %.2f\n", p.TotalAmount)
	fmt.Printf("Remaining Amount: %.2f\n", p.RemainingAmount)
	fmt.Printf("Effective Date: %s\n", p.EffectiveDate.Format("2006-01-02"))
	fmt.Printf("Status: %s\n", p.Status)
	fmt.Println("--------------------------------------------------")
}

func (c *CLI) printClaim(claim *common.Claim) {
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Claim ID: %s\n", claim.ClaimID)
	fmt.Printf("Policy Number: %s\n", claim.PolicyNumber)
	fmt.Printf("Incident Date: %s\n", claim.IncidentDate.Format("2006-01-02"))
	fmt.Printf("Incident Reason: %s\n", claim.IncidentReason)
	fmt.Printf("Requested Amount: %.2f\n", claim.RequestedAmount)
	if claim.LiabilityRatio != "" {
		fmt.Printf("Liability Ratio: %s\n", claim.LiabilityRatio)
	}
	fmt.Printf("Payout Percentage: %.0f%%\n", claim.PayoutPercentage*100)
	fmt.Printf("Payout Amount: %.2f\n", claim.PayoutAmount)
	fmt.Printf("Status: %s\n", claim.Status)
	if claim.RejectReason != "" {
		fmt.Printf("Reject Reason: %s\n", claim.RejectReason)
	}
	fmt.Printf("Created At: %s\n", claim.CreatedAt.Format("2006-01-02 15:04:05"))
	if claim.ReviewedAt != nil {
		fmt.Printf("Reviewed At: %s\n", claim.ReviewedAt.Format("2006-01-02 15:04:05"))
	}
	if claim.PaidAt != nil {
		fmt.Printf("Paid At: %s\n", claim.PaidAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Println("--------------------------------------------------")
}

func (c *CLI) CreatePolicy() {
	fmt.Println("\n=== Create New Policy ===")

	policyNumber := c.readInput("Policy Number: ")
	if policyNumber == "" {
		fmt.Println("Error: Policy number cannot be empty")
		return
	}

	fmt.Println("\nPolicy Types:")
	fmt.Println("  1. Medical Insurance (medical)")
	fmt.Println("  2. Auto Insurance (auto)")
	fmt.Println("  3. Accident Insurance (accident)")
	typeInput := c.readInput("Select Policy Type (1-3): ")

	var policyType common.PolicyType
	switch typeInput {
	case "1":
		policyType = common.PolicyTypeMedical
	case "2":
		policyType = common.PolicyTypeAuto
	case "3":
		policyType = common.PolicyTypeAccident
	default:
		fmt.Println("Invalid policy type")
		return
	}

	totalAmount, err := c.readFloat("Total Amount: ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	effectiveDate, err := c.readDate("Effective Date (YYYY-MM-DD): ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	req := common.CreatePolicyRequest{
		PolicyNumber:  policyNumber,
		PolicyType:    policyType,
		TotalAmount:   totalAmount,
		EffectiveDate: effectiveDate,
	}

	policy, err := c.client.CreatePolicy(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("\nPolicy created successfully!")
	c.printPolicy(policy)
}

func (c *CLI) SubmitClaim() {
	fmt.Println("\n=== Submit Claim ===")

	policyNumber := c.readInput("Policy Number: ")
	if policyNumber == "" {
		fmt.Println("Error: Policy number cannot be empty")
		return
	}

	incidentDate, err := c.readDate("Incident Date (YYYY-MM-DD): ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	incidentReason := c.readInput("Incident Reason: ")
	if incidentReason == "" {
		fmt.Println("Error: Incident reason cannot be empty")
		return
	}

	if len(incidentReason) > common.MaxReasonLength {
		fmt.Printf("Error: Incident reason cannot exceed %d characters\n", common.MaxReasonLength)
		return
	}

	requestedAmount, err := c.readFloat("Requested Amount: ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var liabilityRatio common.LiabilityRatio
	fmt.Println("\nIs this an auto insurance claim?")
	fmt.Println("  1. Yes")
	fmt.Println("  2. No")
	autoChoice := c.readInput("Select (1-2): ")

	if autoChoice == "1" {
		fmt.Println("\nLiability Ratios:")
		fmt.Println("  1. Full (100%)")
		fmt.Println("  2. Main (70%)")
		fmt.Println("  3. Equal (50%)")
		fmt.Println("  4. Secondary (30%)")
		fmt.Println("  5. None (0%)")
		ratioChoice := c.readInput("Select Liability Ratio (1-5): ")

		switch ratioChoice {
		case "1":
			liabilityRatio = common.LiabilityFull
		case "2":
			liabilityRatio = common.LiabilityMain
		case "3":
			liabilityRatio = common.LiabilityEqual
		case "4":
			liabilityRatio = common.LiabilitySecondary
		case "5":
			liabilityRatio = common.LiabilityNone
		default:
			fmt.Println("Invalid liability ratio")
			return
		}
	}

	req := common.SubmitClaimRequest{
		PolicyNumber:    policyNumber,
		IncidentDate:    incidentDate,
		IncidentReason:  incidentReason,
		RequestedAmount: requestedAmount,
		LiabilityRatio:  liabilityRatio,
	}

	claim, err := c.client.SubmitClaim(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("\nClaim submitted successfully!")
	c.printClaim(claim)
}

func (c *CLI) GetClaimStatus() {
	fmt.Println("\n=== Check Claim Status ===")

	claimID := c.readInput("Claim ID: ")
	if claimID == "" {
		fmt.Println("Error: Claim ID cannot be empty")
		return
	}

	claim, err := c.client.GetClaim(claimID)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("\nClaim Status:")
	c.printClaim(claim)
}

func (c *CLI) GetPolicyClaims() {
	fmt.Println("\n=== View Policy Claims History ===")

	policyNumber := c.readInput("Policy Number: ")
	if policyNumber == "" {
		fmt.Println("Error: Policy number cannot be empty")
		return
	}

	claims, err := c.client.GetPolicyClaims(policyNumber)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(claims) == 0 {
		fmt.Println("No claims found for this policy")
		return
	}

	fmt.Printf("\nFound %d claim(s) for policy %s:\n", len(claims), policyNumber)
	for _, claim := range claims {
		c.printClaim(claim)
	}
}

func (c *CLI) ListPendingClaims() {
	fmt.Println("\n=== Pending Claims (For Admin) ===")

	claims, err := c.client.GetPendingClaims()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(claims) == 0 {
		fmt.Println("No pending claims")
		return
	}

	fmt.Printf("\nFound %d pending claim(s):\n", len(claims))
	for _, claim := range claims {
		c.printClaim(claim)
	}
}

func (c *CLI) ListApprovedClaims() {
	fmt.Println("\n=== Approved Claims (For Admin) ===")

	claims, err := c.client.GetApprovedClaims()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(claims) == 0 {
		fmt.Println("No approved claims")
		return
	}

	fmt.Printf("\nFound %d approved claim(s):\n", len(claims))
	for _, claim := range claims {
		c.printClaim(claim)
	}
}

func (c *CLI) ReviewClaim() {
	fmt.Println("\n=== Review Claim (For Admin) ===")

	claimID := c.readInput("Claim ID: ")
	if claimID == "" {
		fmt.Println("Error: Claim ID cannot be empty")
		return
	}

	claim, err := c.client.GetClaim(claimID)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if claim.Status != common.ClaimStatusPending {
		fmt.Println("Error: Only pending claims can be reviewed")
		return
	}

	fmt.Println("\nClaim Details:")
	c.printClaim(claim)

	fmt.Println("\n1. Approve")
	fmt.Println("2. Reject")
	choice := c.readInput("Select action (1-2): ")

	var approved bool
	var rejectReason string

	if choice == "1" {
		approved = true
	} else if choice == "2" {
		approved = false
		rejectReason = c.readInput("Reject Reason: ")
		if rejectReason == "" {
			fmt.Println("Error: Reject reason is required")
			return
		}
	} else {
		fmt.Println("Invalid choice")
		return
	}

	req := common.ReviewClaimRequest{
		ClaimID:      claimID,
		Approved:     approved,
		RejectReason: rejectReason,
	}

	updatedClaim, err := c.client.ReviewClaim(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("\nClaim reviewed successfully!")
	c.printClaim(updatedClaim)
}

func (c *CLI) ConfirmPayment() {
	fmt.Println("\n=== Confirm Payment (For Admin) ===")

	claimID := c.readInput("Claim ID: ")
	if claimID == "" {
		fmt.Println("Error: Claim ID cannot be empty")
		return
	}

	claim, err := c.client.GetClaim(claimID)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if claim.Status != common.ClaimStatusApproved {
		fmt.Println("Error: Only approved claims can be paid")
		return
	}

	fmt.Println("\nClaim Details:")
	c.printClaim(claim)

	confirm := c.readInput("\nConfirm payment? (y/n): ")
	if strings.ToLower(confirm) != "y" {
		fmt.Println("Payment cancelled")
		return
	}

	req := common.ConfirmPaymentRequest{
		ClaimID: claimID,
	}

	updatedClaim, err := c.client.ConfirmPayment(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("\nPayment confirmed successfully!")
	c.printClaim(updatedClaim)
}

func (c *CLI) ListAllPolicies() {
	fmt.Println("\n=== All Policies ===")

	policies, err := c.client.ListAllPolicies()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(policies) == 0 {
		fmt.Println("No policies found")
		return
	}

	fmt.Printf("\nFound %d policy(ies):\n", len(policies))
	for _, p := range policies {
		c.printPolicy(p)
	}
}

func (c *CLI) ListAllClaims() {
	fmt.Println("\n=== All Claims ===")

	claims, err := c.client.ListAllClaims()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(claims) == 0 {
		fmt.Println("No claims found")
		return
	}

	fmt.Printf("\nFound %d claim(s):\n", len(claims))
	for _, claim := range claims {
		c.printClaim(claim)
	}
}
