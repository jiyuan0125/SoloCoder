package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"insurance-claim/pkg/common"
)

const defaultServerURL = "http://localhost:9003"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := os.Getenv("CLAIM_SERVER_URL")
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	command := os.Args[1]
	args := os.Args[2:]

	client := NewAPIClient(serverURL)

	var err error
	switch command {
	case "create-policy":
		err = cmdCreatePolicy(client, args)
	case "submit-claim":
		err = cmdSubmitClaim(client, args)
	case "list":
		err = cmdListClaims(client)
	case "get":
		err = cmdGetCase(client, args)
	case "assign-investigator":
		err = cmdAssignInvestigator(client, args)
	case "submit-investigation":
		err = cmdSubmitInvestigation(client, args)
	case "assign-assessor":
		err = cmdAssignAssessor(client, args)
	case "submit-assessment":
		err = cmdSubmitAssessment(client, args)
	case "calculate-payout":
		err = cmdCalculatePayout(client, args)
	case "approve-payout":
		err = cmdApprovePayout(client, args)
	case "mark-paid":
		err = cmdMarkAsPaid(client, args)
	case "flag-review":
		err = cmdFlagForReview(client, args)
	case "resolve-review":
		err = cmdResolveReview(client, args)
	case "help", "-h", "--help":
		printUsage()
		return
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Insurance Claim Management Client

Usage:
  claim-client <command> [arguments]

Commands:
  create-policy --policy-no <no> --limit <amount>
  submit-claim --policy-no <no> --accident-time <time> --location <loc> --type <type> --desc <desc> --amount <amt>
  list
  get <case_id|case_no>
  assign-investigator --case-id <id> --investigator-id <id> --investigator-name <name> --operator <op>
  submit-investigation --case-id <id> --investigator-id <id> --investigator-name <name> --cause <cause> --liability <liability> --damaged <parts> [--has-valid-reason] --reason-desc <desc>
  assign-assessor --case-id <id> --assessor-id <id> --assessor-name <name> --operator <op>
  submit-assessment --case-id <id> --assessor-id <id> --assessor-name <name> --items <json>
  calculate-payout <case_id>
  approve-payout --case-id <id> --operator <op>
  mark-paid --case-id <id> --operator <op>
  flag-review --case-id <id> --operator <op> --reason <reason>
  resolve-review --case-id <id> --operator <op> --reject --reason <reason>

Accident Types: traffic, natural, theft, glass, scratch
Liability Types: full, major, equal, minor, none
Repair Types: repair, replace

Examples:
  claim-client create-policy --policy-no POL001 --limit 100000
  claim-client submit-claim --policy-no POL001 --accident-time "2026-05-10T08:00:00" --location "Beijing" --type traffic --desc "Car accident" --amount 5000
  claim-client list
  claim-client get CL20260511000001`)
}

func parseFlags(args []string, fs *flag.FlagSet) error {
	return fs.Parse(args)
}

func cmdCreatePolicy(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("create-policy", flag.ExitOnError)
	policyNo := fs.String("policy-no", "", "Policy number")
	limit := fs.Int64("limit", 0, "Policy limit")
	if err := parseFlags(args, fs); err != nil {
		return err
	}

	if *policyNo == "" || *limit <= 0 {
		return fmt.Errorf("policy-no and limit are required")
	}

	policy := &common.Policy{
		PolicyNo: *policyNo,
		Limit:    *limit,
	}

	if err := client.CreatePolicy(policy); err != nil {
		return err
	}

	fmt.Printf("Policy created: %s (limit: %d)\n", *policyNo, *limit)
	return nil
}

func cmdSubmitClaim(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("submit-claim", flag.ExitOnError)
	policyNo := fs.String("policy-no", "", "Policy number")
	accidentTimeStr := fs.String("accident-time", "", "Accident time (RFC3339)")
	location := fs.String("location", "", "Accident location")
	accidentType := fs.String("type", "", "Accident type")
	desc := fs.String("desc", "", "Accident description")
	amount := fs.Int64("amount", 0, "Estimated amount")
	if err := parseFlags(args, fs); err != nil {
		return err
	}

	if *policyNo == "" || *accidentTimeStr == "" || *location == "" || *accidentType == "" || *amount <= 0 {
		return fmt.Errorf("all fields are required")
	}

	accidentTime, err := time.Parse(time.RFC3339, *accidentTimeStr)
	if err != nil {
		return fmt.Errorf("invalid accident time: %v", err)
	}

	req := &common.SubmitClaimRequest{
		PolicyNo:            *policyNo,
		AccidentTime:        accidentTime,
		AccidentLocation:    *location,
		AccidentType:        common.AccidentType(*accidentType),
		AccidentDescription: *desc,
		EstimatedAmount:     *amount,
	}

	result, err := client.SubmitClaim(req)
	if err != nil {
		return err
	}

	fmt.Printf("Claim submitted successfully!\n")
	fmt.Printf("  Case ID: %s\n", result.CaseID)
	fmt.Printf("  Case No: %s\n", result.CaseNo)
	return nil
}

func cmdListClaims(client *APIClient) error {
	result, err := client.ListClaims()
	if err != nil {
		return err
	}

	if len(result.Cases) == 0 {
		fmt.Println("No claims found")
		return nil
	}

	fmt.Printf("Found %d claim(s):\n\n", len(result.Cases))
	for _, c := range result.Cases {
		fmt.Printf("Case No: %s\n", c.CaseNo)
		fmt.Printf("  ID: %s\n", c.ID)
		fmt.Printf("  Policy: %s\n", c.PolicyNo)
		fmt.Printf("  Status: %s\n", c.Status.String())
		fmt.Printf("  Location: %s\n", c.AccidentLocation)
		fmt.Printf("  Type: %s\n", c.AccidentType.String())
		fmt.Println()
	}
	return nil
}

func cmdGetCase(client *APIClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("case ID or case no is required")
	}

	result, err := client.GetCase(args[0])
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(result.Case, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func cmdAssignInvestigator(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("assign-investigator", flag.ExitOnError)
	caseID := fs.String("case-id", "", "Case ID")
	investigatorID := fs.String("investigator-id", "", "Investigator ID")
	investigatorName := fs.String("investigator-name", "", "Investigator name")
	operator := fs.String("operator", "", "Operator")
	if err := parseFlags(args, fs); err != nil {
		return err
	}

	if *caseID == "" || *investigatorID == "" || *investigatorName == "" || *operator == "" {
		return fmt.Errorf("all fields are required")
	}

	req := &common.AssignInvestigatorRequest{
		CaseID:           *caseID,
		InvestigatorID:   *investigatorID,
		InvestigatorName: *investigatorName,
		Operator:         *operator,
	}

	if err := client.AssignInvestigator(req); err != nil {
		return err
	}

	fmt.Println("Investigator assigned successfully")
	return nil
}

func cmdSubmitInvestigation(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("submit-investigation", flag.ExitOnError)
	caseID := fs.String("case-id", "", "Case ID")
	investigatorID := fs.String("investigator-id", "", "Investigator ID")
	investigatorName := fs.String("investigator-name", "", "Investigator name")
	cause := fs.String("cause", "", "Cause analysis")
	liability := fs.String("liability", "", "Liability type")
	damaged := fs.String("damaged", "", "Damaged parts")
	hasValidReason := fs.Bool("has-valid-reason", false, "Has valid reason for late reporting")
	reasonDesc := fs.String("reason-desc", "", "Reason description")
	if err := parseFlags(args, fs); err != nil {
		return err
	}

	if *caseID == "" || *investigatorID == "" || *investigatorName == "" || *cause == "" || *liability == "" || *damaged == "" {
		return fmt.Errorf("required fields: case-id, investigator-id, investigator-name, cause, liability, damaged")
	}

	req := &common.SubmitInvestigationRequest{
		CaseID:            *caseID,
		InvestigatorID:    *investigatorID,
		InvestigatorName:  *investigatorName,
		CauseAnalysis:     *cause,
		Liability:         common.LiabilityType(*liability),
		DamagedParts:      *damaged,
		HasValidReason:    *hasValidReason,
		ReasonDescription: *reasonDesc,
	}

	if err := client.SubmitInvestigation(req); err != nil {
		return err
	}

	fmt.Println("Investigation report submitted successfully")
	return nil
}

func cmdAssignAssessor(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("assign-assessor", flag.ExitOnError)
	caseID := fs.String("case-id", "", "Case ID")
	assessorID := fs.String("assessor-id", "", "Assessor ID")
	assessorName := fs.String("assessor-name", "", "Assessor name")
	operator := fs.String("operator", "", "Operator")
	if err := parseFlags(args, fs); err != nil {
		return err
	}

	if *caseID == "" || *assessorID == "" || *assessorName == "" || *operator == "" {
		return fmt.Errorf("all fields are required")
	}

	req := &common.AssignAssessorRequest{
		CaseID:       *caseID,
		AssessorID:   *assessorID,
		AssessorName: *assessorName,
		Operator:     *operator,
	}

	if err := client.AssignAssessor(req); err != nil {
		return err
	}

	fmt.Println("Assessor assigned successfully")
	return nil
}

func cmdSubmitAssessment(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("submit-assessment", flag.ExitOnError)
	caseID := fs.String("case-id", "", "Case ID")
	assessorID := fs.String("assessor-id", "", "Assessor ID")
	assessorName := fs.String("assessor-name", "", "Assessor name")
	itemsJSON := fs.String("items", "", "Damage items as JSON array")
	if err := parseFlags(args, fs); err != nil {
		return err
	}

	if *caseID == "" || *assessorID == "" || *assessorName == "" || *itemsJSON == "" {
		return fmt.Errorf("all fields are required")
	}

	var items []common.DamageItemRequest
	if err := json.Unmarshal([]byte(*itemsJSON), &items); err != nil {
		return fmt.Errorf("invalid items JSON: %v", err)
	}

	req := &common.SubmitAssessmentRequest{
		CaseID:       *caseID,
		AssessorID:   *assessorID,
		AssessorName: *assessorName,
		Items:        items,
	}

	if err := client.SubmitAssessment(req); err != nil {
		return err
	}

	fmt.Println("Assessment submitted successfully")
	return nil
}

func cmdCalculatePayout(client *APIClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("case ID is required")
	}

	result, err := client.CalculatePayout(args[0])
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(result.Payout, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func cmdApprovePayout(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("approve-payout", flag.ExitOnError)
	caseID := fs.String("case-id", "", "Case ID")
	operator := fs.String("operator", "", "Operator")
	if err := parseFlags(args, fs); err != nil {
		return err
	}

	if *caseID == "" || *operator == "" {
		return fmt.Errorf("case-id and operator are required")
	}

	if err := client.ApprovePayout(*caseID, *operator); err != nil {
		return err
	}

	fmt.Println("Payout approved successfully")
	return nil
}

func cmdMarkAsPaid(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("mark-paid", flag.ExitOnError)
	caseID := fs.String("case-id", "", "Case ID")
	operator := fs.String("operator", "", "Operator")
	if err := parseFlags(args, fs); err != nil {
		return err
	}

	if *caseID == "" || *operator == "" {
		return fmt.Errorf("case-id and operator are required")
	}

	if err := client.MarkAsPaid(*caseID, *operator); err != nil {
		return err
	}

	fmt.Println("Case marked as paid successfully")
	return nil
}

func cmdFlagForReview(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("flag-review", flag.ExitOnError)
	caseID := fs.String("case-id", "", "Case ID")
	operator := fs.String("operator", "", "Operator")
	reason := fs.String("reason", "", "Reason for review")
	if err := parseFlags(args, fs); err != nil {
		return err
	}

	if *caseID == "" || *operator == "" || *reason == "" {
		return fmt.Errorf("all fields are required")
	}

	req := &common.FlagForReviewRequest{
		CaseID:   *caseID,
		Operator: *operator,
		Reason:   *reason,
	}

	if err := client.FlagForReview(req); err != nil {
		return err
	}

	fmt.Println("Case flagged for review successfully")
	return nil
}

func cmdResolveReview(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("resolve-review", flag.ExitOnError)
	caseID := fs.String("case-id", "", "Case ID")
	operator := fs.String("operator", "", "Operator")
	shouldReject := fs.Bool("reject", false, "Reject the case")
	reason := fs.String("reason", "", "Resolution reason")
	if err := parseFlags(args, fs); err != nil {
		return err
	}

	if *caseID == "" || *operator == "" || *reason == "" {
		return fmt.Errorf("case-id, operator, and reason are required")
	}

	req := &common.ResolveReviewRequest{
		CaseID:       *caseID,
		Operator:     *operator,
		ShouldReject: *shouldReject,
		Reason:       *reason,
	}

	if err := client.ResolveReview(req); err != nil {
		return err
	}

	if *shouldReject {
		fmt.Println("Case rejected successfully")
	} else {
		fmt.Println("Review resolved, case returned to normal workflow")
	}
	return nil
}

func init() {
	_ = strings.TrimSpace
}
