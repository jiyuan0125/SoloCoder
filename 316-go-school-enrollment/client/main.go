package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "服务端地址")
	flag.Parse()

	client := NewAPIClient(*serverURL)
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("====================================")
	fmt.Println("  学校在线招生报名系统 - 客户端")
	fmt.Printf("  服务端: %s\n", *serverURL)
	fmt.Println("====================================")

	for {
		fmt.Println("\n=== 主菜单 ===")
		fmt.Println("1. 家长入口")
		fmt.Println("2. 管理员入口")
		fmt.Println("0. 退出")
		fmt.Print("\n请选择: ")

		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(input)

		switch choice {
		case "1":
			runParentMode(client)
		case "2":
			runAdminMode(client)
		case "0":
			fmt.Println("再见!")
			return
		default:
			fmt.Println("无效选择，请重试")
		}
	}
}
