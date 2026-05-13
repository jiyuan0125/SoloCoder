package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"monitor-alert/internal/types"
)

type Notifier interface {
	Name() string
	Send(ctx context.Context, alert types.Alert, status types.AlertStatus) error
}

type StdoutNotifier struct {
	name string
}

func NewStdoutNotifier(name string) *StdoutNotifier {
	return &StdoutNotifier{name: name}
}

func (n *StdoutNotifier) Name() string {
	return n.name
}

func (n *StdoutNotifier) Send(ctx context.Context, alert types.Alert, status types.AlertStatus) error {
	statusText := "触发"
	if status == types.AlertStatusResolved {
		statusText = "恢复"
	}
	
	msg := fmt.Sprintf("[%s][%s] 告警 %s: %s - %s\n",
		time.Now().Format(time.RFC3339),
		alert.Severity,
		statusText,
		alert.RuleName,
		alert.Message,
	)
	
	_, err := fmt.Fprint(os.Stdout, msg)
	return err
}

type EmailNotifier struct {
	name   string
	config types.EmailConfig
}

func NewEmailNotifier(name string, config types.EmailConfig) *EmailNotifier {
	return &EmailNotifier{name: name, config: config}
}

func (n *EmailNotifier) Name() string {
	return n.name
}

func (n *EmailNotifier) Send(ctx context.Context, alert types.Alert, status types.AlertStatus) error {
	statusText := "Firing"
	if status == types.AlertStatusResolved {
		statusText = "Resolved"
	}
	
	subject := fmt.Sprintf("[%s] %s: %s", alert.Severity, statusText, alert.RuleName)
	
	body := fmt.Sprintf(`告警规则: %s
严重级别: %s
状态: %s
消息: %s
时间: %s
`, alert.RuleName, alert.Severity, statusText, alert.Message, alert.StartedAt.Format(time.RFC3339))
	
	if alert.ResolvedAt != nil {
		body += fmt.Sprintf("恢复时间: %s\n", alert.ResolvedAt.Format(time.RFC3339))
	}
	
	auth := smtp.PlainAuth("", n.config.Username, n.config.Password, n.config.SMTPHost)
	
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		n.config.From,
		n.config.To,
		subject,
		body,
	)
	
	addr := fmt.Sprintf("%s:%d", n.config.SMTPHost, n.config.SMTPPort)
	
	return smtp.SendMail(addr, auth, n.config.From, n.config.To, []byte(msg))
}

type WebhookNotifier struct {
	name   string
	config types.WebhookConfig
}

func NewWebhookNotifier(name string, config types.WebhookConfig) *WebhookNotifier {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.Method == "" {
		config.Method = "POST"
	}
	
	return &WebhookNotifier{name: name, config: config}
}

func (n *WebhookNotifier) Name() string {
	return n.name
}

func (n *WebhookNotifier) Send(ctx context.Context, alert types.Alert, status types.AlertStatus) error {
	payload := map[string]interface{}{
		"rule_name":  alert.RuleName,
		"severity":   alert.Severity,
		"status":     status,
		"message":    alert.Message,
		"started_at": alert.StartedAt.Format(time.RFC3339),
	}
	
	if alert.ResolvedAt != nil {
		payload["resolved_at"] = alert.ResolvedAt.Format(time.RFC3339)
	}
	
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	ctx, cancel := context.WithTimeout(ctx, n.config.Timeout)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, n.config.Method, n.config.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	
	for k, v := range n.config.Headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook 返回状态码 %d: %s", resp.StatusCode, string(respBody))
	}
	
	return nil
}

type NotificationManager struct {
	notifiers map[string]Notifier
}

func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		notifiers: make(map[string]Notifier),
	}
}

func (m *NotificationManager) Register(name string, notifier Notifier) {
	m.notifiers[name] = notifier
}

func (m *NotificationManager) Send(ctx context.Context, alert types.Alert, status types.AlertStatus, channels []string) {
	if len(channels) == 0 {
		return
	}
	
	for _, ch := range channels {
		notifier, exists := m.notifiers[ch]
		if !exists {
			fmt.Fprintf(os.Stderr, "ERROR: 通知渠道不存在: %s\n", ch)
			continue
		}
		
		if err := notifier.Send(ctx, alert, status); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: 通过渠道 '%s' 发送通知失败: %v\n", ch, err)
			log.Printf("WARN: 通知发送失败，继续运行")
		}
	}
}
