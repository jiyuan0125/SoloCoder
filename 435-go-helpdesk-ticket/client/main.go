package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go-helpdesk-ticket/common"
	"os"
	"strings"
)

func main() {
	client := NewAPIClient()

	flag.Usage = printUsage
	flag.Parse()

	if flag.NArg() == 0 {
		printUsage()
		os.Exit(1)
	}

	command := flag.Arg(0)

	switch command {
	case "create":
		handleCreate(client)
	case "get":
		handleGet(client)
	case "list":
		handleList(client)
	case "reassign":
		handleReassign(client)
	case "status":
		handleStatus(client)
	case "rate":
		handleRate(client)
	case "link":
		handleLink(client)
	case "transfer":
		handleTransfer(client)
	case "logs":
		handleLogs(client)
	case "stats":
		handleStats(client)
	case "handlers":
		handleHandlers(client)
	case "templates":
		handleTemplates(client)
	case "classify":
		handleClassify(client)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`IT Helpdesk Ticket CLI

Usage:
  ticket-cli <command> [options]

Commands:
  create      Create a new ticket
  get         Get ticket by ID
  list        List all tickets
  reassign    Reassign ticket to another handler
  status      Update ticket status
  rate        Rate a closed ticket
  link        Link tickets (parent-child)
  transfer    Transfer ticket to another team
  logs        Get operation logs for a ticket
  stats       Get statistics report
  handlers    List all handlers
  templates   List quick reply templates
  classify    Auto-classify ticket content

Examples:
  ticket-cli create -title "网络问题" -desc "无法连接WiFi" -cat network -prio normal -uid u001 -uname "张三"
  ticket-cli list
  ticket-cli get TKxxxxxx
  ticket-cli status TKxxxxxx -s in_progress -o h1
  ticket-cli reassign TKxxxxxx -hid h2 -hname "李网络" -reason "我有事请假"`)
}

func handleCreate(client *APIClient) {
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	title := createCmd.String("title", "", "Ticket title")
	desc := createCmd.String("desc", "", "Ticket description")
	cat := createCmd.String("cat", "", "Category: network/hardware/software/account")
	prio := createCmd.String("prio", "normal", "Priority: normal/urgent/critical")
	uid := createCmd.String("uid", "", "Submitter ID")
	uname := createCmd.String("uname", "", "Submitter name")
	autoClassify := createCmd.Bool("auto", false, "Enable auto classification")

	createCmd.Parse(os.Args[2:])

	if *title == "" || *desc == "" || *uid == "" || *uname == "" {
		fmt.Println("Error: -title, -desc, -uid, -uname are required")
		os.Exit(1)
	}

	req := &common.CreateTicketRequest{
		Title:         *title,
		Description:   *desc,
		Priority:      common.TicketPriority(*prio),
		SubmitterID:   *uid,
		SubmitterName: *uname,
		AutoClassify:  *autoClassify,
	}

	if *cat != "" {
		req.Category = common.TicketCategory(*cat)
	}

	ticket, err := client.CreateTicket(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printTicket(ticket)
}

func handleGet(client *APIClient) {
	if flag.NArg() < 2 {
		fmt.Println("Usage: ticket-cli get <ticket_id>")
		os.Exit(1)
	}

	id := flag.Arg(1)
	ticket, err := client.GetTicket(id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printTicket(ticket)
}

func handleList(client *APIClient) {
	tickets, err := client.ListTickets()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total tickets: %d\n\n", len(tickets))
	for _, t := range tickets {
		fmt.Printf("ID: %s\n", t.ID)
		fmt.Printf("  Title: %s\n", t.Title)
		fmt.Printf("  Status: %s | Priority: %s | Category: %s\n", t.Status, t.Priority, t.Category)
		fmt.Printf("  Assignee: %s\n", t.AssigneeName)
		fmt.Println()
	}
}

func handleReassign(client *APIClient) {
	if flag.NArg() < 2 {
		fmt.Println("Usage: ticket-cli reassign <ticket_id> -hid <handler_id> -hname <handler_name> -reason <reason>")
		os.Exit(1)
	}

	id := flag.Arg(1)
	reassignCmd := flag.NewFlagSet("reassign", flag.ExitOnError)
	hid := reassignCmd.String("hid", "", "New handler ID")
	hname := reassignCmd.String("hname", "", "New handler name")
	reason := reassignCmd.String("reason", "", "Reassign reason")

	reassignCmd.Parse(os.Args[3:])

	if *hid == "" || *hname == "" || *reason == "" {
		fmt.Println("Error: -hid, -hname, -reason are required")
		os.Exit(1)
	}

	req := &common.ReassignTicketRequest{
		NewAssigneeID:   *hid,
		NewAssigneeName: *hname,
		Reason:          *reason,
	}

	if err := client.ReassignTicket(id, req); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Ticket reassigned successfully")
}

func handleStatus(client *APIClient) {
	if flag.NArg() < 2 {
		fmt.Println("Usage: ticket-cli status <ticket_id> -s <status> -o <operator_id> [-c <comment>]")
		os.Exit(1)
	}

	id := flag.Arg(1)
	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)
	status := statusCmd.String("s", "", "Status: new/assigned/in_progress/resolved/closed")
	operator := statusCmd.String("o", "", "Operator ID")
	comment := statusCmd.String("c", "", "Optional comment")

	statusCmd.Parse(os.Args[3:])

	if *status == "" || *operator == "" {
		fmt.Println("Error: -s and -o are required")
		os.Exit(1)
	}

	req := &common.UpdateStatusRequest{
		Status:     common.TicketStatus(*status),
		OperatorID: *operator,
		Comment:    *comment,
	}

	if err := client.UpdateStatus(id, req); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Status updated successfully")
}

func handleRate(client *APIClient) {
	if flag.NArg() < 2 {
		fmt.Println("Usage: ticket-cli rate <ticket_id> -r <rating_1-5> -uid <user_id> -uname <user_name> [-c <comment>]")
		os.Exit(1)
	}

	id := flag.Arg(1)
	rateCmd := flag.NewFlagSet("rate", flag.ExitOnError)
	rating := rateCmd.Int("r", 0, "Rating: 1-5")
	uid := rateCmd.String("uid", "", "User ID (must be submitter)")
	uname := rateCmd.String("uname", "", "User name")
	comment := rateCmd.String("c", "", "Optional comment")

	rateCmd.Parse(os.Args[3:])

	if *rating < 1 || *rating > 5 || *uid == "" || *uname == "" {
		fmt.Println("Error: -r (1-5), -uid, -uname are required")
		os.Exit(1)
	}

	req := &common.RateTicketRequest{
		Rating:   *rating,
		UserID:   *uid,
		UserName: *uname,
		Comment:  *comment,
	}

	if err := client.RateTicket(id, req); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Ticket rated successfully")
}

func handleLink(client *APIClient) {
	linkCmd := flag.NewFlagSet("link", flag.ExitOnError)
	parent := linkCmd.String("parent", "", "Parent ticket ID")
	child := linkCmd.String("child", "", "Child ticket ID")

	linkCmd.Parse(os.Args[2:])

	if *parent == "" || *child == "" {
		fmt.Println("Usage: ticket-cli link -parent <parent_id> -child <child_id>")
		os.Exit(1)
	}

	req := &common.LinkTicketsRequest{
		ParentID: *parent,
		ChildID:  *child,
	}

	if err := client.LinkTickets(req); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Tickets linked successfully")
}

func handleTransfer(client *APIClient) {
	if flag.NArg() < 2 {
		fmt.Println("Usage: ticket-cli transfer <ticket_id> -team <team> -reason <reason> -o <operator_id> -oname <operator_name>")
		os.Exit(1)
	}

	id := flag.Arg(1)
	transferCmd := flag.NewFlagSet("transfer", flag.ExitOnError)
	team := transferCmd.String("team", "", "Target team: network/hardware/software/account")
	reason := transferCmd.String("reason", "", "Transfer reason")
	oid := transferCmd.String("o", "", "Operator ID")
	oname := transferCmd.String("oname", "", "Operator name")

	transferCmd.Parse(os.Args[3:])

	if *team == "" || *reason == "" || *oid == "" || *oname == "" {
		fmt.Println("Error: -team, -reason, -o, -oname are required")
		os.Exit(1)
	}

	req := &common.TransferTicketRequest{
		TargetTeam:   *team,
		Reason:       *reason,
		OperatorID:   *oid,
		OperatorName: *oname,
	}

	if err := client.TransferTicket(id, req); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Ticket transferred successfully")
}

func handleLogs(client *APIClient) {
	if flag.NArg() < 2 {
		fmt.Println("Usage: ticket-cli logs <ticket_id>")
		os.Exit(1)
	}

	id := flag.Arg(1)
	logs, err := client.GetLogs(id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Operation logs for ticket %s:\n\n", id)
	for i, log := range logs {
		fmt.Printf("%d. [%s] %s\n", i+1, log.CreatedAt.Format("2006-01-02 15:04:05"), log.OperationType)
		fmt.Printf("   Operator: %s (%s)\n", log.OperatorName, log.OperatorID)
		if log.OldValue != "" || log.NewValue != "" {
			fmt.Printf("   Old: %s -> New: %s\n", log.OldValue, log.NewValue)
		}
		if log.Comment != "" {
			fmt.Printf("   Comment: %s\n", log.Comment)
		}
		fmt.Println()
	}
}

func handleStats(client *APIClient) {
	stats, err := client.GetStatistics()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Statistics Report ===")
	fmt.Printf("Total Tickets: %d\n", stats.TotalTickets)
	fmt.Printf("Average Handle Time: %.2f hours\n", stats.AverageHandleTime)
	fmt.Printf("Overall SLA Rate: %.1f%%\n\n", stats.OverallSLARate)

	fmt.Println("By Category:")
	for cat, s := range stats.ByCategory {
		fmt.Printf("\n  %s:\n", strings.ToUpper(string(cat)))
		fmt.Printf("    Count: %d\n", s.Count)
		fmt.Printf("    Avg Handle Time: %.2f hours\n", s.HandleTimeHours)
		fmt.Printf("    SLA Rate: %.1f%%\n", s.SLARate)
	}
}

func handleHandlers(client *APIClient) {
	handlers, err := client.GetHandlers()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Handlers ===")
	for id, h := range handlers {
		role := "Handler"
		if h.IsManager {
			role = "Manager"
		}
		fmt.Printf("ID: %s | Name: %s | Category: %s | Level: %d | Role: %s\n",
			id, h.Name, h.Category, h.Level, role)
	}
}

func handleTemplates(client *APIClient) {
	templates, err := client.GetTemplates()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Quick Reply Templates ===")
	for id, t := range templates {
		fmt.Printf("\nID: %s\n", id)
		fmt.Printf("Title: %s\n", t.Title)
		fmt.Printf("Category: %s\n", t.Category)
		fmt.Printf("Content: %s\n", t.Content)
	}
}

func handleClassify(client *APIClient) {
	classifyCmd := flag.NewFlagSet("classify", flag.ExitOnError)
	title := classifyCmd.String("title", "", "Title text")
	desc := classifyCmd.String("desc", "", "Description text")

	classifyCmd.Parse(os.Args[2:])

	if *title == "" && *desc == "" {
		fmt.Println("Usage: ticket-cli classify -title <text> -desc <text>")
		os.Exit(1)
	}

	category, err := client.AutoClassify(*title, *desc)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Auto-classified category: %s\n", category)
}

func printTicket(t *common.Ticket) {
	data, _ := json.MarshalIndent(t, "", "  ")
	fmt.Println(string(data))
}
