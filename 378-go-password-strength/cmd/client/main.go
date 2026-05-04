package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "服务端地址")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("用法: password-client [选项] <密码>")
		fmt.Println("选项:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	password := args[0]

	client := NewClient(*serverURL)
	resp, err := client.Evaluate(password)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	PrintResponse(resp)
}
