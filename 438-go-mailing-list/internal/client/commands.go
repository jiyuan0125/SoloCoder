package client

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"go-mailing-list/pkg/protocol"
)

type CLI struct {
	client *APIClient
}

func NewCLI(baseURL string) *CLI {
	return &CLI{
		client: NewAPIClient(baseURL),
	}
}

func (c *CLI) Run(args []string) {
	if len(args) < 1 {
		c.printUsage()
		os.Exit(1)
	}

	command := args[0]
	switch command {
	case "list-create":
		c.cmdListCreate(args[1:])
	case "list-get":
		c.cmdListGet(args[1:])
	case "list-list":
		c.cmdListList(args[1:])
	case "list-pause":
		c.cmdListPause(args[1:])
	case "list-resume":
		c.cmdListResume(args[1:])
	case "subscribe":
		c.cmdSubscribe(args[1:])
	case "unsubscribe":
		c.cmdUnsubscribe(args[1:])
	case "subscribers":
		c.cmdSubscribers(args[1:])
	case "import":
		c.cmdImport(args[1:])
	case "template-create":
		c.cmdTemplateCreate(args[1:])
	case "template-list":
		c.cmdTemplateList(args[1:])
	case "send":
		c.cmdSend(args[1:])
	case "task-get":
		c.cmdTaskGet(args[1:])
	case "task-list":
		c.cmdTaskList(args[1:])
	case "records":
		c.cmdRecords(args[1:])
	case "alerts":
		c.cmdAlerts(args[1:])
	case "alert-resolve":
		c.cmdAlertResolve(args[1:])
	case "stats":
		c.cmdStats(args[1:])
	case "help", "-h", "--help":
		c.printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		c.printUsage()
		os.Exit(1)
	}
}

func (c *CLI) printUsage() {
	fmt.Println(`Mailing List CLI

Usage:
  mlc [command] [options]

Commands:
  list-create <name> [description]      Create a new mailing list
  list-get <list-id>                     Get mailing list details
  list-list                               List all mailing lists
  list-pause <list-id>                   Pause a mailing list
  list-resume <list-id>                  Resume a mailing list
  subscribe <list-id> <email> [name]     Subscribe to a list
  unsubscribe <list-id> <email>          Unsubscribe from a list
  subscribers <list-id>                   List subscribers of a list
  import <list-id> <csv-file>             Import subscribers from CSV
  template-create <name> <subject> [html] [text]  Create email template
  template-list                            List all templates
  send <list-id> [options]                 Send email campaign
    Options:
      --template <id>       Use template ID
      --subject <text>      Email subject (if not using template)
      --html <file>         HTML body file
      --text <file>         Text body file
      --schedule <time>     Schedule for later (RFC3339 format)
      --abtest              Enable A/B testing
      --subjectA <text>     Subject for group A
      --subjectB <text>     Subject for group B
  task-get <task-id>                      Get task details
  task-list                                List all tasks
  records <task-id>                        Get send records for a task
  alerts                                   List all alerts
  alert-resolve <alert-id>                 Resolve an alert
  stats                                    Show system statistics
  help                                     Show this help message

Examples:
  mlc list-create "Newsletter" "Weekly newsletter"
  mlc subscribe newsletter-123 user@example.com "John Doe"
  mlc send newsletter-123 --template welcome-001
  mlc send newsletter-123 --subject "Hello" --html body.html
`)
}

func (c *CLI) cmdListCreate(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: list-create <name> [description]")
		os.Exit(1)
	}

	name := args[0]
	description := ""
	if len(args) > 1 {
		description = strings.Join(args[1:], " ")
	}

	list, err := c.client.CreateMailingList(name, description)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printMailingList(list)
}

func (c *CLI) cmdListGet(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: list-get <list-id>")
		os.Exit(1)
	}

	list, err := c.client.GetMailingList(args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printMailingList(list)
}

func (c *CLI) cmdListList(args []string) {
	lists, err := c.client.ListMailingLists()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tPAUSED\tCREATED")
	for _, list := range lists {
		fmt.Fprintf(w, "%s\t%s\t%v\t%s\n",
			shortID(list.ID), list.Name, list.IsPaused, list.CreatedAt.Format(time.RFC3339))
	}
	w.Flush()
}

func (c *CLI) cmdListPause(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: list-pause <list-id>")
		os.Exit(1)
	}

	err := c.client.PauseList(args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("List paused successfully")
}

func (c *CLI) cmdListResume(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: list-resume <list-id>")
		os.Exit(1)
	}

	err := c.client.ResumeList(args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("List resumed successfully")
}

func (c *CLI) cmdSubscribe(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: subscribe <list-id> <email> [name]")
		os.Exit(1)
	}

	listID := args[0]
	email := args[1]
	name := ""
	if len(args) > 2 {
		name = strings.Join(args[2:], " ")
	}

	sub, err := c.client.Subscribe(listID, email, name, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printSubscriber(sub)
}

func (c *CLI) cmdUnsubscribe(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: unsubscribe <list-id> <email>")
		os.Exit(1)
	}

	err := c.client.Unsubscribe(args[0], args[1])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Unsubscribed successfully")
}

func (c *CLI) cmdSubscribers(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: subscribers <list-id>")
		os.Exit(1)
	}

	subs, err := c.client.ListSubscribers(args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "EMAIL\tNAME\tSTATUS\tSUBSCRIBED")
	for _, sub := range subs {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			sub.Email, sub.Name, sub.Status, sub.SubscribedAt.Format(time.RFC3339))
	}
	w.Flush()
}

func (c *CLI) cmdImport(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: import <list-id> <csv-file>")
		os.Exit(1)
	}

	total, imported, skipped, err := c.client.ImportSubscribers(args[0], args[1])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Import complete:\n")
	fmt.Printf("  Total: %d\n", total)
	fmt.Printf("  Imported: %d\n", imported)
	fmt.Printf("  Skipped: %d\n", skipped)
}

func (c *CLI) cmdTemplateCreate(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: template-create <name> <subject> [html] [text]")
		os.Exit(1)
	}

	name := args[0]
	subject := args[1]
	htmlBody := ""
	textBody := ""

	if len(args) > 2 {
		htmlBody = readFileOrDefault(args[2], "")
	}
	if len(args) > 3 {
		textBody = readFileOrDefault(args[3], "")
	}

	template, err := c.client.CreateTemplate(name, subject, htmlBody, textBody, true, true, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printTemplate(template)
}

func (c *CLI) cmdTemplateList(args []string) {
	templates, err := c.client.ListTemplates()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSUBJECT\tCREATED")
	for _, t := range templates {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			shortID(t.ID), t.Name, truncate(t.Subject, 30), t.CreatedAt.Format(time.RFC3339))
	}
	w.Flush()
}

func (c *CLI) cmdSend(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: send <list-id> [options]")
		os.Exit(1)
	}

	req := &protocol.SendCampaignRequest{
		ListID: args[0],
	}

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--template":
			if i+1 >= len(args) {
				fmt.Println("Error: --template requires an ID")
				os.Exit(1)
			}
			req.TemplateID = args[i+1]
			i++
		case "--subject":
			if i+1 >= len(args) {
				fmt.Println("Error: --subject requires text")
				os.Exit(1)
			}
			req.Subject = args[i+1]
			i++
		case "--html":
			if i+1 >= len(args) {
				fmt.Println("Error: --html requires a file path")
				os.Exit(1)
			}
			req.HTMLBody = readFileOrDefault(args[i+1], "")
			i++
		case "--text":
			if i+1 >= len(args) {
				fmt.Println("Error: --text requires a file path")
				os.Exit(1)
			}
			req.TextBody = readFileOrDefault(args[i+1], "")
			i++
		case "--schedule":
			if i+1 >= len(args) {
				fmt.Println("Error: --schedule requires RFC3339 time")
				os.Exit(1)
			}
			t, err := time.Parse(time.RFC3339, args[i+1])
			if err != nil {
				fmt.Printf("Error: invalid time format: %v\n", err)
				os.Exit(1)
			}
			req.ScheduledAt = &t
			i++
		case "--abtest":
			req.IsABTest = true
		case "--subjectA":
			if i+1 >= len(args) {
				fmt.Println("Error: --subjectA requires text")
				os.Exit(1)
			}
			req.SubjectA = args[i+1]
			i++
		case "--subjectB":
			if i+1 >= len(args) {
				fmt.Println("Error: --subjectB requires text")
				os.Exit(1)
			}
			req.SubjectB = args[i+1]
			i++
		}
	}

	task, err := c.client.SendCampaign(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printTask(task)
}

func (c *CLI) cmdTaskGet(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: task-get <task-id>")
		os.Exit(1)
	}

	task, err := c.client.GetTask(args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printTask(task)
}

func (c *CLI) cmdTaskList(args []string) {
	tasks, err := c.client.ListTasks()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tLIST\tSTATUS\tSENT\tFAILED\tBOUNCED")
	for _, t := range tasks {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%d\n",
			shortID(t.ID), shortID(t.ListID), t.Status, t.SentCount, t.FailedCount, t.BouncedCount)
	}
	w.Flush()
}

func (c *CLI) cmdRecords(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: records <task-id>")
		os.Exit(1)
	}

	records, err := c.client.GetSendRecords(args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "EMAIL\tSTATUS\tSENT_AT\tOPENED\tCLICKED")
	for _, r := range records {
		opened := "-"
		if r.OpenedAt != nil {
			opened = "yes"
		}
		clicked := "-"
		if r.ClickedAt != nil {
			clicked = "yes"
		}
		sentAt := "-"
		if r.SentAt != nil {
			sentAt = r.SentAt.Format(time.RFC3339)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			truncate(r.Email, 25), r.Status, sentAt, opened, clicked)
	}
	w.Flush()
}

func (c *CLI) cmdAlerts(args []string) {
	alerts, err := c.client.GetAlerts()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tLEVEL\tMESSAGE\tRESOLVED\tCREATED")
	for _, a := range alerts {
		fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%s\n",
			shortID(a.ID), a.Level, truncate(a.Message, 30), a.Resolved, a.CreatedAt.Format(time.RFC3339))
	}
	w.Flush()
}

func (c *CLI) cmdAlertResolve(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: alert-resolve <alert-id>")
		os.Exit(1)
	}

	err := c.client.ResolveAlert(args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Alert resolved successfully")
}

func (c *CLI) cmdStats(args []string) {
	stats, err := c.client.GetStats()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("System Statistics:")
	fmt.Printf("  Total Mailing Lists: %d\n", stats.Stats.TotalLists)
	fmt.Printf("  Total Subscribers:   %d\n", stats.Stats.TotalSubscribers)
	fmt.Printf("  Total Tasks:         %d\n", stats.Stats.TotalTasks)
	fmt.Printf("  Total Sent:          %d\n", stats.Stats.TotalSent)
	fmt.Printf("  Total Bounced:       %d\n", stats.Stats.TotalBounced)
	fmt.Printf("  Active Alerts:       %d\n", stats.Stats.ActiveAlerts)
}

func printMailingList(list *protocol.MailingList) {
	data, _ := json.MarshalIndent(list, "", "  ")
	fmt.Println(string(data))
}

func printSubscriber(sub *protocol.SubscriberListEntry) {
	data, _ := json.MarshalIndent(sub, "", "  ")
	fmt.Println(string(data))
}

func printTemplate(t *protocol.EmailTemplate) {
	data, _ := json.MarshalIndent(t, "", "  ")
	fmt.Println(string(data))
}

func printTask(task *protocol.SendTask) {
	data, _ := json.MarshalIndent(task, "", "  ")
	fmt.Println(string(data))
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func readFileOrDefault(path, defaultVal string) string {
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data)
		}
	}
	return defaultVal
}

func parseCustomFields(args []string, startIdx int) map[string]string {
	fields := make(map[string]string)
	for i := startIdx; i < len(args); i++ {
		parts := strings.SplitN(args[i], "=", 2)
		if len(parts) == 2 {
			fields[parts[0]] = parts[1]
		}
	}
	return fields
}

func parseIntArg(args []string, idx int, defaultVal int) int {
	if idx >= len(args) {
		return defaultVal
	}
	val, err := strconv.Atoi(args[idx])
	if err != nil {
		return defaultVal
	}
	return val
}
