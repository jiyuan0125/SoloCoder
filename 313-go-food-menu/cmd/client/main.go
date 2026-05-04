package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/example/food-menu/client"
)

const (
	defaultServerURL = "http://localhost:8080"
)

func main() {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	apiClient := client.NewAPIClient(serverURL)

	fmt.Println("========================================")
	fmt.Println("       餐厅菜品管理系统")
	fmt.Println("========================================")
	fmt.Printf("服务端地址: %s\n", serverURL)
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("请选择模式：")
		fmt.Println("1. 管理员模式")
		fmt.Println("2. 顾客模式")
		fmt.Println("3. 退出")
		fmt.Print("请输入选择 (1-3): ")

		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(input)

		switch choice {
		case "1":
			adminCLI := client.NewAdminCLI(apiClient)
			adminCLI.Run()
		case "2":
			customerCLI := client.NewCustomerCLI(apiClient)
			customerCLI.Run()
		case "3":
			fmt.Println("感谢使用，再见！")
			return
		default:
			fmt.Println("无效的选择，请重新输入")
		}

		fmt.Println()
	}
}
