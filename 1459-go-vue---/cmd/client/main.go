package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"contract-management/client"
	"contract-management/common"
)

const defaultBaseURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	baseURL := os.Getenv("SERVER_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	apiClient := client.NewAPIClient(baseURL)

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "create":
		handleCreate(apiClient, args)
	case "get":
		handleGet(apiClient, args)
	case "list":
		handleList(apiClient)
	case "update-amount":
		handleUpdateAmount(apiClient, args)
	case "complete-milestone":
		handleCompleteMilestone(apiClient, args)
	case "pay-milestone":
		handlePayMilestone(apiClient, args)
	case "progress":
		handleProgress(apiClient, args)
	case "audit-logs":
		handleAuditLogs(apiClient, args)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func handleCreate(apiClient *client.APIClient, args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	contractNo := fs.String("contract-no", "", "Contract number")
	partyA := fs.String("party-a", "", "Party A name")
	partyB := fs.String("party-b", "", "Party B name")
	totalAmount := fs.Int64("total-amount", 0, "Total amount in cents")
	signDateStr := fs.String("sign-date", "", "Sign date (YYYY-MM-DD)")
	contractType := fs.String("type", "", "Contract type")
	operator := fs.String("operator", "", "Operator name")
	milestonesFlag := fs.String("milestones", "", "Milestones as JSON array")
	weekdaysFlag := fs.String("weekdays", "", "Service weekdays (comma-separated, 0=Sunday, 1=Monday)")
	startTime := fs.String("start-time", "", "Service start time (HH:MM)")
	endTime := fs.String("end-time", "", "Service end time (HH:MM)")

	fs.Parse(args)

	if *contractNo == "" || *partyA == "" || *partyB == "" || *totalAmount <= 0 || *signDateStr == "" || *operator == "" || *milestonesFlag == "" {
		fmt.Println("Missing required fields for create")
		os.Exit(1)
	}

	signDate, err := time.Parse("2006-01-02", *signDateStr)
	if err != nil {
		fmt.Printf("Invalid sign date: %v\n", err)
		os.Exit(1)
	}

	var milestonesInput []struct {
		Name              string `json:"name"`
		PlannedDate       string `json:"planned_date"`
		PlannedPercentage int    `json:"planned_percentage"`
		Owner             string `json:"owner"`
	}
	if err := json.Unmarshal([]byte(*milestonesFlag), &milestonesInput); err != nil {
		fmt.Printf("Invalid milestones JSON: %v\n", err)
		os.Exit(1)
	}

	req := common.CreateContractRequest{
		ContractNo:   *contractNo,
		PartyA:       *partyA,
		PartyB:       *partyB,
		TotalAmount:  *totalAmount,
		SignDate:     signDate,
		ContractType: *contractType,
		Operator:     *operator,
	}

	for _, m := range milestonesInput {
		plannedDate, err := time.Parse("2006-01-02", m.PlannedDate)
		if err != nil {
			fmt.Printf("Invalid planned date for milestone %s: %v\n", m.Name, err)
			os.Exit(1)
		}
		req.Milestones = append(req.Milestones, common.MilestoneRequest{
			Name:              m.Name,
			PlannedDate:       plannedDate,
			PlannedPercentage: m.PlannedPercentage,
			Owner:             m.Owner,
		})
	}

	if *weekdaysFlag != "" && *startTime != "" && *endTime != "" {
		weekdayParts := strings.Split(*weekdaysFlag, ",")
		weekdays := []int{}
		for _, wd := range weekdayParts {
			w, err := strconv.Atoi(strings.TrimSpace(wd))
			if err != nil {
				fmt.Printf("Invalid weekday: %v\n", err)
				os.Exit(1)
			}
			weekdays = append(weekdays, w)
		}
		req.ServiceWindow = &common.ServiceWindow{
			StartTime: *startTime,
			EndTime:   *endTime,
			Weekdays:  weekdays,
		}
	}

	contract, err := apiClient.CreateContract(req)
	if err != nil {
		fmt.Printf("Error creating contract: %v\n", err)
		os.Exit(1)
	}

	printJSON(contract)
}

func handleGet(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: get <contract-id>")
		os.Exit(1)
	}

	contract, err := apiClient.GetContract(args[0])
	if err != nil {
		fmt.Printf("Error getting contract: %v\n", err)
		os.Exit(1)
	}

	printJSON(contract)
}

func handleList(apiClient *client.APIClient) {
	list, err := apiClient.ListContracts()
	if err != nil {
		fmt.Printf("Error listing contracts: %v\n", err)
		os.Exit(1)
	}

	printJSON(list)
}

func handleUpdateAmount(apiClient *client.APIClient, args []string) {
	fs := flag.NewFlagSet("update-amount", flag.ExitOnError)
	newAmount := fs.Int64("amount", 0, "New total amount in cents")
	operator := fs.String("operator", "", "Operator name")

	fs.Parse(args)

	if fs.NArg() < 1 || *newAmount <= 0 || *operator == "" {
		fmt.Println("Usage: update-amount <contract-id> -amount=<new-amount> -operator=<operator>")
		os.Exit(1)
	}

	contractID := fs.Arg(0)
	contract, err := apiClient.UpdateContractAmount(contractID, common.UpdateContractAmountRequest{
		NewTotalAmount: *newAmount,
		Operator:       *operator,
	})
	if err != nil {
		fmt.Printf("Error updating contract amount: %v\n", err)
		os.Exit(1)
	}

	printJSON(contract)
}

func handleCompleteMilestone(apiClient *client.APIClient, args []string) {
	fs := flag.NewFlagSet("complete-milestone", flag.ExitOnError)
	actualDateStr := fs.String("actual-date", "", "Actual completion date (RFC3339)")
	approved := fs.Bool("approved", false, "Whether milestone is approved")
	operator := fs.String("operator", "", "Operator name")

	fs.Parse(args)

	if fs.NArg() < 2 || *actualDateStr == "" || *operator == "" {
		fmt.Println("Usage: complete-milestone <contract-id> <milestone-id> -actual-date=<date> -operator=<operator> [-approved]")
		os.Exit(1)
	}

	contractID := fs.Arg(0)
	milestoneID := fs.Arg(1)

	actualDate, err := time.Parse(time.RFC3339, *actualDateStr)
	if err != nil {
		fmt.Printf("Invalid actual date: %v\n", err)
		os.Exit(1)
	}

	milestone, err := apiClient.CompleteMilestone(contractID, milestoneID, common.CompleteMilestoneRequest{
		ActualDate: actualDate,
		Approved:   *approved,
		Operator:   *operator,
	})
	if err != nil {
		fmt.Printf("Error completing milestone: %v\n", err)
		os.Exit(1)
	}

	printJSON(milestone)
}

func handlePayMilestone(apiClient *client.APIClient, args []string) {
	fs := flag.NewFlagSet("pay-milestone", flag.ExitOnError)
	paidAmount := fs.Int64("amount", 0, "Paid amount in cents")
	operator := fs.String("operator", "", "Operator name")

	fs.Parse(args)

	if fs.NArg() < 2 || *paidAmount <= 0 || *operator == "" {
		fmt.Println("Usage: pay-milestone <contract-id> <milestone-id> -amount=<paid-amount> -operator=<operator>")
		os.Exit(1)
	}

	contractID := fs.Arg(0)
	milestoneID := fs.Arg(1)

	payment, err := apiClient.PayMilestone(contractID, milestoneID, common.PayMilestoneRequest{
		PaidAmount: *paidAmount,
		Operator:   *operator,
	})
	if err != nil {
		fmt.Printf("Error paying milestone: %v\n", err)
		os.Exit(1)
	}

	printJSON(payment)
}

func handleProgress(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: progress <contract-id>")
		os.Exit(1)
	}

	progress, err := apiClient.GetProgress(args[0])
	if err != nil {
		fmt.Printf("Error getting progress: %v\n", err)
		os.Exit(1)
	}

	printJSON(progress)
}

func handleAuditLogs(apiClient *client.APIClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: audit-logs <contract-id>")
		os.Exit(1)
	}

	logs, err := apiClient.GetAuditLogs(args[0])
	if err != nil {
		fmt.Printf("Error getting audit logs: %v\n", err)
		os.Exit(1)
	}

	printJSON(logs)
}

func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func printUsage() {
	fmt.Println("Contract Management Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  contract-client <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create              Create a new contract")
	fmt.Println("  get <id>            Get contract by ID")
	fmt.Println("  list                List all contracts")
	fmt.Println("  update-amount       Update contract total amount")
	fmt.Println("  complete-milestone  Complete a milestone")
	fmt.Println("  pay-milestone       Pay for a milestone")
	fmt.Println("  progress            Get contract progress")
	fmt.Println("  audit-logs          Get contract audit logs")
	fmt.Println("  help                Show this help")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  SERVER_URL          Server base URL (default: http://localhost:8080)")
}
