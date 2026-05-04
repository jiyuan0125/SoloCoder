package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"realestate/pkg/server"
)

func main() {
	port := flag.Int("port", 8080, "HTTP服务端口")
	dataFile := flag.String("data", "data.json", "数据存储文件路径")
	flag.Parse()

	store := server.NewStore(*dataFile)
	handler := server.NewHandler(store)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("服务端启动，监听地址: %s", addr)
	log.Printf("数据存储文件: %s", *dataFile)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
