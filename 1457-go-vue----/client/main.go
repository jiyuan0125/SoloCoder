package main

import (
	"bytes"
	"complaint-system/common"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

func (c *Client) doJSON(method, path string, reqBody, respBody interface{}) error {
	var body io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("%s", errResp.Error)
	}

	if respBody != nil && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil && err != io.EOF {
			return err
		}
	} else if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil && err != io.EOF {
			return err
		}
	}

	return nil
}

func (c *Client) CreateTicket(req *common.CreateTicketRequest) (string, error) {
	var resp common.CreateTicketResponse
	if err := c.doJSON("POST", "/api/tickets", req, &resp); err != nil {
		return "", err
	}
	return resp.TicketNo, nil
}

func (c *Client) GetTicket(ticketNo string) (*common.TicketResponse, error) {
	var resp common.TicketResponse
	if err := c.doJSON("GET", "/api/tickets/"+ticketNo, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ListTickets() ([]common.TicketResponse, error) {
	var resp common.ListTicketsResponse
	if err := c.doJSON("GET", "/api/tickets", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Tickets, nil
}

func (c *Client) DispatchTicket(ticketNo string, req *common.DispatchTicketRequest) error {
	return c.doJSON("POST", "/api/tickets/"+ticketNo+"/dispatch", req, nil)
}

func (c *Client) StartProcessing(ticketNo string) error {
	return c.doJSON("POST", "/api/tickets/"+ticketNo+"/start-processing", nil, nil)
}

func (c *Client) CompleteProcessing(ticketNo string, req *common.CompleteProcessingRequest) error {
	return c.doJSON("POST", "/api/tickets/"+ticketNo+"/complete-processing", req, nil)
}

func (c *Client) ReviewTicket(ticketNo string, req *common.ReviewRequest) error {
	return c.doJSON("POST", "/api/tickets/"+ticketNo+"/review", req, nil)
}

func (c *Client) GetStatistics() (*common.StatisticsResponse, error) {
	var resp common.StatisticsResponse
	if err := c.doJSON("GET", "/api/statistics", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func printUsage() {
	fmt.Println(`Usage: complaint-client [command] [options]

Commands:
  create     Create a new complaint ticket
  get        Get a ticket by ticket no
  list       List all tickets
  dispatch   Dispatch a ticket
  start      Start processing a ticket
  complete   Complete processing a ticket
  review     Review a ticket
  stats      Get statistics

Global Options:
  -server     Server address (default: http://localhost:8080)

Use "complaint-client [command] -help" for detailed command usage.`)
}

func printCreateUsage() {
	fmt.Println(`Usage: complaint-client create [options]

Options:
  -name       Complainer name (required)
  -phone      Contact phone (required)
  -content    Complaint content (required)
  -channel    Complaint channel: phone, wechat, app, onsite (required)
  -type       Complaint type: service_quality, product_quality, logistics, after_sales, other (required)
  -region     Region (required)
  -external   External system ID (optional)`)
}

func printDispatchUsage() {
	fmt.Println(`Usage: complaint-client dispatch -ticket <ticket_no> [options]

Options:
  -ticket     Ticket no (required)
  -dept       Responsible department (required)
  -person     Responsible person (required)
  -urgency    Urgency level: normal, urgent, super (required)
  -dispatcher Dispatcher ID (optional)`)
}

func printCompleteUsage() {
	fmt.Println(`Usage: complaint-client complete -ticket <ticket_no> -result <result>

Options:
  -ticket     Ticket no (required)
  -result     Processing result (required)`)
}

func printReviewUsage() {
	fmt.Println(`Usage: complaint-client review -ticket <ticket_no> -result <result> [options]

Options:
  -ticket     Ticket no (required)
  -result     Review result: satisfied, basicly_satisfied, dissatisfied (required)
  -remark     Review remark (optional)
  -reviewer   Reviewer ID (optional)`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var serverAddr string
	flag.StringVar(&serverAddr, "server", "http://localhost:8080", "Server address")

	args := os.Args[1:]
	cmd := args[0]

	for i, arg := range args {
		if arg == "-server" && i+1 < len(args) {
			serverAddr = args[i+1]
			args = append(args[:i], args[i+2:]...)
			if len(args) == 0 {
				printUsage()
				os.Exit(1)
			}
			cmd = args[0]
			break
		}
	}

	client := NewClient(serverAddr)

	switch cmd {
	case "create":
		handleCreate(client, args[1:])
	case "get":
		handleGet(client, args[1:])
	case "list":
		handleList(client)
	case "dispatch":
		handleDispatch(client, args[1:])
	case "start":
		handleStart(client, args[1:])
	case "complete":
		handleComplete(client, args[1:])
	case "review":
		handleReview(client, args[1:])
	case "stats":
		handleStats(client)
	case "-help", "--help", "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func handleCreate(client *Client, args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	var name, phone, content, channel, typ, region, external string
	fs.StringVar(&name, "name", "", "Complainer name")
	fs.StringVar(&phone, "phone", "", "Contact phone")
	fs.StringVar(&content, "content", "", "Complaint content")
	fs.StringVar(&channel, "channel", "", "Complaint channel")
	fs.StringVar(&typ, "type", "", "Complaint type")
	fs.StringVar(&region, "region", "", "Region")
	fs.StringVar(&external, "external", "", "External system ID")
	fs.Usage = printCreateUsage
	fs.Parse(args)

	if name == "" || phone == "" || content == "" || channel == "" || typ == "" || region == "" {
		printCreateUsage()
		os.Exit(1)
	}

	req := &common.CreateTicketRequest{
		ComplainerName:   name,
		ContactPhone:     phone,
		Content:          content,
		ComplaintChannel: common.ComplaintChannel(channel),
		ComplaintType:    common.ComplaintType(typ),
		Region:           region,
		ExternalSystemID: external,
	}

	ticketNo, err := client.CreateTicket(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Ticket created: %s\n", ticketNo)
}

func handleGet(client *Client, args []string) {
	if len(args) != 2 || args[0] != "-ticket" {
		fmt.Println("Usage: complaint-client get -ticket <ticket_no>")
		os.Exit(1)
	}
	ticket, err := client.GetTicket(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	printTicket(ticket)
}

func handleList(client *Client) {
	tickets, err := client.ListTickets()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	for _, t := range tickets {
		printTicket(&t)
		fmt.Println("---")
	}
	fmt.Printf("Total: %d tickets\n", len(tickets))
}

func handleDispatch(client *Client, args []string) {
	fs := flag.NewFlagSet("dispatch", flag.ExitOnError)
	var ticketNo, dept, person, urgency, dispatcher string
	fs.StringVar(&ticketNo, "ticket", "", "Ticket no")
	fs.StringVar(&dept, "dept", "", "Responsible department")
	fs.StringVar(&person, "person", "", "Responsible person")
	fs.StringVar(&urgency, "urgency", "", "Urgency level")
	fs.StringVar(&dispatcher, "dispatcher", "", "Dispatcher ID")
	fs.Usage = printDispatchUsage
	fs.Parse(args)

	if ticketNo == "" || dept == "" || person == "" || urgency == "" {
		printDispatchUsage()
		os.Exit(1)
	}

	req := &common.DispatchTicketRequest{
		ResponsibleDepartment: dept,
		ResponsiblePerson:     person,
		UrgencyLevel:          common.UrgencyLevel(urgency),
		DispatcherID:          dispatcher,
	}

	if err := client.DispatchTicket(ticketNo, req); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Ticket dispatched successfully")
}

func handleStart(client *Client, args []string) {
	if len(args) != 2 || args[0] != "-ticket" {
		fmt.Println("Usage: complaint-client start -ticket <ticket_no>")
		os.Exit(1)
	}
	if err := client.StartProcessing(args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Started processing")
}

func handleComplete(client *Client, args []string) {
	fs := flag.NewFlagSet("complete", flag.ExitOnError)
	var ticketNo, result string
	fs.StringVar(&ticketNo, "ticket", "", "Ticket no")
	fs.StringVar(&result, "result", "", "Processing result")
	fs.Usage = printCompleteUsage
	fs.Parse(args)

	if ticketNo == "" || result == "" {
		printCompleteUsage()
		os.Exit(1)
	}

	req := &common.CompleteProcessingRequest{
		ProcessingResult: result,
	}

	if err := client.CompleteProcessing(ticketNo, req); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Processing completed")
}

func handleReview(client *Client, args []string) {
	fs := flag.NewFlagSet("review", flag.ExitOnError)
	var ticketNo, result, remark, reviewer string
	fs.StringVar(&ticketNo, "ticket", "", "Ticket no")
	fs.StringVar(&result, "result", "", "Review result")
	fs.StringVar(&remark, "remark", "", "Review remark")
	fs.StringVar(&reviewer, "reviewer", "", "Reviewer ID")
	fs.Usage = printReviewUsage
	fs.Parse(args)

	if ticketNo == "" || result == "" {
		printReviewUsage()
		os.Exit(1)
	}

	req := &common.ReviewRequest{
		ReviewResult: common.ReviewResult(result),
		Remark:       remark,
		ReviewerID:   reviewer,
	}

	if err := client.ReviewTicket(ticketNo, req); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Review completed")
}

func handleStats(client *Client) {
	stats, err := client.GetStatistics()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Today New:        %d\n", stats.TodayNewCount)
	fmt.Printf("Processing:       %d\n", stats.ProcessingCount)
	fmt.Printf("Closed:           %d\n", stats.ClosedCount)
	fmt.Printf("Avg Processing:   %.2f hours\n", stats.AvgProcessingHours)
	fmt.Printf("Review Satisfaction: %.2f%%\n", stats.ReviewSatisfaction)
}

func printTicket(t *common.TicketResponse) {
	fmt.Printf("Ticket No: %s\n", t.TicketNo)
	fmt.Printf("Complainer: %s (%s)\n", t.ComplainerName, t.ContactPhone)
	fmt.Printf("Channel: %s | Type: %s | Region: %s\n", t.ComplaintChannel, t.ComplaintType, t.Region)
	fmt.Printf("Status: %s\n", t.Status)
	if t.UrgencyLevel != nil {
		fmt.Printf("Urgency: %s\n", *t.UrgencyLevel)
	}
	if t.ResponsibleDepartment != "" {
		fmt.Printf("Responsible: %s - %s\n", t.ResponsibleDepartment, t.ResponsiblePerson)
	}
	fmt.Printf("Created: %s\n", t.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Deadline: %s\n", t.ResponseDeadline.Format("2006-01-02 15:04:05"))
	if t.ProcessedAt != nil {
		fmt.Printf("Processed: %s\n", t.ProcessedAt.Format("2006-01-02 15:04:05"))
	}
	if t.ClosedAt != nil {
		fmt.Printf("Closed: %s\n", t.ClosedAt.Format("2006-01-02 15:04:05"))
	}
	if t.RetryCount > 0 {
		fmt.Printf("Retry Count: %d\n", t.RetryCount)
	}
	if t.IsEscalated {
		fmt.Printf("Escalated: Yes\n")
	}
	if t.ReviewResult != nil {
		fmt.Printf("Review: %s\n", *t.ReviewResult)
	}
	fmt.Printf("Content: %s\n", t.Content)
	if t.ProcessingResult != "" {
		fmt.Printf("Result: %s\n", t.ProcessingResult)
	}
	if t.ReviewRemark != "" {
		fmt.Printf("Review Remark: %s\n", t.ReviewRemark)
	}
	if t.ExternalSystemID != "" {
		fmt.Printf("External ID: %s\n", t.ExternalSystemID)
	}
}
