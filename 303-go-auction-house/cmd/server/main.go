package main

import (
	"auction-house/internal/server/handler"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	handler.SetupRoutes()

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("拍卖服务端启动，监听端口: %s\n", port)
	fmt.Println("按 Ctrl+C 停止服务")

	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	<-stop
	fmt.Println("\n服务已停止")
}
