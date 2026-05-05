package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/notify-relay/internal/cli"
)

var (
	addr = flag.String("addr", "127.0.0.1:9876", "服务地址")
)

func printUsage() {
	fmt.Println("用法: notify-cli [选项] <命令> [参数]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -addr string    服务地址 (默认 \"127.0.0.1:9876\")")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  add <文件路径>     添加监控文件")
	fmt.Println("  remove <文件路径>  移除监控文件")
	fmt.Println("  list               列出所有监控文件")
	fmt.Println("  status             查看服务状态")
	fmt.Println("  logs               查看转发日志")
	fmt.Println()
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]
	client := cli.NewClient(*addr)

	var err error

	switch command {
	case "add":
		if len(args) < 2 {
			fmt.Println("错误: add 命令需要指定文件路径")
			os.Exit(1)
		}
		err = client.AddWatch(args[1])
	case "remove":
		if len(args) < 2 {
			fmt.Println("错误: remove 命令需要指定文件路径")
			os.Exit(1)
		}
		err = client.RemoveWatch(args[1])
	case "list":
		err = client.ListWatches()
	case "status":
		err = client.GetStatus()
	case "logs":
		err = client.GetForwardLogs()
	default:
		fmt.Printf("错误: 未知命令 '%s'\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}
