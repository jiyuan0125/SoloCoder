package main

import (
	"flag"
	"fmt"
	"os"

	"luggage-tracking/client"
)

var (
	serverURL string
	showHelp  bool
)

func init() {
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "服务端地址")
	flag.BoolVar(&showHelp, "h", false, "显示帮助信息")
	flag.BoolVar(&showHelp, "help", false, "显示帮助信息")
}

func main() {
	flag.Parse()

	if showHelp {
		printHelp()
		return
	}

	fmt.Println()
	cli := client.NewCLI(serverURL)
	cli.Run()
}

func printHelp() {
	fmt.Println("航空公司行李追踪系统 - 客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client [选项]")
	fmt.Println()
	fmt.Println("选项:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("功能说明:")
	fmt.Println("  1. 工作人员模式:")
	fmt.Println("     - 扫码上报行李 (值机、安检、分拣、装机、到达、传送带)")
	fmt.Println("     - 补录遗漏的扫码节点")
	fmt.Println("     - 查询某航班当天所有行李进度")
	fmt.Println("     - 取消航班 (标记所有未完成行李为'航班取消')")
	fmt.Println()
	fmt.Println("  2. 旅客模式:")
	fmt.Println("     - 通过行李牌号查询行李进度")
	fmt.Println()
	fmt.Println("  3. 管理员模式:")
	fmt.Println("     - 查看操作日志")
	fmt.Println("     - 查看审计日志 (敏感操作)")
	fmt.Println("     - 检查服务状态")
	fmt.Println()
	fmt.Println("行李牌号规则: 10位数字或字母组合")
	fmt.Println()
	os.Exit(0)
}
