package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	port := flag.String("port", "8765", "服务监听端口")
	flag.Parse()

	history, err := NewHistoryManager()
	if err != nil {
		fmt.Printf("初始化历史记录管理器失败: %v\n", err)
		os.Exit(1)
	}

	server := NewServer(*port, history)

	go func() {
		if err := server.Start(); err != nil {
			fmt.Printf("服务器启动失败: %v\n", err)
			os.Exit(1)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n正在关闭服务...")
	server.Stop()
	fmt.Println("服务已关闭")
}
