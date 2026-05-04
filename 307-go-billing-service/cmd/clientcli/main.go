package main

import (
	"billing/pkg/api"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
)

var serverURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	customerID := os.Args[1]
	cmd := os.Args[2]
	args := os.Args[3:]

	flagSet := flag.NewFlagSet(cmd, flag.ExitOnError)
	serverFlag := flagSet.String("server", "http://localhost:8080", "Server URL")
	flagSet.Parse(args)

	if serverFlag != nil {
		serverURL = *serverFlag
	}

	switch cmd {
	case "info":
		handleCustomerInfo(customerID)
	case "bills":
		handleCustomerBills(customerID)
	case "bill-detail":
		handleBillDetail(args)
	case "usage":
		handleUsage(customerID, args)
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Billing Client CLI - Usage:")
	fmt.Println()
	fmt.Println("  clientcli <customer_id> info               View customer info")
	fmt.Println("  clientcli <customer_id> bills              List all your bills")
	fmt.Println("  clientcli <customer_id> bill-detail <bill_id>  View bill detail")
	fmt.Println("  clientcli <customer_id> usage [year] [month]   View current usage")
	fmt.Println()
	fmt.Println("  Options:")
	fmt.Println("    -server=http://host:port                  Specify server URL")
}

func handleCustomerInfo(customerID string) {
	resp, err := http.Get(serverURL + "/api/customers/" + customerID)
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

	var result api.GetCustomerResponse
	json.Unmarshal(body, &result)

	fmt.Println("========================================")
	fmt.Println("Customer Information")
	fmt.Println("========================================")
	fmt.Printf("ID:       %s\n", result.Customer.ID)
	fmt.Printf("Name:     %s\n", result.Customer.Name)
	fmt.Printf("Email:    %s\n", result.Customer.Email)
	fmt.Printf("Plan ID:  %s\n", result.Customer.CurrentPlanID)
	fmt.Printf("Registered: %s\n", result.Customer.RegisteredAt.Format("2006-01-02"))
	fmt.Printf("Active:   %v\n", result.Customer.IsActive)
}

func handleCustomerBills(customerID string) {
	resp, err := http.Get(serverURL + "/api/customers/" + customerID + "/bills")
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

	if len(result.Bills) == 0 {
		fmt.Println("No bills found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Period\tTotal\tStatus\tDue Date")
	for _, b := range result.Bills {
		statusDisplay := b.Status
		if b.Status == api.BillStatusOverdue {
			statusDisplay = "OVERDUE"
		} else if b.Status == api.BillStatusSeriousOverdue {
			statusDisplay = "SERIOUS_OVERDUE"
		}

		fmt.Fprintf(w, "%d-%02d\t%.2f\t%s\t%s\n",
			b.Year, b.Month, b.TotalAmount, statusDisplay, b.DueDate.Format("2006-01-02"))
	}
	w.Flush()
}

func handleBillDetail(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: clientcli <customer_id> bill-detail <bill_id>")
		os.Exit(1)
	}

	billID := args[0]
	resp, err := http.Get(serverURL + "/api/bills/" + billID)
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

	var result api.GetBillResponse
	json.Unmarshal(body, &result)

	printBillDetail(result.Bill)
}

func handleUsage(customerID string, args []string) {
	url := serverURL + "/api/usage/" + customerID
	if len(args) >= 1 {
		url += "?year=" + args[0]
		if len(args) >= 2 {
			url += "&month=" + args[1]
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

	fmt.Println("========================================")
	fmt.Println("Usage Details")
	fmt.Println("========================================")
	fmt.Printf("Period:    %d-%02d\n", result.Usage.Year, result.Usage.Month)
	fmt.Printf("SMS Used:  %d\n", result.Usage.SmsCount)
	fmt.Printf("Storage:   %.2f GB\n", result.Usage.StorageUsageGB)
	fmt.Printf("Updated:   %s\n", result.Usage.LastUpdated.Format("2006-01-02 15:04:05"))
}

func printBillDetail(bill *api.Bill) {
	fmt.Println("========================================")
	fmt.Println("Bill Details")
	fmt.Println("========================================")
	fmt.Printf("Period:       %d-%02d\n", bill.Year, bill.Month)
	fmt.Printf("Generated:    %s\n", bill.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println("--- Breakdown ---")
	fmt.Printf("Plan Fee:           %.2f\n", bill.PlanFee)
	fmt.Printf("SMS Overage:        %.2f\n", bill.SmsOverageFee)
	fmt.Printf("Storage Overage:    %.2f\n", bill.StorageOverageFee)
	fmt.Println("------------------------")
	fmt.Printf("Total Amount:       %.2f\n", bill.TotalAmount)
	fmt.Println()
	fmt.Printf("Status:             %s\n", bill.Status)
	fmt.Printf("Due Date:           %s\n", bill.DueDate.Format("2006-01-02"))

	if bill.Status == api.BillStatusOverdue {
		fmt.Println()
		fmt.Println("*** NOTICE: This bill is overdue. Please contact support. ***")
	} else if bill.Status == api.BillStatusSeriousOverdue {
		fmt.Println()
		fmt.Println("*** WARNING: This bill is SERIOUSLY OVERDUE (over 60 days). ***")
		fmt.Println("*** Services may be suspended. Please make payment immediately. ***")
	}

	if bill.PaymentRecord != nil {
		fmt.Println()
		fmt.Println("--- Payment ---")
		fmt.Printf("Paid At:    %s\n", bill.PaymentRecord.MarkedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Amount:     %.2f\n", bill.PaymentRecord.Amount)
		if bill.PaymentRecord.Remark != "" {
			fmt.Printf("Remark:     %s\n", bill.PaymentRecord.Remark)
		}
	}

	if len(bill.PlanBreakdown) > 1 {
		fmt.Println()
		fmt.Println("--- Plan Changes This Month ---")
		for _, p := range bill.PlanBreakdown {
			fmt.Printf("  %s: %s - %s (%d days) - %.2f\n",
				p.PlanID,
				p.StartDate.Format("2006-01-02"),
				p.EndDate.Format("2006-01-02"),
				p.Days, p.Fee)
		}
	}
}
