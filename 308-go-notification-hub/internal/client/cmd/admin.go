package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"notification-hub/internal/client/api"
	"notification-hub/pkg/models"
)

type AdminCommand struct {
	client *api.Client
	reader *bufio.Reader
}

func NewAdminCommand(client *api.Client) *AdminCommand {
	return &AdminCommand{
		client: client,
		reader: bufio.NewReader(os.Stdin),
	}
}

func (c *AdminCommand) Run() error {
	for {
		fmt.Println("\n=== 管理员菜单 ===")
		fmt.Println("1. 创建通知")
		fmt.Println("2. 查看统计")
		fmt.Println("0. 返回主菜单")
		fmt.Print("请选择: ")

		choice, _ := c.reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			c.createNotification()
		case "2":
			c.viewStatistics()
		case "0":
			return nil
		default:
			fmt.Println("无效选择，请重试")
		}
	}
}

func (c *AdminCommand) createNotification() {
	fmt.Println("\n--- 创建通知 ---")

	fmt.Print("请输入通知标题: ")
	title, _ := c.reader.ReadString('\n')
	title = strings.TrimSpace(title)

	fmt.Print("请输入通知内容: ")
	content, _ := c.reader.ReadString('\n')
	content = strings.TrimSpace(content)

	channels := c.selectChannels()
	if len(channels) == 0 {
		fmt.Println("至少需要选择一个渠道")
		return
	}

	targetAll := c.selectTargetType()
	var targetUsers []string

	if !targetAll {
		targetUsers = c.selectTargetUsers()
		if len(targetUsers) == 0 {
			fmt.Println("未选择目标用户")
			return
		}
	}

	req := &models.CreateNotificationRequest{
		Title:       title,
		Content:     content,
		Channels:    channels,
		TargetAll:   targetAll,
		TargetUsers: targetUsers,
	}

	resp, err := c.client.CreateNotification(req)
	if err != nil {
		fmt.Printf("创建通知失败: %v\n", err)
		return
	}

	fmt.Printf("\n通知创建成功!\n")
	fmt.Printf("通知ID: %s\n", resp.NotificationID)
	fmt.Printf("目标用户数: %d\n", resp.TotalUsers)
}

func (c *AdminCommand) selectChannels() []models.Channel {
	fmt.Println("\n请选择发送渠道（可多选，用逗号分隔）:")
	fmt.Println("1. 站内信 (insite)")
	fmt.Println("2. 邮件 (email)")
	fmt.Println("3. 短信 (sms)")
	fmt.Print("选择: ")

	input, _ := c.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	selections := strings.Split(input, ",")
	channelMap := map[string]models.Channel{
		"1": models.ChannelInSite,
		"2": models.ChannelEmail,
		"3": models.ChannelSMS,
	}

	var channels []models.Channel
	selected := make(map[models.Channel]bool)

	for _, sel := range selections {
		sel = strings.TrimSpace(sel)
		if ch, ok := channelMap[sel]; ok && !selected[ch] {
			channels = append(channels, ch)
			selected[ch] = true
		}
	}

	return channels
}

func (c *AdminCommand) selectTargetType() bool {
	fmt.Println("\n请选择目标用户类型:")
	fmt.Println("1. 发送给全部用户")
	fmt.Println("2. 发送给指定用户")
	fmt.Print("选择: ")

	choice, _ := c.reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	return choice == "1"
}

func (c *AdminCommand) selectTargetUsers() []string {
	fmt.Println("\n可用用户: user1, user2, user3, user4, user5")
	fmt.Print("请输入目标用户（用逗号分隔）: ")

	input, _ := c.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	users := strings.Split(input, ",")
	var result []string
	selected := make(map[string]bool)

	for _, u := range users {
		u = strings.TrimSpace(u)
		if u != "" && !selected[u] {
			result = append(result, u)
			selected[u] = true
		}
	}

	return result
}

func (c *AdminCommand) viewStatistics() {
	fmt.Println("\n--- 通知统计 ---")

	resp, err := c.client.GetStatistics()
	if err != nil {
		fmt.Printf("获取统计失败: %v\n", err)
		return
	}

	if len(resp.Statistics) == 0 {
		fmt.Println("暂无通知数据")
		return
	}

	for _, stat := range resp.Statistics {
		fmt.Printf("\n通知ID: %s\n", stat.NotificationID)
		fmt.Printf("标题: %s\n", stat.Title)
		fmt.Printf("目标用户总数: %d\n", stat.TotalUsers)
		fmt.Println("渠道统计:")

		for channel, chStat := range stat.ChannelStats {
			fmt.Printf("  [%s]: 总数=%d, 已送达=%d, 失败=%d, 待处理=%d\n",
				channel, chStat.Total, chStat.Delivered, chStat.Failed, chStat.Pending)
		}
		fmt.Println("---")
	}
}
