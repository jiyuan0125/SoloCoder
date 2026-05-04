package main

import (
	"billing/pkg/api"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"
	"time"
)

var serverURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	flagSet := flag.NewFlagSet(cmd, flag.ExitOnError)
	serverFlag := flagSet.String("server", "http://localhost:8080", "Server URL")
	flagSet.Parse(args)

	if serverFlag != nil {
		serverURL = *serverFlag
	}

	switch cmd {
	case "plans":
		handlePlans()
	case "create-plan":
		handleCreatePlan(args)
	case "customers":
		handleCustomers()
	case "create-customer":
		handleCreateCustomer(args)
	case "change-plan":
		handleChangePlan(args)
	case "bills":
		handleListAllBills()
	case "customer-bills":
		handleCustomerBills(args)
	case "generate-bill":
		handleGenerateBill(args)
	case "mark-paid":
		handleMarkPaid(args)
	case "pricing":
		handleGetPricing()
	case "update-pricing":
		handleUpdatePricing(args)
	case "usage":
		handleUsage(args)
	case "record-usage":
		handleRecordUsage(args)
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Billing Admin CLI - Usage:")
	fmt.Println()
	fmt.Println("  Plans:")
	fmt.Println("    admincli plans                          List all plans")
	fmt.Println("    admincli create-plan <name> <fee> <sms_quota> <storage_gb> [desc]")
	fmt.Println()
	fmt.Println("  Customers:")
	fmt.Println("    admincli customers                      List all customers")
	fmt.Println("    admincli create-customer <name> <email> <plan_id>")
	fmt.Println("    admincli change-plan <customer_id> <new_plan_id> [effective_date]")
	fmt.Println()
	fmt.Println("  Usage:")
	fmt.Println("    admincli usage <customer_id> [year] [month]")
	fmt.Println("    admincli record-usage <customer_id> <sms_incr> <storage_gb>")
	fmt.Println()
	fmt.Println("  Bills:")
	fmt.Println("    admincli bills                          List all bills")
	fmt.Println("    admincli customer-bills <customer_id>  List customer bills")
	fmt.Println("    admincli generate-bill <customer_id> <year> <month>")
	fmt.Println("    admincli mark-paid <bill_id> <operator> <amount> [remark]")
	fmt.Println()
	fmt.Println("  Pricing:")
	fmt.Println("    admincli pricing                        Get pricing config")
	fmt.Println("    admincli update-pricing [sms=price] [storage=price] [due=days] [overdue=days]")
	fmt.Println()
	fmt.Println("  Options:")
	fmt.Println("    -server=http://host:port                Specify server URL")
}

func handlePlans() {
	resp, err := http.Get(serverURL + "/api/plans")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	var result api.ListPlansResponse
	json.Unmarshal(body, &result)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tMonthly Fee\tSMS Quota\tStorage GB\tDescription")
	for _, p := range result.Plans {
		fmt.Fprintf(w, "%s\t%s\t%.2f\t%d\t%.1f\t%s\n",
			p.ID, p.Name, p.MonthlyFee, p.SmsQuota, p.StorageQuotaGB, p.Description)
	}
	w.Flush()
}

func handleCreatePlan(args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: admincli create-plan <name> <fee> <sms_quota> <storage_gb> [desc]")
		os.Exit(1)
	}

	name := args[0]
	fee, _ := strconv.ParseFloat(args[1], 64)
	smsQuota, _ := strconv.Atoi(args[2])
	storageGB, _ := strconv.ParseFloat(args[3], 64)
	desc := ""
	if len(args) > 4 {
		desc = args[4]
	}

	req := api.CreatePlanRequest{
		Name:           name,
		MonthlyFee:     fee,
		SmsQuota:       smsQuota,
		StorageQuotaGB: storageGB,
		Description:    desc,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/plans", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Error: %s\n", string(respBody))
		os.Exit(1)
	}

	var result api.CreatePlanResponse
	json.Unmarshal(respBody, &result)
	fmt.Printf("Created plan: %s (ID: %s)\n", result.Plan.Name, result.Plan.ID)
}

func handleCustomers() {
	resp, err := http.Get(serverURL + "/api/customers")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	var result api.ListCustomersResponse
	json.Unmarshal(body, &result)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tEmail\tPlan ID\tRegistered At\tActive")
	for _, c := range result.Customers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%v\n",
			c.ID, c.Name, c.Email, c.CurrentPlanID,
			c.RegisteredAt.Format("2006-01-02"), c.IsActive)
	}
	w.Flush()
}

func handleCreateCustomer(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: admincli create-customer <name> <email> <plan_id>")
		os.Exit(1)
	}

	req := api.CreateCustomerRequest{
		Name:   args[0],
		Email:  args[1],
		PlanID: args[2],
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/customers", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Error: %s\n", string(respBody))
		os.Exit(1)
	}

	var result api.CreateCustomerResponse
	json.Unmarshal(respBody, &result)
	fmt.Printf("Created customer: %s (ID: %s)\n", result.Customer.Name, result.Customer.ID)
}

func handleChangePlan(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: admincli change-plan <customer_id> <new_plan_id> [effective_date]")
		os.Exit(1)
	}

	var effectiveDate time.Time
	if len(args) >= 3 {
		t, err := time.Parse("2006-01-02", args[2])
		if err != nil {
			fmt.Printf("Invalid date format (use YYYY-MM-DD): %v\n", err)
			os.Exit(1)
		}
		effectiveDate = t
	} else {
		effectiveDate = time.Now()
	}

	req := api.ChangePlanRequest{
		CustomerID:    args[0],
		NewPlanID:     args[1],
		EffectiveDate: effectiveDate,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/plan-changes", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(respBody))
		os.Exit(1)
	}

	var result api.ChangePlanResponse
	json.Unmarshal(respBody, &result)
	fmt.Printf("Plan changed from %s to %s, effective %s\n",
		result.PlanChange.FromPlanID, result.PlanChange.ToPlanID,
		result.PlanChange.EffectiveDate.Format("2006-01-02"))
}

func handleUsage(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: admincli usage <customer_id> [year] [month]")
		os.Exit(1)
	}

	customerID := args[0]
	url := serverURL + "/api/usage/" + customerID
	if len(args) >= 2 {
		url += "?year=" + args[1]
		if len(args) >= 3 {
			url += "&month=" + args[2]
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	var result api.GetUsageResponse
	json.Unmarshal(body, &result)

	fmt.Printf("Usage for Customer %s, %d-%02d:\n", result.Usage.CustomerID, result.Usage.Year, result.Usage.Month)
	fmt.Printf("  SMS Count:    %d\n", result.Usage.SmsCount)
	fmt.Printf("  Storage Used: %.2f GB\n", result.Usage.StorageUsageGB)
	fmt.Printf("  Last Updated: %s\n", result.Usage.LastUpdated.Format("2006-01-02 15:04:05"))
}

func handleRecordUsage(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: admincli record-usage <customer_id> <sms_incr> <storage_gb>")
		os.Exit(1)
	}

	smsIncr, _ := strconv.Atoi(args[1])
	storageGB, _ := strconv.ParseFloat(args[2], 64)

	req := api.RecordUsageRequest{
		CustomerID:    args[0],
		SmsIncrement:  smsIncr,
		StorageUpdate: storageGB,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/usage", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(respBody))
		os.Exit(1)
	}

	var result api.RecordUsageResponse
	json.Unmarshal(respBody, &result)
	fmt.Printf("Usage recorded. SMS: %d, Storage: %.2f GB\n", result.Usage.SmsCount, result.Usage.StorageUsageGB)
}

func handleListAllBills() {
	resp, err := http.Get(serverURL + "/api/bills")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	var result api.ListAllBillsResponse
	json.Unmarshal(body, &result)

	printBills(result.Bills)
}

func handleCustomerBills(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: admincli customer-bills <customer_id>")
		os.Exit(1)
	}

	resp, err := http.Get(serverURL + "/api/customers/" + args[0] + "/bills")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	var result api.ListCustomerBillsResponse
	json.Unmarshal(body, &result)

	printBills(result.Bills)
}

func handleGenerateBill(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: admincli generate-bill <customer_id> <year> <month>")
		os.Exit(1)
	}

	year, _ := strconv.Atoi(args[1])
	month, _ := strconv.Atoi(args[2])

	req := api.GenerateBillRequest{
		CustomerID: args[0],
		Year:       year,
		Month:      month,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/bills/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Error: %s\n", string(respBody))
		os.Exit(1)
	}

	var result api.GenerateBillResponse
	json.Unmarshal(respBody, &result)
	printBillDetail(result.Bill)
}

func handleMarkPaid(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: admincli mark-paid <bill_id> <operator> <amount> [remark]")
		os.Exit(1)
	}

	amount, _ := strconv.ParseFloat(args[2], 64)
	remark := ""
	if len(args) >= 4 {
		remark = args[3]
	}

	req := api.MarkPaidRequest{
		BillID:   args[0],
		Operator: args[1],
		Amount:   amount,
		Remark:   remark,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/bills/mark-paid", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(respBody))
		os.Exit(1)
	}

	var result api.MarkPaidResponse
	json.Unmarshal(respBody, &result)
	fmt.Printf("Bill %s marked as paid by %s\n", result.Bill.ID, result.Bill.PaymentRecord.Operator)
}

func handleGetPricing() {
	resp, err := http.Get(serverURL + "/api/pricing")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	var result api.GetPricingResponse
	json.Unmarshal(body, &result)

	fmt.Println("Pricing Configuration:")
	fmt.Printf("  SMS Price per unit:    %.2f\n", result.Pricing.SmsPricePerUnit)
	fmt.Printf("  Storage Price per GB:  %.2f\n", result.Pricing.StoragePricePerGB)
	fmt.Printf("  Payment Due Days:      %d\n", result.Pricing.PaymentDueDays)
	fmt.Printf("  Serious Overdue Days:  %d\n", result.Pricing.SeriousOverdueDays)
}

func handleUpdatePricing(args []string) {
	req := &api.UpdatePricingRequest{}

	for _, arg := range args {
		if len(arg) < 4 {
			continue
		}
		eqIdx := -1
		for i, c := range arg {
			if c == '=' {
				eqIdx = i
				break
			}
		}
		if eqIdx < 0 {
			continue
		}

		key := arg[:eqIdx]
		value := arg[eqIdx+1:]

		switch key {
		case "sms":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				req.SmsPricePerUnit = &v
			}
		case "storage":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				req.StoragePricePerGB = &v
			}
		case "due":
			if v, err := strconv.Atoi(value); err == nil {
				req.PaymentDueDays = &v
			}
		case "overdue":
			if v, err := strconv.Atoi(value); err == nil {
				req.SeriousOverdueDays = &v
			}
		}
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest(http.MethodPut, serverURL+"/api/pricing", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(respBody))
		os.Exit(1)
	}

	var result api.GetPricingResponse
	json.Unmarshal(respBody, &result)
	fmt.Println("Pricing updated. Current config:")
	fmt.Printf("  SMS: %.2f, Storage: %.2f, Due: %d days, Overdue: %d days\n",
		result.Pricing.SmsPricePerUnit, result.Pricing.StoragePricePerGB,
		result.Pricing.PaymentDueDays, result.Pricing.SeriousOverdueDays)
}

func printBills(bills []*api.Bill) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tCustomer\tPeriod\tPlan Fee\tSMS Over\tStorage Over\tTotal\tStatus\tDue Date")
	for _, b := range bills {
		statusDisplay := b.Status
		if b.Status == api.BillStatusOverdue {
			statusDisplay = "OVERDUE"
		} else if b.Status == api.BillStatusSeriousOverdue {
			statusDisplay = "SERIOUS_OVERDUE"
		}

		fmt.Fprintf(w, "%s\t%s\t%d-%02d\t%.2f\t%.2f\t%.2f\t%.2f\t%s\t%s\n",
			b.ID, b.CustomerID, b.Year, b.Month,
			b.PlanFee, b.SmsOverageFee, b.StorageOverageFee,
			b.TotalAmount, statusDisplay, b.DueDate.Format("2006-01-02"))
	}
	w.Flush()
}

func printBillDetail(bill *api.Bill) {
	fmt.Println("========================================")
	fmt.Println("Bill Details")
	fmt.Println("========================================")
	fmt.Printf("Bill ID:      %s\n", bill.ID)
	fmt.Printf("Customer ID:  %s\n", bill.CustomerID)
	fmt.Printf("Period:       %d-%02d\n", bill.Year, bill.Month)
	fmt.Printf("Generated:    %s\n", bill.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println("--- Charges ---")
	fmt.Printf("Plan Fee:           %.2f\n", bill.PlanFee)
	fmt.Printf("SMS Overage Fee:    %.2f\n", bill.SmsOverageFee)
	fmt.Printf("Storage Overage Fee: %.2f\n", bill.StorageOverageFee)
	fmt.Println("------------------------")
	fmt.Printf("Total Amount:       %.2f\n", bill.TotalAmount)
	fmt.Println()
	fmt.Printf("Status:             %s\n", bill.Status)
	fmt.Printf("Due Date:           %s\n", bill.DueDate.Format("2006-01-02"))

	if bill.PaymentRecord != nil {
		fmt.Println()
		fmt.Println("--- Payment Record ---")
		fmt.Printf("Operator:           %s\n", bill.PaymentRecord.Operator)
		fmt.Printf("Marked At:          %s\n", bill.PaymentRecord.MarkedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Amount:             %.2f\n", bill.PaymentRecord.Amount)
		fmt.Printf("Remark:             %s\n", bill.PaymentRecord.Remark)
	}

	if len(bill.PlanBreakdown) > 1 {
		fmt.Println()
		fmt.Println("--- Plan Proration Breakdown ---")
		for _, p := range bill.PlanBreakdown {
			fmt.Printf("  Plan %s: %s to %s (%d days) - %.2f\n",
				p.PlanID, p.StartDate.Format("2006-01-02"),
				p.EndDate.Format("2006-01-02"), p.Days, p.Fee)
		}
	}
}
