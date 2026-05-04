package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"notification-hub/internal/client/api"
	"notification-hub/internal/client/cmd"
)

func main() {
	serverURL := getEnv("SERVER_URL", "http://localhost:8080")

	client := api.NewClient(serverURL)
	adminCmd := cmd.NewAdminCommand(client)
	userCmd := cmd.NewUserCommand(client)

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== 多渠道通知中心客户端 ===")
	fmt.Printf("服务端地址: %s\n", serverURL)

	for {
		fmt.Println("\n=== 主菜单 ===")
		fmt.Println("1. 管理员模式")
		fmt.Println("2. 用户模式")
		fmt.Println("0. 退出")
		fmt.Print("请选择: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			if err := adminCmd.Run(); err != nil {
				fmt.Printf("管理员命令执行失败: %v\n", err)
			}
		case "2":
			if err := userCmd.Run(); err != nil {
				fmt.Printf("用户命令执行失败: %v\n", err)
			}
		case "0":
			fmt.Println("再见!")
			return
		default:
			fmt.Println("无效选择，请重试")
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
