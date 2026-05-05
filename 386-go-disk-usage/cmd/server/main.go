package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"disk-usage/internal/server"
)

func main() {
	addr := flag.String("addr", ":8765", "服务监听地址")
	flag.Parse()
	
	srv := server.NewTCPServer(*addr)
	
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigCh
		log.Println("\n收到停止信号，正在关闭服务...")
		srv.Stop()
		os.Exit(0)
	}()
	
	if err := srv.Start(); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
