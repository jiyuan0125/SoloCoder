package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/food-menu/server"
)

const (
	defaultPort = ":8080"
)

func main() {
	store := server.NewStore()
	persistence := server.NewPersistence(store)

	log.Println("正在加载数据...")
	if err := persistence.LoadAll(); err != nil {
		log.Printf("加载数据失败: %v", err)
	} else {
		log.Println("数据加载完成")
	}

	handler := server.NewHandler(store, persistence)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	} else if port[0] != ':' {
		port = ":" + port
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n收到关闭信号，正在保存数据...")
		if err := persistence.SaveAll(); err != nil {
			log.Printf("保存数据失败: %v", err)
		} else {
			log.Println("数据保存完成")
		}
		os.Exit(0)
	}()

	log.Printf("服务端启动成功，监听端口 %s", port)
	log.Printf("按 Ctrl+C 停止服务")

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
