package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"property-management/pkg/api"
	"property-management/pkg/core/models"
	"strings"
	"time"
)

const defaultServerURL = "http://localhost:8080"

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) do(method, path string, body interface{}) (*api.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp api.Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return &apiResp, nil
}

func main() {
	serverURL := flag.String("server", defaultServerURL, "Server URL")
	flag.Parse()

	if flag.NArg() < 1 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(*serverURL)
	command := flag.Arg(0)
	args := flag.Args()[1:]

	var err error
	switch command {
	case "create-user":
		err = cmdCreateUser(client, args)
	case "list-users":
		err = cmdListUsers(client)
	case "get-user":
		err = cmdGetUser(client, args)
	case "create-repair":
		err = cmdCreateRepair(client, args)
	case "assign-repair":
		err = cmdAssignRepair(client, args)
	case "start-repair":
		err = cmdStartRepair(client, args)
	case "complete-repair":
		err = cmdCompleteRepair(client, args)
	case "confirm-repair":
		err = cmdConfirmRepair(client, args)
	case "get-repair":
		err = cmdGetRepair(client, args)
	case "list-repairs":
		err = cmdListRepairs(client, args)
	case "create-bill":
		err = cmdCreateBill(client, args)
	case "generate-bills":
		err = cmdGenerateBills(client, args)
	case "pay-bill":
		err = cmdPayBill(client, args)
	case "get-bill":
		err = cmdGetBill(client, args)
	case "list-bills":
		err = cmdListBills(client, args)
	case "get-penalty":
		err = cmdGetPenalty(client, args)
	case "create-announcement":
		err = cmdCreateAnnouncement(client, args)
	case "update-announcement":
		err = cmdUpdateAnnouncement(client, args)
	case "delete-announcement":
		err = cmdDeleteAnnouncement(client, args)
	case "get-announcement":
		err = cmdGetAnnouncement(client, args)
	case "list-announcements":
		err = cmdListAnnouncements(client, args)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Property Management System CLI")
	fmt.Println()
	fmt.Println("Usage: pmc [global-options] <command> [arguments]")
	fmt.Println()
	fmt.Println("Global Options:")
	fmt.Println("  -server string   Server URL (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  Users:")
	fmt.Println("    create-user <username> <name> <role> [building] [unit] [phone]")
	fmt.Println("    list-users")
	fmt.Println("    get-user <id>")
	fmt.Println()
	fmt.Println("  Repairs:")
	fmt.Println("    create-repair <owner-id> <type> <description> <expected-time>")
	fmt.Println("      Types: water_electric, door_window, public_facility, other")
	fmt.Println("    assign-repair <repair-id> <supervisor-id> <staff-id>")
	fmt.Println("    start-repair <repair-id> <staff-id>")
	fmt.Println("    complete-repair <repair-id> <staff-id> <result> [image-urls...]")
	fmt.Println("    confirm-repair <repair-id> <owner-id>")
	fmt.Println("    get-repair <id>")
	fmt.Println("    list-repairs [owner-id] [status]")
	fmt.Println()
	fmt.Println("  Bills:")
	fmt.Println("    create-bill <user-id> <type> <amount> <due-date> <month> <year>")
	fmt.Println("      Types: property, water, electric, parking, repair")
	fmt.Println("      Amount in cents (e.g., 200000 = 2000.00 yuan)")
	fmt.Println("    generate-bills <type> <amount> <due-date> <month> <year>")
	fmt.Println("    pay-bill <bill-id> <user-id> <amount> <method>")
	fmt.Println("    get-bill <id>")
	fmt.Println("    list-bills [user-id] [status]")
	fmt.Println("      Status: pending, partial, paid")
	fmt.Println("    get-penalty <bill-id>")
	fmt.Println()
	fmt.Println("  Announcements:")
	fmt.Println("    create-announcement <title> <content> <scope> <effective> <expiry> <created-by> [building]")
	fmt.Println("      Scope: all, building")
	fmt.Println("    update-announcement <id> <title> <content> <scope> <effective> <expiry> [building]")
	fmt.Println("    delete-announcement <id>")
	fmt.Println("    get-announcement <id>")
	fmt.Println("    list-announcements [building] [valid-only]")
}

func printResponse(resp *api.Response) {
	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(data))
}

func cmdCreateUser(client *Client, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: create-user <username> <name> <role> [building] [unit] [phone]")
	}
	role := models.Role(args[2])
	building := ""
	unit := ""
	phone := ""
	if len(args) > 3 {
		building = args[3]
	}
	if len(args) > 4 {
		unit = args[4]
	}
	if len(args) > 5 {
		phone = args[5]
	}

	req := map[string]interface{}{
		"username": args[0],
		"name":     args[1],
		"role":     role,
		"building": building,
		"unit":     unit,
		"phone":    phone,
	}
	resp, err := client.do(http.MethodPost, "/users", req)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdListUsers(client *Client) error {
	resp, err := client.do(http.MethodGet, "/users", nil)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdGetUser(client *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get-user <id>")
	}
	resp, err := client.do(http.MethodGet, "/users/"+args[0], nil)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdCreateRepair(client *Client, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: create-repair <owner-id> <type> <description> <expected-time>")
	}
	expectedTime, err := time.Parse(time.RFC3339, args[3])
	if err != nil {
		return fmt.Errorf("invalid expected time format, use RFC3339 (e.g., 2024-01-15T10:00:00Z): %v", err)
	}
	req := api.CreateRepairRequest{
		OwnerID:      args[0],
		RepairType:   models.RepairType(args[1]),
		Description:  args[2],
		ExpectedTime: expectedTime,
	}
	resp, err := client.do(http.MethodPost, "/repairs", req)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdAssignRepair(client *Client, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: assign-repair <repair-id> <supervisor-id> <staff-id>")
	}
	req := api.AssignRepairRequest{
		SupervisorID: args[1],
		StaffID:      args[2],
	}
	resp, err := client.do(http.MethodPost, "/repairs/"+args[0]+"/assign", req)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdStartRepair(client *Client, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: start-repair <repair-id> <staff-id>")
	}
	resp, err := client.do(http.MethodPost, "/repairs/"+args[0]+"/start?staff_id="+args[1], nil)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdCompleteRepair(client *Client, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: complete-repair <repair-id> <staff-id> <result> [image-urls...]")
	}
	imageURLs := []string{}
	if len(args) > 3 {
		imageURLs = args[3:]
	}
	req := api.CompleteRepairRequest{
		StaffID:        args[1],
		Result:         args[2],
		ResultImageURLs: imageURLs,
	}
	resp, err := client.do(http.MethodPost, "/repairs/"+args[0]+"/complete", req)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdConfirmRepair(client *Client, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: confirm-repair <repair-id> <owner-id>")
	}
	req := api.ConfirmRepairRequest{
		OwnerID: args[1],
	}
	resp, err := client.do(http.MethodPost, "/repairs/"+args[0]+"/confirm", req)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdGetRepair(client *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get-repair <id>")
	}
	resp, err := client.do(http.MethodGet, "/repairs/"+args[0], nil)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdListRepairs(client *Client, args []string) error {
	path := "/repairs"
	if len(args) > 0 {
		path += "?owner_id=" + args[0]
	}
	if len(args) > 1 {
		if strings.Contains(path, "?") {
			path += "&status=" + args[1]
		} else {
			path += "?status=" + args[1]
		}
	}
	resp, err := client.do(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdCreateBill(client *Client, args []string) error {
	if len(args) < 6 {
		return fmt.Errorf("usage: create-bill <user-id> <type> <amount> <due-date> <month> <year>")
	}
	amount, err := parseInt64(args[2])
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}
	dueDate, err := time.Parse(time.RFC3339, args[3])
	if err != nil {
		return fmt.Errorf("invalid due date format: %v", err)
	}
	month, err := parseInt(args[4])
	if err != nil {
		return fmt.Errorf("invalid month: %v", err)
	}
	year, err := parseInt(args[5])
	if err != nil {
		return fmt.Errorf("invalid year: %v", err)
	}
	req := api.CreateBillRequest{
		UserID:   args[0],
		BillType: models.BillType(args[1]),
		Amount:   amount,
		DueDate:  dueDate,
		Month:    month,
		Year:     year,
	}
	resp, err := client.do(http.MethodPost, "/bills", req)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdGenerateBills(client *Client, args []string) error {
	if len(args) < 5 {
		return fmt.Errorf("usage: generate-bills <type> <amount> <due-date> <month> <year>")
	}
	amount, err := parseInt64(args[1])
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}
	dueDate, err := time.Parse(time.RFC3339, args[2])
	if err != nil {
		return fmt.Errorf("invalid due date format: %v", err)
	}
	month, err := parseInt(args[3])
	if err != nil {
		return fmt.Errorf("invalid month: %v", err)
	}
	year, err := parseInt(args[4])
	if err != nil {
		return fmt.Errorf("invalid year: %v", err)
	}
	req := api.GenerateMonthlyBillsRequest{
		BillType: models.BillType(args[0]),
		Amount:   amount,
		DueDate:  dueDate,
		Month:    month,
		Year:     year,
	}
	resp, err := client.do(http.MethodPost, "/bills/generate", req)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdPayBill(client *Client, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: pay-bill <bill-id> <user-id> <amount> <method>")
	}
	amount, err := parseInt64(args[2])
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}
	req := api.PayBillRequest{
		UserID:        args[1],
		Amount:        amount,
		PaymentMethod: args[3],
	}
	resp, err := client.do(http.MethodPost, "/bills/"+args[0]+"/pay", req)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdGetBill(client *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get-bill <id>")
	}
	resp, err := client.do(http.MethodGet, "/bills/"+args[0], nil)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdListBills(client *Client, args []string) error {
	path := "/bills"
	if len(args) > 0 {
		path += "?user_id=" + args[0]
	}
	if len(args) > 1 {
		if strings.Contains(path, "?") {
			path += "&status=" + args[1]
		} else {
			path += "?status=" + args[1]
		}
	}
	resp, err := client.do(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdGetPenalty(client *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get-penalty <bill-id>")
	}
	resp, err := client.do(http.MethodGet, "/bills/"+args[0]+"/penalty", nil)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdCreateAnnouncement(client *Client, args []string) error {
	if len(args) < 6 {
		return fmt.Errorf("usage: create-announcement <title> <content> <scope> <effective> <expiry> <created-by> [building]")
	}
	effective, err := time.Parse(time.RFC3339, args[3])
	if err != nil {
		return fmt.Errorf("invalid effective time format: %v", err)
	}
	expiry, err := time.Parse(time.RFC3339, args[4])
	if err != nil {
		return fmt.Errorf("invalid expiry time format: %v", err)
	}
	building := ""
	if len(args) > 6 {
		building = args[6]
	}
	req := api.CreateAnnouncementRequest{
		Title:          args[0],
		Content:        args[1],
		Scope:          models.AnnouncementScope(args[2]),
		TargetBuilding: building,
		EffectiveTime:  effective,
		ExpiryTime:     expiry,
		CreatedBy:      args[5],
	}
	resp, err := client.do(http.MethodPost, "/announcements", req)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdUpdateAnnouncement(client *Client, args []string) error {
	if len(args) < 6 {
		return fmt.Errorf("usage: update-announcement <id> <title> <content> <scope> <effective> <expiry> [building]")
	}
	effective, err := time.Parse(time.RFC3339, args[4])
	if err != nil {
		return fmt.Errorf("invalid effective time format: %v", err)
	}
	expiry, err := time.Parse(time.RFC3339, args[5])
	if err != nil {
		return fmt.Errorf("invalid expiry time format: %v", err)
	}
	building := ""
	if len(args) > 6 {
		building = args[6]
	}
	req := api.UpdateAnnouncementRequest{
		Title:          args[1],
		Content:        args[2],
		Scope:          models.AnnouncementScope(args[3]),
		TargetBuilding: building,
		EffectiveTime:  effective,
		ExpiryTime:     expiry,
	}
	resp, err := client.do(http.MethodPut, "/announcements/"+args[0], req)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdDeleteAnnouncement(client *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: delete-announcement <id>")
	}
	resp, err := client.do(http.MethodDelete, "/announcements/"+args[0], nil)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdGetAnnouncement(client *Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get-announcement <id>")
	}
	resp, err := client.do(http.MethodGet, "/announcements/"+args[0], nil)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func cmdListAnnouncements(client *Client, args []string) error {
	path := "/announcements"
	if len(args) > 0 {
		path += "?building=" + args[0]
	}
	if len(args) > 1 {
		if strings.Contains(path, "?") {
			path += "&valid_only=" + args[1]
		} else {
			path += "?valid_only=" + args[1]
		}
	}
	resp, err := client.do(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func parseInt64(s string) (int64, error) {
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
