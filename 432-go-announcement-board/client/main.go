package main

import (
	"announcement-board/common"
	"fmt"
	"os"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("       公告管理系统 - 命令行客户端")
	fmt.Println("==============================================")
	fmt.Println("服务地址: http://localhost:8080")
	fmt.Println("")

	loginUser()

	user, err := GetUserInfo()
	if err != nil {
		fmt.Printf("连接服务失败: %v\n", err)
		fmt.Println("请确保服务端已启动 (go run ./server)")
		os.Exit(1)
	}

	switch user.Role {
	case common.RoleAdmin:
		runAdminMenu()
	case common.RoleManager:
		runManagerMenu()
	default:
		runEmployeeMenu()
	}
}
