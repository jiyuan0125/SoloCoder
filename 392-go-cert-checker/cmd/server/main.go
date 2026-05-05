package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cert-checker/internal/server"
)

const defaultPort = 8080

func main() {
	port := flag.Int("port", defaultPort, "服务监听端口")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	srv := server.NewServer()

	go func() {
		log.Printf("证书检查服务启动，监听端口 %d", *port)
		if err := srv.Start(addr); err != nil {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("正在停止服务...")
	srv.Stop()
	log.Println("服务已停止")
}
