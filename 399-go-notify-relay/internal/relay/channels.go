package relay

import (
	"fmt"
	"unicode/utf8"
)

type Channel interface {
	Name() string
	Send(alert *Alert) error
}

type EmailChannel struct{}

func (c *EmailChannel) Name() string {
	return "email"
}

func (c *EmailChannel) Send(alert *Alert) error {
	title := fmt.Sprintf("告警: %s", alert.Level)
	fmt.Println("[EMAIL] ====================================")
	fmt.Printf("[EMAIL] 标题: %s\n", title)
	fmt.Println("[EMAIL] ------------------------------------")
	fmt.Printf("[EMAIL] 正文: %s\n", alert.Message)
	fmt.Println("[EMAIL] ====================================")
	return nil
}

type SMSChannel struct{}

func (c *SMSChannel) Name() string {
	return "sms"
}

func (c *SMSChannel) Send(alert *Alert) error {
	content := alert.Message
	maxLength := 70
	charCount := utf8.RuneCountInString(content)
	
	if charCount > maxLength {
		runes := []rune(content)
		content = string(runes[:maxLength]) + "..."
	}
	fmt.Println("[SMS] ====================================")
	fmt.Printf("[SMS] 内容: %s\n", content)
	fmt.Println("[SMS] ====================================")
	return nil
}

type DingtalkChannel struct{}

func (c *DingtalkChannel) Name() string {
	return "dingtalk"
}

func (c *DingtalkChannel) Send(alert *Alert) error {
	emoji := GetLevelEmoji(alert.Level)
	markdown := fmt.Sprintf("### %s 告警通知\n\n**级别**: %s %s\n\n**消息**: %s", emoji, alert.Level, emoji, alert.Message)
	fmt.Println("[DINGTALK] ====================================")
	fmt.Printf("[DINGTALK] Markdown消息:\n")
	fmt.Println(markdown)
	fmt.Println("[DINGTALK] ====================================")
	return nil
}

func GetAllChannels() []Channel {
	return []Channel{
		&EmailChannel{},
		&SMSChannel{},
		&DingtalkChannel{},
	}
}
