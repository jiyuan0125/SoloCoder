package main

import (
	"fmt"
	"notification-center/common"
	"strings"
	"text/tabwriter"
	"os"
)

type Printer struct {
	writer *tabwriter.Writer
}

func NewPrinter() *Printer {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	return &Printer{writer: w}
}

func (p *Printer) Flush() {
	p.writer.Flush()
}

func (p *Printer) PrintSuccess(message string) {
	fmt.Printf("[SUCCESS] %s\n", message)
}

func (p *Printer) PrintError(err error) {
	fmt.Printf("[ERROR] %v\n", err)
}

func (p *Printer) PrintNotification(notif common.Notification) {
	fmt.Println("==================================================")
	fmt.Printf("ID:         %s\n", notif.ID)
	fmt.Printf("Type:       %s\n", notif.Type)
	fmt.Printf("Priority:   %s", notif.Priority)
	if notif.Priority == common.PriorityHigh {
		fmt.Printf(" (HIGH - TOP DISPLAY)")
	}
	fmt.Println()
	fmt.Printf("Sender:     %s\n", notif.Sender)
	fmt.Printf("Receiver:   %s\n", notif.Receiver)
	fmt.Printf("Title:      %s\n", notif.Title)
	fmt.Printf("Status:     %s\n", notif.Status)
	fmt.Printf("Created:    %s\n", notif.CreatedAt.Format("2006-01-02 15:04:05"))
	if notif.ReadAt != nil {
		fmt.Printf("Read At:    %s\n", notif.ReadAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("Archived:   %v\n", notif.IsArchived)
	fmt.Printf("Emergency:  %v\n", notif.IsEmergency)
	fmt.Println("--------------------------------------------------")
	fmt.Println("Content:")
	fmt.Println(notif.Content)
	fmt.Println("==================================================")
}

func (p *Printer) PrintNotificationList(resp *common.ListResponse) {
	fmt.Printf("Total: %d, Unread: %d\n\n", resp.Total, resp.UnreadCount)
	
	if len(resp.Notifications) == 0 {
		fmt.Println("No notifications found.")
		return
	}
	
	fmt.Fprintln(p.writer, "ID\tTYPE\tPRIORITY\tSTATUS\tTITLE\tCREATED")
	fmt.Fprintln(p.writer, "--\t----\t--------\t------\t-----\t-------")
	
	for _, n := range resp.Notifications {
		priority := string(n.Priority)
		if n.Priority == common.PriorityHigh {
			priority = "★" + priority
		}
		title := truncate(n.Title, 30)
		created := n.CreatedAt.Format("01-02 15:04")
		fmt.Fprintf(p.writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			truncate(n.ID, 8), n.Type, priority, n.Status, title, created)
	}
	p.writer.Flush()
}

func (p *Printer) PrintTemplate(template common.NotificationTemplate) {
	fmt.Println("==================================================")
	fmt.Printf("ID:        %s\n", template.ID)
	fmt.Printf("Name:      %s\n", template.Name)
	fmt.Printf("Type:      %s\n", template.Type)
	fmt.Printf("Title:     %s\n", template.Title)
	fmt.Printf("Variables: %v\n", template.Variables)
	fmt.Printf("Created:   %s\n", template.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated:   %s\n", template.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println("--------------------------------------------------")
	fmt.Println("Content Template:")
	fmt.Println(template.Content)
	fmt.Println("==================================================")
}

func (p *Printer) PrintTemplateList(templates []common.NotificationTemplate) {
	if len(templates) == 0 {
		fmt.Println("No templates found.")
		return
	}
	
	fmt.Fprintln(p.writer, "ID\tNAME\tTYPE\tTITLE\tVARIABLES")
	fmt.Fprintln(p.writer, "--\t----\t----\t-----\t---------")
	
	for _, t := range templates {
		vars := strings.Join(t.Variables, ",")
		if vars == "" {
			vars = "-"
		}
		fmt.Fprintf(p.writer, "%s\t%s\t%s\t%s\t%s\n",
			truncate(t.ID, 8), truncate(t.Name, 15), t.Type, 
			truncate(t.Title, 20), truncate(vars, 20))
	}
	p.writer.Flush()
}

func (p *Printer) PrintUnreadCount(resp *common.UnreadCountResponse) {
	fmt.Printf("Receiver: %s\n", resp.Receiver)
	fmt.Printf("Unread Count: %d\n", resp.UnreadCount)
}

func (p *Printer) PrintMarkReadResult(resp *common.MarkReadResponse) {
	fmt.Printf("Success: %v\n", resp.Success)
	fmt.Printf("Marked as Read: %d\n", resp.MarkedCount)
}

func (p *Printer) PrintSendResult(resp *common.SendResponse) {
	fmt.Printf("Success: %v\n", resp.Success)
	if resp.NotificationID != "" {
		fmt.Printf("Notification ID: %s\n", resp.NotificationID)
	}
}

func (p *Printer) PrintFailedLog(log common.FailedLog) {
	fmt.Println("==================================================")
	fmt.Printf("Log ID:           %s\n", log.ID)
	fmt.Printf("Notification ID:  %s\n", log.NotificationID)
	fmt.Printf("Receiver:         %s\n", log.Receiver)
	fmt.Printf("Title:            %s\n", log.Title)
	fmt.Printf("Retry Count:      %d\n", log.RetryCount)
	fmt.Printf("Failed At:        %s\n", log.FailedAt.Format("2006-01-02 15:04:05"))
	fmt.Println("--------------------------------------------------")
	fmt.Println("Error Messages:")
	for i, msg := range log.ErrorMessages {
		fmt.Printf("  %d. %s\n", i+1, msg)
	}
	fmt.Println("==================================================")
}

func (p *Printer) PrintFailedLogList(logs []common.FailedLog) {
	if len(logs) == 0 {
		fmt.Println("No failed logs found.")
		return
	}
	
	fmt.Fprintln(p.writer, "ID\tRECEIVER\tTITLE\tRETRIES\tFAILED AT")
	fmt.Fprintln(p.writer, "--\t--------\t-----\t-------\t---------")
	
	for _, l := range logs {
		fmt.Fprintf(p.writer, "%s\t%s\t%s\t%d\t%s\n",
			truncate(l.ID, 8), truncate(l.Receiver, 12), 
			truncate(l.Title, 20), l.RetryCount,
			l.FailedAt.Format("01-02 15:04"))
	}
	p.writer.Flush()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
