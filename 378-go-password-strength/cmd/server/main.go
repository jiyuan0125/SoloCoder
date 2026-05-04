package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("port", 8080, "服务端口")
	flag.Parse()

	handler := NewEvaluateHandler()

	http.Handle("/evaluate", handler)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("密码强度评估服务启动，监听端口 %d", *port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
