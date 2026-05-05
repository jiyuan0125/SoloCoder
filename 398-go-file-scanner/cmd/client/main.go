package main

import (
	"fmt"
	"log"
	"os"

	"filescanner/internal/client"
	"filescanner/internal/protocol"
)

func main() {
	args, err := client.ParseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nUse -help for usage information.\n")
		os.Exit(1)
	}

	fmt.Printf("Connecting to server: %s\n", args.ServerAddr)
	fmt.Printf("Scanning directory: %s\n", args.Directory)

	if args.Severity != "" {
		fmt.Printf("Filtering by severity: %s and above\n", args.Severity)
	}

	if args.ConfigFile != "" {
		fmt.Printf("Using config file: %s\n", args.ConfigFile)
	}

	fmt.Println()

	// 创建客户端并连接
	c := client.NewClient(args.ServerAddr)
	if err := c.Connect(); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer c.Disconnect()

	fmt.Println("Connected to server successfully.")
	fmt.Println("Sending scan request...")
	fmt.Println()

	// 构建扫描请求
	req := &protocol.ScanRequest{
		Directory:  args.Directory,
		ConfigFile: args.ConfigFile,
		Severity:   args.Severity,
		OutputJSON: args.OutputJSON,
	}

	// 发送扫描请求
	resp, err := c.SendScanRequest(req)
	if err != nil {
		log.Fatalf("Scan failed: %v", err)
	}

	// 打印结果
	client.PrintResults(resp, args.OutputJSON)

	// 根据是否有匹配设置退出码
	if resp.Summary.TotalMatchesFound > 0 {
		os.Exit(1)
	}
}
