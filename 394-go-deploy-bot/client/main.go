package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "deploy":
		handleDeploy()
	case "rollback":
		handleRollback()
	case "status":
		handleStatus()
	case "history":
		handleHistory()
	default:
		fmt.Printf("未知命令: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Go 自动部署工具 - 客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client deploy [选项]    提交部署任务")
	fmt.Println("  client rollback [选项]  执行回滚操作")
	fmt.Println("  client status            查看当前状态")
	fmt.Println("  client history           查看部署历史")
	fmt.Println()
	fmt.Println("部署选项:")
	fmt.Println("  --binary      二进制文件路径 (必需)")
	fmt.Println("  --backup      备份目录 (默认: backups)")
	fmt.Println("  --health-url  健康检查URL")
	fmt.Println("  --health-timeout 健康检查超时时间(秒) (默认: 5)")
	fmt.Println("  --build-cmd   编译命令")
	fmt.Println("  --pull-cmd    拉取代码命令 (如: git pull)")
	fmt.Println("  --server      服务端地址 (默认: localhost:9876)")
	fmt.Println()
	fmt.Println("回滚选项:")
	fmt.Println("  --backup-path 备份文件路径")
	fmt.Println("  --binary      二进制文件路径 (必需)")
	fmt.Println("  --backup      备份目录")
	fmt.Println("  --health-url  健康检查URL")
	fmt.Println("  --server      服务端地址 (默认: localhost:9876)")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  client deploy --binary ./app --backup ./backups --build-cmd \"go build -o app .\" --pull-cmd \"git pull\"")
	fmt.Println("  client deploy --binary ./app --health-url http://localhost:8080/health --build-cmd \"go build -o app .\"")
	fmt.Println("  client rollback --binary ./app --backup-path ./backups/app-20240101-120000")
	fmt.Println("  client status --server localhost:9876")
	fmt.Println("  client history")
}
