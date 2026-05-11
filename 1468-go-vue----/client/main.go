package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"tender-management/common"
)

type APIClient struct {
	BaseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{BaseURL: baseURL}
}

func (c *APIClient) request(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp common.Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("unexpected response: %s", string(respBody))
	}

	if !apiResp.Success {
		return fmt.Errorf("error: %s", apiResp.Message)
	}

	if result != nil && apiResp.Data != nil {
		dataJSON, err := json.Marshal(apiResp.Data)
		if err != nil {
			return err
		}
		return json.Unmarshal(dataJSON, result)
	}

	return nil
}

func printResponse(msg string, data interface{}) {
	fmt.Println("✓", msg)
	if data != nil {
		jsonBytes, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(jsonBytes))
	}
}

func cmdProjectCreate(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("project-create", flag.ExitOnError)
	name := fs.String("name", "", "Project name")
	desc := fs.String("desc", "", "Description")
	method := fs.String("method", "public", "Tender method: public|invite")
	bidDeadlineStr := fs.String("bid-deadline", "", "Bid deadline (RFC3339)")
	openTimeStr := fs.String("open-time", "", "Open time (RFC3339)")
	evalMethod := fs.String("eval", "lowest_price", "Evaluation method: lowest_price|comprehensive_score")
	invited := fs.String("invited", "", "Invited suppliers (comma separated)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" || *bidDeadlineStr == "" || *openTimeStr == "" {
		return fmt.Errorf("name, bid-deadline and open-time are required")
	}

	bidDeadline, err := time.Parse(time.RFC3339, *bidDeadlineStr)
	if err != nil {
		return fmt.Errorf("invalid bid-deadline: %v", err)
	}

	openTime, err := time.Parse(time.RFC3339, *openTimeStr)
	if err != nil {
		return fmt.Errorf("invalid open-time: %v", err)
	}

	var invitedSuppliers []string
	if *invited != "" {
		for _, s := range strings.Split(*invited, ",") {
			invitedSuppliers = append(invitedSuppliers, strings.TrimSpace(s))
		}
	}

	req := common.CreateProjectRequest{
		Name:             *name,
		Description:      *desc,
		Method:           common.TenderMethod(*method),
		BidDeadline:      bidDeadline,
		OpenTime:         openTime,
		EvaluationMethod: common.EvaluationMethod(*evalMethod),
		InvitedSuppliers: invitedSuppliers,
	}

	var result common.ProjectDetailData
	if err := client.request(http.MethodPost, "/projects/create", req, &result); err != nil {
		return err
	}

	printResponse("Project created", result.Project)
	return nil
}

func cmdProjectUpdate(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("project-update", flag.ExitOnError)
	id := fs.String("id", "", "Project ID")
	name := fs.String("name", "", "Project name")
	desc := fs.String("desc", "", "Description")
	method := fs.String("method", "", "Tender method: public|invite")
	bidDeadlineStr := fs.String("bid-deadline", "", "Bid deadline (RFC3339)")
	openTimeStr := fs.String("open-time", "", "Open time (RFC3339)")
	evalMethod := fs.String("eval", "", "Evaluation method: lowest_price|comprehensive_score")
	invited := fs.String("invited", "", "Invited suppliers (comma separated)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == "" {
		return fmt.Errorf("project id is required")
	}

	req := common.UpdateProjectRequest{
		Name:        *name,
		Description: *desc,
	}

	if *method != "" {
		req.Method = common.TenderMethod(*method)
	}
	if *evalMethod != "" {
		req.EvaluationMethod = common.EvaluationMethod(*evalMethod)
	}
	if *bidDeadlineStr != "" {
		t, err := time.Parse(time.RFC3339, *bidDeadlineStr)
		if err != nil {
			return fmt.Errorf("invalid bid-deadline: %v", err)
		}
		req.BidDeadline = &t
	}
	if *openTimeStr != "" {
		t, err := time.Parse(time.RFC3339, *openTimeStr)
		if err != nil {
			return fmt.Errorf("invalid open-time: %v", err)
		}
		req.OpenTime = &t
	}
	if *invited != "" {
		var list []string
		for _, s := range strings.Split(*invited, ",") {
			list = append(list, strings.TrimSpace(s))
		}
		req.InvitedSuppliers = list
	}

	var result common.ProjectDetailData
	if err := client.request(http.MethodPost, "/projects/"+*id+"/update", req, &result); err != nil {
		return err
	}

	printResponse("Project updated", result.Project)
	return nil
}

func cmdProjectPublish(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("project-publish", flag.ExitOnError)
	id := fs.String("id", "", "Project ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" {
		return fmt.Errorf("project id is required")
	}

	var result common.ProjectDetailData
	if err := client.request(http.MethodPost, "/projects/"+*id+"/publish", nil, &result); err != nil {
		return err
	}

	printResponse("Project published", result.Project)
	return nil
}

func cmdProjectList(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("project-list", flag.ExitOnError)
	role := fs.String("role", "", "User role: supplier")
	supplierID := fs.String("supplier", "", "Supplier ID (for supplier role)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	path := "/projects"
	query := make([]string, 0)
	if *role != "" {
		query = append(query, "role="+*role)
	}
	if *supplierID != "" {
		query = append(query, "supplier_id="+*supplierID)
	}
	if len(query) > 0 {
		path += "?" + strings.Join(query, "&")
	}

	var result common.ProjectListData
	if err := client.request(http.MethodGet, path, nil, &result); err != nil {
		return err
	}

	printResponse("Projects", result.Projects)
	return nil
}

func cmdProjectGet(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("project-get", flag.ExitOnError)
	id := fs.String("id", "", "Project ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" {
		return fmt.Errorf("project id is required")
	}

	var result common.ProjectDetailData
	if err := client.request(http.MethodGet, "/projects/"+*id, nil, &result); err != nil {
		return err
	}

	printResponse("Project", result.Project)
	return nil
}

func cmdProjectBids(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("project-bids", flag.ExitOnError)
	id := fs.String("id", "", "Project ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" {
		return fmt.Errorf("project id is required")
	}

	var result struct {
		Bids []common.Bid `json:"bids"`
	}
	if err := client.request(http.MethodGet, "/projects/"+*id+"/bids", nil, &result); err != nil {
		return err
	}

	printResponse("Project bids", result.Bids)
	return nil
}

func cmdBidSubmit(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("bid-submit", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID")
	supplierID := fs.String("supplier", "", "Supplier ID")
	amountStr := fs.String("amount", "", "Amount (in cents)")
	techPlan := fs.String("tech", "", "Technical plan description")
	duration := fs.Int("duration", 0, "Duration days")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *projectID == "" || *supplierID == "" || *amountStr == "" {
		return fmt.Errorf("project, supplier and amount are required")
	}

	amount, err := strconv.ParseInt(*amountStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	req := common.SubmitBidRequest{
		ProjectID:     *projectID,
		SupplierID:    *supplierID,
		Amount:        amount,
		TechnicalPlan: *techPlan,
		DurationDays:  *duration,
	}

	var result common.BidDetailData
	if err := client.request(http.MethodPost, "/bids", req, &result); err != nil {
		return err
	}

	printResponse("Bid submitted", result.Bid)
	return nil
}

func cmdBidUpdate(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("bid-update", flag.ExitOnError)
	id := fs.String("id", "", "Bid ID")
	amountStr := fs.String("amount", "", "Amount (in cents)")
	techPlan := fs.String("tech", "", "Technical plan description")
	duration := fs.Int("duration", 0, "Duration days")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == "" || *amountStr == "" {
		return fmt.Errorf("bid id and amount are required")
	}

	amount, err := strconv.ParseInt(*amountStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	req := common.UpdateBidRequest{
		Amount:        amount,
		TechnicalPlan: *techPlan,
		DurationDays:  *duration,
	}

	var result common.BidDetailData
	if err := client.request(http.MethodPost, "/bids/"+*id+"/update", req, &result); err != nil {
		return err
	}

	printResponse("Bid updated", result.Bid)
	return nil
}

func cmdBidGet(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("bid-get", flag.ExitOnError)
	id := fs.String("id", "", "Bid ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" {
		return fmt.Errorf("bid id is required")
	}

	var result common.BidDetailData
	if err := client.request(http.MethodGet, "/bids/"+*id, nil, &result); err != nil {
		return err
	}

	printResponse("Bid", result.Bid)
	return nil
}

func cmdBidGetBySupplier(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("bid-get-by-supplier", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID")
	supplierID := fs.String("supplier", "", "Supplier ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *projectID == "" || *supplierID == "" {
		return fmt.Errorf("project and supplier are required")
	}

	var result common.BidDetailData
	if err := client.request(http.MethodGet, "/bids/"+*projectID+"/supplier/"+*supplierID, nil, &result); err != nil {
		return err
	}

	printResponse("Bid", result.Bid)
	return nil
}

func cmdOpenProject(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("open-project", flag.ExitOnError)
	id := fs.String("id", "", "Project ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" {
		return fmt.Errorf("project id is required")
	}

	req := common.OpenProjectRequest{ProjectID: *id}
	var result common.OpeningResultData
	if err := client.request(http.MethodPost, "/opening", req, &result); err != nil {
		return err
	}

	printResponse("Project opened", result.Result)
	return nil
}

func cmdOpeningResult(client *APIClient, args []string) error {
	fs := flag.NewFlagSet("opening-result", flag.ExitOnError)
	id := fs.String("id", "", "Project ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" {
		return fmt.Errorf("project id is required")
	}

	var result common.OpeningResultData
	if err := client.request(http.MethodGet, "/opening/"+*id, nil, &result); err != nil {
		return err
	}

	printResponse("Opening result", result.Result)
	return nil
}

func printUsage() {
	fmt.Println("Tender Management CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tender-cli [--server URL] <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  project-create       Create a new tender project")
	fmt.Println("  project-update       Update an existing draft project")
	fmt.Println("  project-publish      Publish a draft project")
	fmt.Println("  project-list         List projects")
	fmt.Println("  project-get          Get a project by ID")
	fmt.Println("  project-bids         List bids for a project")
	fmt.Println()
	fmt.Println("  bid-submit           Submit a bid")
	fmt.Println("  bid-update           Update an existing bid")
	fmt.Println("  bid-get              Get a bid by ID")
	fmt.Println("  bid-get-by-supplier  Get bid by project and supplier")
	fmt.Println()
	fmt.Println("  open-project         Open a project for bidding")
	fmt.Println("  opening-result       Get opening result")
	fmt.Println()
	fmt.Println("Global Flags:")
	fmt.Println("  --server             Server base URL (default: http://localhost:8080)")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := "http://localhost:8080"
	args := os.Args[1:]

	for i, arg := range args {
		if arg == "--server" && i+1 < len(args) {
			serverURL = args[i+1]
			args = append(args[:i], args[i+2:]...)
			break
		}
		if strings.HasPrefix(arg, "--server=") {
			serverURL = strings.TrimPrefix(arg, "--server=")
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	client := NewAPIClient(serverURL)
	cmd := args[0]
	cmdArgs := args[1:]

	var err error
	switch cmd {
	case "project-create":
		err = cmdProjectCreate(client, cmdArgs)
	case "project-update":
		err = cmdProjectUpdate(client, cmdArgs)
	case "project-publish":
		err = cmdProjectPublish(client, cmdArgs)
	case "project-list":
		err = cmdProjectList(client, cmdArgs)
	case "project-get":
		err = cmdProjectGet(client, cmdArgs)
	case "project-bids":
		err = cmdProjectBids(client, cmdArgs)
	case "bid-submit":
		err = cmdBidSubmit(client, cmdArgs)
	case "bid-update":
		err = cmdBidUpdate(client, cmdArgs)
	case "bid-get":
		err = cmdBidGet(client, cmdArgs)
	case "bid-get-by-supplier":
		err = cmdBidGetBySupplier(client, cmdArgs)
	case "open-project":
		err = cmdOpenProject(client, cmdArgs)
	case "opening-result":
		err = cmdOpeningResult(client, cmdArgs)
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		fmt.Println()
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Println("✗ Error:", err)
		os.Exit(1)
	}
}
