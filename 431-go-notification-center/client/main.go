package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"notification-center/common"
	"os"
	"strings"
	"time"
)

var (
	serverURL string
	printer   *Printer
)

func main() {
	printer = NewPrinter()
	
	flag.StringVar(&serverURL, "server", "", "Server URL (e.g., http://localhost:8080)")
	flag.Usage = printUsage
	flag.Parse()
	
	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}
	
	client := NewClient(serverURL)
	
	cmd := args[0]
	switch cmd {
	case "send":
		handleSend(client, args[1:])
	case "list":
		handleList(client, args[1:])
	case "mark-read":
		handleMarkRead(client, args[1:])
	case "count":
		handleCount(client, args[1:])
	case "template":
		handleTemplate(client, args[1:])
	case "activity":
		handleActivity(client, args[1:])
	case "logs":
		handleLogs(client, args[1:])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Notification Center Client

Usage:
  nc-client [flags] <command> [arguments]

Flags:
  -server string    Server URL (default: http://localhost:8080)

Commands:
  send          Send a notification
  list          List notifications
  mark-read     Mark notifications as read (batch)
  count         Get unread count
  template      Manage notification templates
  activity      Record user activity
  logs          List failed notification logs
  help          Show this help message

Use "nc-client <command> -help" for more information about a command.`)
}

func handleSend(client *Client, args []string) {
	fs := flag.NewFlagSet("send", flag.ExitOnError)
	notifType := fs.String("type", "system", "Notification type: system, approval, task, security")
	priority := fs.String("priority", "normal", "Priority: high, normal, low")
	sender := fs.String("sender", "", "Sender (required)")
	receiver := fs.String("receiver", "", "Receiver (required)")
	title := fs.String("title", "", "Title (required)")
	content := fs.String("content", "", "Content")
	templateID := fs.String("template", "", "Template ID")
	variables := fs.String("vars", "", "Variables JSON: {\"name\":\"value\"}")
	
	fs.Usage = func() {
		fmt.Println(`Send a notification

Usage:
  nc-client send [flags]

Flags:
  -type string       Notification type (system, approval, task, security) (default "system")
  -priority string   Priority (high, normal, low) (default "normal")
  -sender string     Sender (required)
  -receiver string   Receiver (required)
  -title string      Title (required)
  -content string    Content (or use -template)
  -template string   Template ID (instead of content)
  -vars string       Variables JSON for template: {"name":"value"}

Examples:
  nc-client send -sender system -receiver user001 -title "Welcome" -content "Hello!"
  nc-client send -type security -sender admin -receiver user001 -title "Alert" -content "Security breach detected"
  nc-client send -template tmpl-xxx -vars '{"user":"John","order":"ORD-123"}' -sender system -receiver user001`)
	}
	
	fs.Parse(args)
	
	if *sender == "" || *receiver == "" || *title == "" {
		fs.Usage()
		os.Exit(1)
	}
	
	req := common.SendRequest{
		Type:       common.NotificationType(*notifType),
		Priority:   common.Priority(*priority),
		Sender:     *sender,
		Receiver:   *receiver,
		Title:      *title,
		Content:    *content,
		TemplateID: *templateID,
	}
	
	if *variables != "" {
		var vars map[string]string
		if err := json.Unmarshal([]byte(*variables), &vars); err != nil {
			printer.PrintError(fmt.Errorf("invalid variables JSON: %w", err))
			os.Exit(1)
		}
		req.Variables = vars
	}
	
	resp, err := client.SendNotification(req)
	if err != nil {
		printer.PrintError(err)
		os.Exit(1)
	}
	
	printer.PrintSendResult(resp)
}

func handleList(client *Client, args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	receiver := fs.String("receiver", "", "Receiver (required)")
	notifType := fs.String("type", "", "Filter by type: system, approval, task, security")
	status := fs.String("status", "", "Filter by status: unread, read")
	startTime := fs.String("start", "", "Start time (RFC3339)")
	endTime := fs.String("end", "", "End time (RFC3339)")
	includeArchived := fs.Bool("archived", false, "Include archived notifications")
	page := fs.Int("page", 0, "Page number (1-based)")
	pageSize := fs.Int("size", 0, "Page size")
	
	fs.Usage = func() {
		fmt.Println(`List notifications

Usage:
  nc-client list [flags]

Flags:
  -receiver string   Receiver (required)
  -type string       Filter by type (system, approval, task, security)
  -status string     Filter by status (unread, read)
  -start string      Start time (RFC3339 format)
  -end string        End time (RFC3339 format)
  -archived          Include archived notifications
  -page int          Page number (1-based)
  -size int          Page size

Examples:
  nc-client list -receiver user001
  nc-client list -receiver user001 -type security -status unread
  nc-client list -receiver user001 -archived -page 1 -size 20`)
	}
	
	fs.Parse(args)
	
	if *receiver == "" {
		fs.Usage()
		os.Exit(1)
	}
	
	var t *common.NotificationType
	if *notifType != "" {
		nt := common.NotificationType(*notifType)
		t = &nt
	}
	
	var s *common.ReadStatus
	if *status != "" {
		rs := common.ReadStatus(*status)
		s = &rs
	}
	
	var start, end *time.Time
	if *startTime != "" {
		if t, err := time.Parse(time.RFC3339, *startTime); err == nil {
			start = &t
		}
	}
	if *endTime != "" {
		if t, err := time.Parse(time.RFC3339, *endTime); err == nil {
			end = &t
		}
	}
	
	resp, err := client.ListNotifications(*receiver, t, s, start, end, *includeArchived, *page, *pageSize)
	if err != nil {
		printer.PrintError(err)
		os.Exit(1)
	}
	
	printer.PrintNotificationList(resp)
}

func handleMarkRead(client *Client, args []string) {
	fs := flag.NewFlagSet("mark-read", flag.ExitOnError)
	receiver := fs.String("receiver", "", "Receiver (required)")
	ids := fs.String("ids", "", "Comma-separated notification IDs (required)")
	
	fs.Usage = func() {
		fmt.Println(`Mark notifications as read (batch, max 100)

Usage:
  nc-client mark-read [flags]

Flags:
  -receiver string   Receiver (required)
  -ids string        Comma-separated notification IDs (required, max 100)

Examples:
  nc-client mark-read -receiver user001 -ids "id1,id2,id3"`)
	}
	
	fs.Parse(args)
	
	if *receiver == "" || *ids == "" {
		fs.Usage()
		os.Exit(1)
	}
	
	notificationIDs := strings.Split(*ids, ",")
	for i, id := range notificationIDs {
		notificationIDs[i] = strings.TrimSpace(id)
	}
	
	resp, err := client.MarkAsRead(*receiver, notificationIDs)
	if err != nil {
		printer.PrintError(err)
		os.Exit(1)
	}
	
	printer.PrintMarkReadResult(resp)
}

func handleCount(client *Client, args []string) {
	fs := flag.NewFlagSet("count", flag.ExitOnError)
	receiver := fs.String("receiver", "", "Receiver (required)")
	
	fs.Usage = func() {
		fmt.Println(`Get unread notification count

Usage:
  nc-client count [flags]

Flags:
  -receiver string   Receiver (required)

Examples:
  nc-client count -receiver user001`)
	}
	
	fs.Parse(args)
	
	if *receiver == "" {
		fs.Usage()
		os.Exit(1)
	}
	
	resp, err := client.GetUnreadCount(*receiver)
	if err != nil {
		printer.PrintError(err)
		os.Exit(1)
	}
	
	printer.PrintUnreadCount(resp)
}

func handleTemplate(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println(`Template management commands:

Usage:
  nc-client template <subcommand> [flags]

Subcommands:
  create    Create a new template
  update    Update an existing template
  list      List all templates
  delete    Delete a template
  help      Show this help message`)
		os.Exit(1)
	}
	
	subcmd := args[0]
	switch subcmd {
	case "create":
		handleTemplateCreate(client, args[1:])
	case "update":
		handleTemplateUpdate(client, args[1:])
	case "list":
		handleTemplateList(client, args[1:])
	case "delete":
		handleTemplateDelete(client, args[1:])
	case "help":
		fmt.Println(`Template management commands:

  create    Create a new template
  update    Update an existing template
  list      List all templates
  delete    Delete a template`)
	default:
		fmt.Printf("Unknown template subcommand: %s\n", subcmd)
		os.Exit(1)
	}
}

func handleTemplateCreate(client *Client, args []string) {
	fs := flag.NewFlagSet("template create", flag.ExitOnError)
	name := fs.String("name", "", "Template name (required)")
	notifType := fs.String("type", "system", "Notification type")
	title := fs.String("title", "", "Title template (required)")
	content := fs.String("content", "", "Content template (required)")
	vars := fs.String("vars", "", "Comma-separated variable names: name,order,date")
	
	fs.Usage = func() {
		fmt.Println(`Create a notification template

Usage:
  nc-client template create [flags]

Flags:
  -name string      Template name (required)
  -type string      Notification type (default "system")
  -title string     Title template (required, use {{var}} for variables)
  -content string   Content template (required)
  -vars string      Comma-separated variable names

Examples:
  nc-client template create -name "Order Alert" -title "Order {{order}} Status Update" 
    -content "Dear {{user}}, your order {{order}} is now {{status}}." 
    -vars "user,order,status"`)
	}
	
	fs.Parse(args)
	
	if *name == "" || *title == "" || *content == "" {
		fs.Usage()
		os.Exit(1)
	}
	
	var variables []string
	if *vars != "" {
		parts := strings.Split(*vars, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				variables = append(variables, trimmed)
			}
		}
	}
	
	req := common.TemplateCreateRequest{
		Name:      *name,
		Type:      common.NotificationType(*notifType),
		Title:     *title,
		Content:   *content,
		Variables: variables,
	}
	
	template, err := client.CreateTemplate(req)
	if err != nil {
		printer.PrintError(err)
		os.Exit(1)
	}
	
	printer.PrintTemplate(*template)
}

func handleTemplateUpdate(client *Client, args []string) {
	fs := flag.NewFlagSet("template update", flag.ExitOnError)
	id := fs.String("id", "", "Template ID (required)")
	name := fs.String("name", "", "New name")
	notifType := fs.String("type", "", "New type")
	title := fs.String("title", "", "New title")
	content := fs.String("content", "", "New content")
	vars := fs.String("vars", "", "New variables (comma-separated)")
	
	fs.Usage = func() {
		fmt.Println(`Update a notification template

Usage:
  nc-client template update [flags]

Flags:
  -id string        Template ID (required)
  -name string      New name
  -type string      New type
  -title string     New title
  -content string   New content
  -vars string      New variables`)
	}
	
	fs.Parse(args)
	
	if *id == "" {
		fs.Usage()
		os.Exit(1)
	}
	
	req := common.TemplateUpdateRequest{
		ID: *id,
	}
	if *name != "" {
		req.Name = name
	}
	if *notifType != "" {
		nt := common.NotificationType(*notifType)
		req.Type = &nt
	}
	if *title != "" {
		req.Title = title
	}
	if *content != "" {
		req.Content = content
	}
	if *vars != "" {
		parts := strings.Split(*vars, ",")
		var variables []string
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				variables = append(variables, trimmed)
			}
		}
		req.Variables = &variables
	}
	
	template, err := client.UpdateTemplate(req)
	if err != nil {
		printer.PrintError(err)
		os.Exit(1)
	}
	
	printer.PrintTemplate(*template)
}

func handleTemplateList(client *Client, args []string) {
	templates, err := client.ListTemplates()
	if err != nil {
		printer.PrintError(err)
		os.Exit(1)
	}
	
	printer.PrintTemplateList(templates)
}

func handleTemplateDelete(client *Client, args []string) {
	fs := flag.NewFlagSet("template delete", flag.ExitOnError)
	id := fs.String("id", "", "Template ID (required)")
	
	fs.Usage = func() {
		fmt.Println(`Delete a notification template

Usage:
  nc-client template delete [flags]

Flags:
  -id string   Template ID (required)`)
	}
	
	fs.Parse(args)
	
	if *id == "" {
		fs.Usage()
		os.Exit(1)
	}
	
	err := client.DeleteTemplate(*id)
	if err != nil {
		printer.PrintError(err)
		os.Exit(1)
	}
	
	printer.PrintSuccess(fmt.Sprintf("Template %s deleted", *id))
}

func handleActivity(client *Client, args []string) {
	fs := flag.NewFlagSet("activity", flag.ExitOnError)
	userID := fs.String("user", "", "User ID (required)")
	
	fs.Usage = func() {
		fmt.Println(`Record user activity (prevents marking as inactive)

Usage:
  nc-client activity [flags]

Flags:
  -user string   User ID (required)

Examples:
  nc-client activity -user user001`)
	}
	
	fs.Parse(args)
	
	if *userID == "" {
		fs.Usage()
		os.Exit(1)
	}
	
	err := client.RecordActivity(*userID)
	if err != nil {
		printer.PrintError(err)
		os.Exit(1)
	}
	
	printer.PrintSuccess(fmt.Sprintf("Activity recorded for user %s", *userID))
}

func handleLogs(client *Client, args []string) {
	logs, err := client.ListFailedLogs()
	if err != nil {
		printer.PrintError(err)
		os.Exit(1)
	}
	
	printer.PrintFailedLogList(logs)
}
