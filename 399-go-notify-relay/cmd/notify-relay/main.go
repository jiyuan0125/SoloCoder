package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/notify-relay/internal/relay"
)

var (
	addr     = flag.String("addr", ":9876", "服务监听地址")
	dataPath = flag.String("data", ".", "数据存储路径")
)

func main() {
	flag.Parse()

	absDataPath, err := filepath.Abs(*dataPath)
	if err != nil {
		log.Fatalf("无法获取数据路径: %v", err)
	}

	server, err := relay.NewServer(*addr, absDataPath)
	if err != nil {
		log.Fatalf("创建服务失败: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("收到停止信号，正在关闭服务...")
	server.Stop()
}
