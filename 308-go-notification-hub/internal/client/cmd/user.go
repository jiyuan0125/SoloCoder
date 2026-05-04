package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"notification-hub/internal/client/api"
	"notification-hub/pkg/models"
)

type UserCommand struct {
	client *api.Client
	reader *bufio.Reader
	userID string
}

func NewUserCommand(client *api.Client) *UserCommand {
	return &UserCommand{
		client: client,
		reader: bufio.NewReader(os.Stdin),
	}
}

func (c *UserCommand) Run() error {
	if !c.login() {
		return nil
	}

	for {
		fmt.Printf("\n=== 用户菜单 (当前用户: %s) ===\n", c.userID)
		fmt.Println("1. 查看所有通知")
		fmt.Println("2. 查看未读通知")
		fmt.Println("3. 查看已读通知")
		fmt.Println("4. 标记通知为已读")
		fmt.Println("0. 返回主菜单")
		fmt.Print("请选择: ")

		choice, _ := c.reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			c.listNotifications(nil)
		case "2":
			falseVal := false
			c.listNotifications(&falseVal)
		case "3":
			trueVal := true
			c.listNotifications(&trueVal)
		case "4":
			c.markAsRead()
		case "0":
			return nil
		default:
			fmt.Println("无效选择，请重试")
		}
	}
}

func (c *UserCommand) login() bool {
	fmt.Println("\n--- 用户登录 ---")
	fmt.Print("请输入用户ID (可用: user1, user2, user3, user4, user5): ")

	userID, _ := c.reader.ReadString('\n')
	userID = strings.TrimSpace(userID)

	if userID == "" {
		fmt.Println("用户ID不能为空")
		return false
	}

	c.userID = userID
	return true
}

func (c *UserCommand) listNotifications(isRead *bool) {
	var filter string
	if isRead == nil {
		filter = "所有"
	} else if *isRead {
		filter = "已读"
	} else {
		filter = "未读"
	}

	fmt.Printf("\n--- %s通知列表 ---\n", filter)

	resp, err := c.client.GetUserNotifications(c.userID, isRead)
	if err != nil {
		fmt.Printf("获取通知失败: %v\n", err)
		return
	}

	if len(resp.Notifications) == 0 {
		fmt.Println("暂无通知")
		return
	}

	for i, n := range resp.Notifications {
		fmt.Printf("\n[%d] 送达ID: %s\n", i+1, n.ID)
		fmt.Printf("    通知标题: %s\n", n.Title)
		fmt.Printf("    通知内容: %s\n", n.Content)
		fmt.Printf("    渠道: %s\n", n.Channel)
		fmt.Printf("    状态: %s\n", n.Status)
		fmt.Printf("    是否已读: %v\n", n.IsRead)
		fmt.Printf("    创建时间: %s\n", n.CreatedAt.Format(time.DateTime))
		fmt.Println("    ---")
	}
}

func (c *UserCommand) markAsRead() {
	fmt.Println("\n--- 标记通知为已读 ---")

	resp, err := c.client.GetUserNotifications(c.userID, nil)
	if err != nil {
		fmt.Printf("获取通知失败: %v\n", err)
		return
	}

	var insiteNotifications []models.UserNotification
	for _, n := range resp.Notifications {
		if n.Channel == models.ChannelInSite {
			insiteNotifications = append(insiteNotifications, n)
		}
	}

	if len(insiteNotifications) == 0 {
		fmt.Println("没有可标记已读的站内信通知")
		return
	}

	fmt.Println("您的站内信通知:")
	for i, n := range insiteNotifications {
		readStatus := "未读"
		if n.IsRead {
			readStatus = "已读"
		}
		fmt.Printf("[%d] 送达ID: %s | 标题: %s | 状态: %s\n", i+1, n.ID, n.Title, readStatus)
	}

	fmt.Print("\n请输入要标记为已读的送达ID: ")
	deliveryID, _ := c.reader.ReadString('\n')
	deliveryID = strings.TrimSpace(deliveryID)

	if deliveryID == "" {
		fmt.Println("送达ID不能为空")
		return
	}

	result, err := c.client.MarkAsRead(c.userID, deliveryID)
	if err != nil {
		fmt.Printf("标记已读失败: %v\n", err)
		return
	}

	if result.Success {
		fmt.Println("标记已读成功!")
	} else {
		fmt.Println("标记已读失败")
	}
}
