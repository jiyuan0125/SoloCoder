package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/comment-moderator/internal/server"
)

func main() {
	port := flag.Int("port", 8080, "服务端口")
	flag.Parse()

	store := server.NewStore()
	svc := server.NewService(store)
	handler := server.NewHandler(svc)
	router := handler.SetupRouter()

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("评论审核服务启动，监听端口 %d...", *port)
	log.Printf("健康检查: http://localhost:%d/health", *port)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
