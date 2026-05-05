package client

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"go-feedback-handler/pkg/protocol"
)

type CLI struct {
	apiClient *APIClient
}

func NewCLI(baseURL string) *CLI {
	return &CLI{
		apiClient: NewAPIClient(baseURL),
	}
}

func (c *CLI) Run(args []string) error {
	if len(args) < 1 {
		return c.printUsage()
	}

	cmd := strings.ToLower(args[0])
	switch cmd {
	case "create", "submit":
		return c.cmdCreate(args[1:])
	case "get", "show":
		return c.cmdGet(args[1:])
	case "list", "ls":
		return c.cmdList(args[1:])
	case "update-status", "status":
		return c.cmdUpdateStatus(args[1:])
	case "assign":
		return c.cmdAssign(args[1:])
	case "comment", "reply":
		return c.cmdComment(args[1:])
	case "tags":
		return c.cmdTags(args[1:])
	case "tag-add", "add-tag":
		return c.cmdAddTag(args[1:])
	case "tag-remove", "remove-tag":
		return c.cmdRemoveTag(args[1:])
	case "notifications", "notifs":
		return c.cmdNotifications(args[1:])
	case "user-limit":
		return c.cmdUserLimit(args[1:])
	case "kpi":
		return c.cmdKPI(args[1:])
	case "report":
		return c.cmdReport(args[1:])
	case "health":
		return c.cmdHealth()
	case "help", "-h", "--help":
		return c.printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		return c.printUsage()
	}
}

func (c *CLI) printUsage() error {
	fmt.Println(`Feedback Handler CLI

Usage:
  feedback-cli <command> [arguments]

Commands:
  create, submit   Submit a new feedback
  get, show        Get feedback details by ID
  list, ls         List feedbacks with filters
  update-status    Update feedback status
  assign           Assign feedback to a handler
  comment, reply   Add comment to feedback
  tags             List all tags
  tag-add          Add tag to feedback
  tag-remove       Remove tag from feedback
  notifications    List notifications for a user
  user-limit       Check user limit status
  kpi              Get KPI stats for a handler
  report           Get/generate monthly report
  health           Check server health
  help             Show this help message

Environment:
  FEEDBACK_SERVER  Server base URL (default: http://localhost:8080)

Examples:
  feedback-cli create --user u001 --name "张三" --type bug --content "登录失败"
  feedback-cli list --status pending --priority high
  feedback-cli get fb123abc
  feedback-cli update-status fb123abc processing --handler h001 --name "李四"
  feedback-cli comment fb123abc --user u001 --name "张三" --content "还有问题" --support=false
  feedback-cli tags
  feedback-cli kpi h001
  feedback-cli report --month 2026-05
`)
	return nil
}

func parseFlags(args []string) (map[string]string, []string) {
	flags := make(map[string]string)
	positional := make([]string, 0)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			key := strings.TrimPrefix(arg, "--")
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				flags[key] = args[i+1]
				i++
			} else {
				flags[key] = "true"
			}
		} else if strings.HasPrefix(arg, "-") {
			key := strings.TrimPrefix(arg, "-")
			if len(key) == 1 && i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				flags[key] = args[i+1]
				i++
			} else {
				for _, k := range key {
					flags[string(k)] = "true"
				}
			}
		} else {
			positional = append(positional, arg)
		}
	}
	return flags, positional
}

func (c *CLI) cmdCreate(args []string) error {
	flags, _ := parseFlags(args)

	userID := flags["user"]
	if userID == "" {
		userID = flags["user-id"]
	}
	userName := flags["name"]
	if userName == "" {
		userName = flags["user-name"]
	}
	ftype := flags["type"]
	content := flags["content"]

	if userID == "" || userName == "" || ftype == "" || content == "" {
		return fmt.Errorf("usage: feedback-cli create --user <user_id> --name <user_name> --type <bug|feature|complaint|consultation> --content <content>")
	}

	var feedbackType protocol.FeedbackType
	switch strings.ToLower(ftype) {
	case "bug":
		feedbackType = protocol.FeedbackTypeBug
	case "feature":
		feedbackType = protocol.FeedbackTypeFeature
	case "complaint":
		feedbackType = protocol.FeedbackTypeComplaint
	case "consultation":
		feedbackType = protocol.FeedbackTypeConsultation
	default:
		return fmt.Errorf("invalid type: %s (must be bug, feature, complaint, or consultation)", ftype)
	}

	resp, err := c.apiClient.CreateFeedback(userID, userName, feedbackType, content)
	if err != nil {
		return fmt.Errorf("failed to create feedback: %w", err)
	}

	fmt.Println("Feedback submitted successfully!")
	fmt.Printf("  ID: %s\n", resp.Feedback.ID)
	fmt.Printf("  Type: %s\n", resp.Feedback.Type)
	fmt.Printf("  Priority: %s\n", resp.Feedback.Priority)
	fmt.Printf("  Status: %s\n", resp.Feedback.Status)

	if resp.IsMerged {
		fmt.Printf("\n⚠️  This feedback was merged into: %s\n", resp.MergedInto)
	}
	if resp.InReview {
		fmt.Println("\n⚠️  This feedback is in review queue (user has too many invalid closes)")
	}

	return nil
}

func (c *CLI) cmdGet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: feedback-cli get <feedback_id>")
	}

	id := args[0]
	fb, err := c.apiClient.GetFeedback(id)
	if err != nil {
		return fmt.Errorf("failed to get feedback: %w", err)
	}

	c.printFeedbackDetail(fb)
	return nil
}

func (c *CLI) cmdList(args []string) error {
	flags, _ := parseFlags(args)

	req := &protocol.ListFeedbackRequest{
		Page:     protocol.DefaultPage,
		PageSize: protocol.DefaultPageSize,
	}

	if flags["status"] != "" {
		req.Status = protocol.FeedbackStatus(flags["status"])
	}
	if flags["type"] != "" {
		req.Type = protocol.FeedbackType(flags["type"])
	}
	if flags["priority"] != "" {
		req.Priority = protocol.FeedbackPriority(flags["priority"])
	}
	if flags["user"] != "" {
		req.UserID = flags["user"]
	}
	if flags["handler"] != "" {
		req.HandlerID = flags["handler"]
	}
	if flags["tag"] != "" {
		req.TagID = flags["tag"]
	}
	if flags["in-review"] != "" {
		val := flags["in-review"] == "true" || flags["in-review"] == "1"
		req.InReview = &val
	}
	if flags["page"] != "" {
		p, _ := strconv.Atoi(flags["page"])
		req.Page = p
	}
	if flags["page-size"] != "" {
		ps, _ := strconv.Atoi(flags["page-size"])
		req.PageSize = ps
	}

	resp, err := c.apiClient.ListFeedbacks(req)
	if err != nil {
		return fmt.Errorf("failed to list feedbacks: %w", err)
	}

	c.printFeedbackList(resp)
	return nil
}

func (c *CLI) cmdUpdateStatus(args []string) error {
	flags, positional := parseFlags(args)

	var id string
	if len(positional) > 0 {
		id = positional[0]
	}
	if id == "" {
		id = flags["id"]
	}

	status := flags["status"]
	if status == "" && len(positional) > 1 {
		status = positional[1]
	}

	handlerID := flags["handler"]
	handlerName := flags["name"]
	isInvalid := flags["invalid"] == "true" || flags["invalid"] == "1"

	if id == "" || status == "" {
		return fmt.Errorf("usage: feedback-cli update-status <id> <status> [--handler <handler_id>] [--name <handler_name>] [--invalid=true]")
	}

	var feedbackStatus protocol.FeedbackStatus
	switch strings.ToLower(status) {
	case "pending":
		feedbackStatus = protocol.StatusPending
	case "processing":
		feedbackStatus = protocol.StatusProcessing
	case "resolved":
		feedbackStatus = protocol.StatusResolved
	case "closed":
		feedbackStatus = protocol.StatusClosed
	default:
		return fmt.Errorf("invalid status: %s (must be pending, processing, resolved, or closed)", status)
	}

	err := c.apiClient.UpdateStatus(id, feedbackStatus, handlerID, handlerName, isInvalid)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	fmt.Printf("Feedback %s status updated to %s\n", id, status)
	return nil
}

func (c *CLI) cmdAssign(args []string) error {
	flags, positional := parseFlags(args)

	var id string
	if len(positional) > 0 {
		id = positional[0]
	}
	if id == "" {
		id = flags["id"]
	}

	handlerID := flags["handler"]
	if handlerID == "" && len(positional) > 1 {
		handlerID = positional[1]
	}
	handlerName := flags["name"]

	if id == "" || handlerID == "" {
		return fmt.Errorf("usage: feedback-cli assign <id> <handler_id> [--name <handler_name>]")
	}

	err := c.apiClient.AssignHandler(id, handlerID, handlerName)
	if err != nil {
		return fmt.Errorf("failed to assign handler: %w", err)
	}

	fmt.Printf("Feedback %s assigned to %s\n", id, handlerID)
	return nil
}

func (c *CLI) cmdComment(args []string) error {
	flags, positional := parseFlags(args)

	var id string
	if len(positional) > 0 {
		id = positional[0]
	}
	if id == "" {
		id = flags["id"]
	}

	userID := flags["user"]
	userName := flags["name"]
	content := flags["content"]
	isFromSupport := flags["support"] == "true" || flags["support"] == "1"

	if id == "" || userID == "" || content == "" {
		return fmt.Errorf("usage: feedback-cli comment <id> --user <user_id> [--name <user_name>] --content <content> [--support=true]")
	}

	err := c.apiClient.AddComment(id, userID, userName, content, isFromSupport)
	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}

	fmt.Println("Comment added successfully")
	return nil
}

func (c *CLI) cmdTags(args []string) error {
	tags, err := c.apiClient.ListTags()
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	fmt.Println("Tags:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tColor\tSystem")
	for _, tag := range tags {
		systemMark := ""
		if tag.IsSystem {
			systemMark = "✓"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", tag.ID[:8], tag.Name, tag.Color, systemMark)
	}
	w.Flush()

	return nil
}

func (c *CLI) cmdAddTag(args []string) error {
	flags, positional := parseFlags(args)

	var fbID, tagID string
	if len(positional) >= 2 {
		fbID = positional[0]
		tagID = positional[1]
	}
	if fbID == "" {
		fbID = flags["feedback"]
	}
	if tagID == "" {
		tagID = flags["tag"]
	}

	if fbID == "" || tagID == "" {
		return fmt.Errorf("usage: feedback-cli tag-add <feedback_id> <tag_id>")
	}

	err := c.apiClient.AddTagToFeedback(fbID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add tag: %w", err)
	}

	fmt.Printf("Tag %s added to feedback %s\n", tagID, fbID)
	return nil
}

func (c *CLI) cmdRemoveTag(args []string) error {
	flags, positional := parseFlags(args)

	var fbID, tagID string
	if len(positional) >= 2 {
		fbID = positional[0]
		tagID = positional[1]
	}
	if fbID == "" {
		fbID = flags["feedback"]
	}
	if tagID == "" {
		tagID = flags["tag"]
	}

	if fbID == "" || tagID == "" {
		return fmt.Errorf("usage: feedback-cli tag-remove <feedback_id> <tag_id>")
	}

	err := c.apiClient.RemoveTagFromFeedback(fbID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag: %w", err)
	}

	fmt.Printf("Tag %s removed from feedback %s\n", tagID, fbID)
	return nil
}

func (c *CLI) cmdNotifications(args []string) error {
	flags, _ := parseFlags(args)

	userID := flags["user"]
	if userID == "" && len(args) > 1 {
		userID = args[1]
	}

	if userID == "" {
		return fmt.Errorf("usage: feedback-cli notifications --user <user_id>")
	}

	notifs, err := c.apiClient.GetNotifications(userID)
	if err != nil {
		return fmt.Errorf("failed to get notifications: %w", err)
	}

	if len(notifs) == 0 {
		fmt.Println("No notifications")
		return nil
	}

	fmt.Println("Notifications:")
	for i, notif := range notifs {
		readMark := ""
		if !notif.Read {
			readMark = " [UNREAD]"
		}
		fmt.Printf("\n[%d] %s%s\n", i+1, notif.Type, readMark)
		fmt.Printf("  Message: %s\n", notif.Message)
		fmt.Printf("  Created: %s\n", notif.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func (c *CLI) cmdUserLimit(args []string) error {
	flags, _ := parseFlags(args)

	userID := flags["user"]
	if userID == "" && len(args) > 1 {
		userID = args[1]
	}

	if userID == "" {
		return fmt.Errorf("usage: feedback-cli user-limit --user <user_id>")
	}

	limit, err := c.apiClient.GetUserLimit(userID)
	if err != nil {
		return fmt.Errorf("failed to get user limit: %w", err)
	}

	fmt.Println("User Limit Info:")
	fmt.Printf("  User ID: %s\n", limit.UserID)
	fmt.Printf("  Invalid Close Count: %d/%d\n", limit.InvalidCloseCount, protocol.MaxInvalidCloses)
	fmt.Printf("  Under Review: %v\n", limit.IsUnderReview)
	if limit.IsUnderReview {
		fmt.Printf("  Review Started At: %s\n", limit.ReviewStartedAt.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func (c *CLI) cmdKPI(args []string) error {
	flags, positional := parseFlags(args)

	var handlerID string
	if len(positional) > 0 {
		handlerID = positional[0]
	}
	if handlerID == "" {
		handlerID = flags["handler"]
	}

	period := flags["period"]
	if period == "" {
		period = "month"
	}

	if handlerID == "" {
		return fmt.Errorf("usage: feedback-cli kpi <handler_id> [--period <period>]")
	}

	stats, err := c.apiClient.GetKPIStats(handlerID, period)
	if err != nil {
		return fmt.Errorf("failed to get KPI stats: %w", err)
	}

	fmt.Println("KPI Statistics:")
	fmt.Printf("  Handler ID: %s\n", stats.HandlerID)
	fmt.Printf("  Period: %s\n", stats.Period)
	fmt.Printf("\n  Total Handled: %d\n", stats.TotalHandled)
	fmt.Printf("  Resolved: %d\n", stats.ResolvedCount)
	fmt.Printf("  Closed: %d\n", stats.ClosedCount)
	fmt.Printf("  Escalated: %d\n", stats.EscalatedCount)
	fmt.Printf("\n  Avg First Response Time: %.2f hours\n", stats.AvgFirstResponseTime)
	fmt.Printf("  Avg Resolution Time: %.2f hours\n", stats.AvgResolutionTime)
	fmt.Printf("  SLA Compliance Rate: %.2f%%\n", stats.SLAComplianceRate)

	return nil
}

func (c *CLI) cmdReport(args []string) error {
	flags, _ := parseFlags(args)

	month := flags["month"]
	generate := flags["generate"] == "true"

	if generate {
		report, err := c.apiClient.GenerateMonthlyReport(month)
		if err != nil {
			return fmt.Errorf("failed to generate report: %w", err)
		}
		c.printReport(report)
	} else {
		report, err := c.apiClient.GetMonthlyReport(month)
		if err != nil {
			return fmt.Errorf("failed to get report: %w", err)
		}
		c.printReport(report)
	}

	return nil
}

func (c *CLI) cmdHealth() error {
	err := c.apiClient.HealthCheck()
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	fmt.Println("✓ Server is healthy")
	return nil
}

func (c *CLI) printFeedbackList(resp *protocol.ListFeedbackResponse) {
	if len(resp.Feedbacks) == 0 {
		fmt.Println("No feedbacks found")
		return
	}

	fmt.Printf("Showing %d of %d feedbacks (Page %d/%d)\n\n",
		len(resp.Feedbacks), resp.Total, resp.Page,
		(resp.Total+resp.PageSize-1)/resp.PageSize)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tType\tPriority\tStatus\tHandler\tCreated")
	for _, fb := range resp.Feedbacks {
		handler := "-"
		if fb.HandlerID != "" {
			handler = fb.HandlerID
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			fb.ID[:8], fb.Type, fb.Priority, fb.Status, handler,
			fb.CreatedAt.Format("2006-01-02 15:04"))
	}
	w.Flush()
}

func (c *CLI) printFeedbackDetail(fb *protocol.Feedback) {
	data, _ := json.MarshalIndent(fb, "", "  ")
	fmt.Println(string(data))
}

func (c *CLI) printReport(report *protocol.MonthlyReport) {
	fmt.Println("==================================================")
	fmt.Printf("           MONTHLY REPORT - %s\n", report.Month)
	fmt.Println("==================================================")

	fmt.Printf("\n## Summary\n")
	fmt.Printf("  Total Feedbacks: %d\n", report.TotalFeedbacks)
	fmt.Printf("  Merged: %d\n", report.MergedCount)
	fmt.Printf("  Escalated: %d\n", report.EscalatedCount)
	fmt.Printf("  Invalid: %d\n", report.InvalidCount)

	fmt.Printf("\n## By Type\n")
	for t, count := range report.ByType {
		fmt.Printf("  %s: %d\n", t, count)
	}

	fmt.Printf("\n## By Status\n")
	for s, count := range report.ByStatus {
		fmt.Printf("  %s: %d\n", s, count)
	}

	fmt.Printf("\n## Performance\n")
	fmt.Printf("  Avg Resolution Time: %.2f hours\n", report.AvgResolutionTime)
	fmt.Printf("  Avg First Response Time: %.2f hours\n", report.AvgFirstResponseTime)
	fmt.Printf("  SLA Compliance Rate: %.2f%%\n", report.SLAComplianceRate)

	if len(report.TopTags) > 0 {
		fmt.Printf("\n## Top Tags\n")
		for _, tag := range report.TopTags {
			fmt.Printf("  %s: %d\n", tag.TagName, tag.Count)
		}
	}

	fmt.Printf("\n## Trend Analysis\n")
	fmt.Printf("  %s\n", report.TrendAnalysis)

	if len(report.ImprovementSuggestions) > 0 {
		fmt.Printf("\n## Improvement Suggestions\n")
		for i, s := range report.ImprovementSuggestions {
			fmt.Printf("  %d. %s\n", i+1, s)
		}
	}

	fmt.Printf("\n## Report Generated At\n")
	fmt.Printf("  %s\n", report.CreatedAt.Format("2006-01-02 15:04:05"))
}
