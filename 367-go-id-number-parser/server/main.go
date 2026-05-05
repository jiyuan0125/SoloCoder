package main

import (
	"log"
	"net/http"
)

func main() {
	router := setupRouter()

	log.Println("服务端启动，监听端口 :8080")
	log.Println("可用接口:")
	log.Println("  POST /api/parse - 解析身份证号")
	log.Println("  POST /api/validate - 校验身份证号")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
