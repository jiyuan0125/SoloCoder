package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"supplier-portal/common"
)

func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

func usage() {
	fmt.Println(`Supplier Portal Client

Usage:
  supplier-client [global-options] <command> [command-options] [arguments]

Global Options:
  --server <url>    Server URL (default: http://localhost:8080)
  -h, --help        Show this help

Commands:
  Buyer Commands (for Procurement Specialists):
    create-inquiry    Create a new inquiry (RFQ)
    list-inquiries    List all inquiries
    get-inquiry       Get inquiry details
    update-materials  Update inquiry materials (invalidates existing quotes)
    comparison        Generate price comparison
    award             Award suppliers and create purchase orders
    get-orders        Get orders by inquiry

  Supplier Commands:
    list-todos        List pending quotation reminders
    submit-quote      Submit or update a quotation
    get-quotes        View all quotations for an inquiry
    my-orders         Get orders for a supplier

Examples:
  supplier-client create-inquiry --buyer B001 --title "Office Supplies" \
    --suppliers S001,S002 --deadline "2026-05-15T00:00:00Z" \
    --materials '[{"name":"笔","spec":"0.5mm","quantity":100,"unit":"支"},{"name":"A4纸","spec":"70g","quantity":50,"unit":"包"}]'

  supplier-client submit-quote --supplier S001 --inquiry INQ-123 \
    --items '[{"material_id":"MAT-1","unit_price":1.5,"delivery_days":3},{"material_id":"MAT-2","unit_price":25.0,"delivery_days":5}]' \
    --remark "批量采购可提供额外折扣"

  supplier-client award --inquiry INQ-123 --awards '[{"material_id":"MAT-1","supplier_id":"S001"},{"material_id":"MAT-2","supplier_id":"S002"}]'
`)
	os.Exit(1)
}

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
	}

	client := NewAPIClient(*serverURL)
	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	var err error
	switch cmd {
	case "create-inquiry":
		err = cmdCreateInquiry(client, args)
	case "list-inquiries":
		err = cmdListInquiries(client, args)
	case "get-inquiry":
		err = cmdGetInquiry(client, args)
	case "update-materials":
		err = cmdUpdateMaterials(client, args)
	case "list-todos":
		err = cmdListTodos(client, args)
	case "submit-quote":
		err = cmdSubmitQuote(client, args)
	case "get-quotes":
		err = cmdGetQuotes(client, args)
	case "comparison":
		err = cmdComparison(client, args)
	case "award":
		err = cmdAward(client, args)
	case "get-orders":
		err = cmdGetOrdersByInquiry(client, args)
	case "my-orders":
		err = cmdMyOrders(client, args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		usage()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseMaterialsJSON(jsonStr string) ([]common.MaterialDTO, error) {
	var mats []common.MaterialDTO
	if err := json.Unmarshal([]byte(jsonStr), &mats); err != nil {
		return nil, fmt.Errorf("invalid materials JSON: %v", err)
	}
	return mats, nil
}

func parseItemsJSON(jsonStr string) ([]common.QuotationItemDTO, error) {
	var items []common.QuotationItemDTO
	if err := json.Unmarshal([]byte(jsonStr), &items); err != nil {
		return nil, fmt.Errorf("invalid items JSON: %v", err)
	}
	return items, nil
}

func parseAwardsJSON(jsonStr string) ([]common.AwardedMaterialDTO, error) {
	var awards []common.AwardedMaterialDTO
	if err := json.Unmarshal([]byte(jsonStr), &awards); err != nil {
		return nil, fmt.Errorf("invalid awards JSON: %v", err)
	}
	return awards, nil
}

func cmdCreateInquiry(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("create-inquiry", flag.ExitOnError)
	buyerID := fs.String("buyer", "", "Buyer ID")
	title := fs.String("title", "", "Inquiry title")
	suppliersStr := fs.String("suppliers", "", "Comma-separated supplier IDs")
	deadline := fs.String("deadline", "", "Quote deadline (RFC3339)")
	materialsJSON := fs.String("materials", "", "Materials JSON array")
	fs.Parse(args)

	if *buyerID == "" || *title == "" || *suppliersStr == "" || *deadline == "" || *materialsJSON == "" {
		return fmt.Errorf("missing required parameters")
	}

	supplierIDs := strings.Split(*suppliersStr, ",")
	for i := range supplierIDs {
		supplierIDs[i] = strings.TrimSpace(supplierIDs[i])
	}

	materials, err := parseMaterialsJSON(*materialsJSON)
	if err != nil {
		return err
	}

	req := common.CreateInquiryRequest{
		Title:       *title,
		BuyerID:     *buyerID,
		Materials:   materials,
		SupplierIDs: supplierIDs,
		Deadline:    *deadline,
	}

	result, err := client.CreateInquiry(req)
	if err != nil {
		return err
	}

	fmt.Println("Inquiry created successfully:")
	printJSON(result)
	return nil
}

func cmdListInquiries(client *APIClient, args []string) error {
	result, err := client.ListInquiries()
	if err != nil {
		return err
	}

	fmt.Printf("Total inquiries: %d\n\n", len(result))
	printJSON(result)
	return nil
}

func cmdGetInquiry(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("get-inquiry", flag.ExitOnError)
	id := fs.String("id", "", "Inquiry ID")
	fs.Parse(args)

	if *id == "" {
		return fmt.Errorf("inquiry ID required")
	}

	result, err := client.GetInquiry(*id)
	if err != nil {
		return err
	}

	printJSON(result)
	return nil
}

func cmdUpdateMaterials(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("update-materials", flag.ExitOnError)
	inquiryID := fs.String("inquiry", "", "Inquiry ID")
	materialsJSON := fs.String("materials", "", "Materials JSON array")
	fs.Parse(args)

	if *inquiryID == "" || *materialsJSON == "" {
		return fmt.Errorf("missing required parameters")
	}

	materials, err := parseMaterialsJSON(*materialsJSON)
	if err != nil {
		return err
	}

	req := common.UpdateInquiryMaterialsRequest{
		InquiryID: *inquiryID,
		Materials: materials,
	}

	result, err := client.UpdateInquiryMaterials(req)
	if err != nil {
		return err
	}

	fmt.Println("Materials updated (existing quotes invalidated):")
	printJSON(result)
	return nil
}

func cmdListTodos(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("list-todos", flag.ExitOnError)
	supplierID := fs.String("supplier", "", "Supplier ID")
	fs.Parse(args)

	if *supplierID == "" {
		return fmt.Errorf("supplier ID required")
	}

	result, err := client.GetTodoReminders(*supplierID)
	if err != nil {
		return err
	}

	fmt.Printf("Pending todos for %s: %d\n\n", *supplierID, len(result))
	printJSON(result)
	return nil
}

func cmdSubmitQuote(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("submit-quote", flag.ExitOnError)
	supplierID := fs.String("supplier", "", "Supplier ID")
	inquiryID := fs.String("inquiry", "", "Inquiry ID")
	itemsJSON := fs.String("items", "", "Quote items JSON array")
	remark := fs.String("remark", "", "Additional remarks")
	fs.Parse(args)

	if *supplierID == "" || *inquiryID == "" || *itemsJSON == "" {
		return fmt.Errorf("missing required parameters")
	}

	items, err := parseItemsJSON(*itemsJSON)
	if err != nil {
		return err
	}

	req := common.SubmitQuotationRequest{
		InquiryID:  *inquiryID,
		SupplierID: *supplierID,
		Items:      items,
		Remark:     *remark,
	}

	result, err := client.SubmitQuotation(req)
	if err != nil {
		return err
	}

	fmt.Println("Quotation submitted:")
	printJSON(result)
	return nil
}

func cmdGetQuotes(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("get-quotes", flag.ExitOnError)
	inquiryID := fs.String("inquiry", "", "Inquiry ID")
	fs.Parse(args)

	if *inquiryID == "" {
		return fmt.Errorf("inquiry ID required")
	}

	result, err := client.GetQuotations(*inquiryID)
	if err != nil {
		return err
	}

	fmt.Printf("Quotations for %s: %d\n\n", *inquiryID, len(result))
	printJSON(result)
	return nil
}

func cmdComparison(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("comparison", flag.ExitOnError)
	inquiryID := fs.String("inquiry", "", "Inquiry ID")
	fs.Parse(args)

	if *inquiryID == "" {
		return fmt.Errorf("inquiry ID required")
	}

	result, err := client.GenerateComparison(*inquiryID)
	if err != nil {
		return err
	}

	fmt.Println("Price Comparison Result:")
	printJSON(result)
	return nil
}

func cmdAward(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("award", flag.ExitOnError)
	inquiryID := fs.String("inquiry", "", "Inquiry ID")
	awardsJSON := fs.String("awards", "", "Awards JSON array")
	fs.Parse(args)

	if *inquiryID == "" || *awardsJSON == "" {
		return fmt.Errorf("missing required parameters")
	}

	awards, err := parseAwardsJSON(*awardsJSON)
	if err != nil {
		return err
	}

	req := common.AwardRequest{
		InquiryID: *inquiryID,
		Awards:    awards,
	}

	result, err := client.AwardAndCreateOrders(req)
	if err != nil {
		return err
	}

	fmt.Printf("Purchase orders created: %d\n\n", len(result))
	printJSON(result)
	return nil
}

func cmdGetOrdersByInquiry(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("get-orders", flag.ExitOnError)
	inquiryID := fs.String("inquiry", "", "Inquiry ID")
	fs.Parse(args)

	if *inquiryID == "" {
		return fmt.Errorf("inquiry ID required")
	}

	result, err := client.GetOrdersByInquiry(*inquiryID)
	if err != nil {
		return err
	}

	fmt.Printf("Orders for inquiry %s: %d\n\n", *inquiryID, len(result))
	printJSON(result)
	return nil
}

func cmdMyOrders(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("my-orders", flag.ExitOnError)
	supplierID := fs.String("supplier", "", "Supplier ID")
	fs.Parse(args)

	if *supplierID == "" {
		return fmt.Errorf("supplier ID required")
	}

	result, err := client.GetOrdersBySupplier(*supplierID)
	if err != nil {
		return err
	}

	fmt.Printf("Orders for supplier %s: %d\n\n", *supplierID, len(result))
	printJSON(result)
	return nil
}
