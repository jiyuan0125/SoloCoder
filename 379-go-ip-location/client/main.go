package main

import (
	"flag"
	"fmt"
	"os"
)

const defaultServerURL = "http://localhost:8080"

func main() {
	flag.Usage = usage

	serverURL := flag.String("server", defaultServerURL, "服务端地址 (默认: http://localhost:8080)")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	cmd := args[0]
	restArgs := args[1:]

	client := NewClient(*serverURL)

	switch cmd {
	case "query":
		if len(restArgs) < 1 {
			fmt.Println("用法: ip-client query <IP地址>")
			os.Exit(1)
		}
		if err := client.Query(restArgs[0]); err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}

	case "batch":
		if len(restArgs) < 1 {
			fmt.Println("用法: ip-client batch <IP1> <IP2> ...")
			os.Exit(1)
		}
		if err := client.BatchQuery(restArgs); err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}

	case "same":
		if len(restArgs) < 2 {
			fmt.Println("用法: ip-client same <IP1> <IP2>")
			os.Exit(1)
		}
		if err := client.SameProvince(restArgs[0], restArgs[1]); err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}

	case "info":
		if len(restArgs) < 1 {
			fmt.Println("用法: ip-client info <IP地址>")
			os.Exit(1)
		}
		if err := client.IPInfo(restArgs[0]); err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}

	case "health":
		if err := client.Health(); err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("未知命令: %s\n\n", cmd)
		flag.Usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("IP归属地查询客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  ip-client [选项] 命令 [参数...]")
	fmt.Println()
	fmt.Println("选项:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  query <IP>      查询单个IP的归属地")
	fmt.Println("  batch <IPs...>  批量查询多个IP的归属地")
	fmt.Println("  same <IP1> <IP2> 判断两个IP是否属于同一省份")
	fmt.Println("  info <IP>       查询IP的类型和是否为内网地址")
	fmt.Println("  health          检查服务端健康状态")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  ip-client query 1.1.1.1")
	fmt.Println("  ip-client batch 1.1.1.1 2.2.2.2 3.3.3.3")
	fmt.Println("  ip-client same 1.1.1.1 14.0.0.1")
	fmt.Println("  ip-client info 192.168.1.1")
}
