package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		port     int
		logFile  string
		logLevel string
	)

	flag.IntVar(&port, "port", 9876, "服务监听端口")
	flag.StringVar(&logFile, "log", "deploy-server.log", "日志文件路径")
	flag.StringVar(&logLevel, "level", "info", "日志级别: debug, info, warn, error")
	flag.Parse()

	if err := initLogger(logFile, logLevel); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	Info("部署服务启动，端口: %d", port)

	server := NewServer(fmt.Sprintf(":%d", port))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		Info("收到信号: %v，正在停止服务", sig)
		server.Stop()
	}()

	if err := server.Start(); err != nil {
		Error("服务启动失败: %v", err)
		os.Exit(1)
	}

	Info("服务已停止")
}
