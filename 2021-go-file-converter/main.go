package main

import (
	"log"
	"net/http"

	"file-converter/internal/db"
	"file-converter/internal/handler"
)

func main() {
	if err := db.Init("./converter.db"); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	log.Println("数据库初始化成功")

	mux := http.NewServeMux()
	handler.InitRoutes(mux)

	log.Println("文件格式转换服务启动在端口 8500")
	log.Println("支持的格式: CSV<->JSON, JSON<->YAML, Markdown->HTML")
	log.Println("API端点:")
	log.Println("  GET  /formats                - 查看支持的格式")
	log.Println("  POST /convert                - 上传文件并转换")
	log.Println("  GET  /jobs/{id}              - 查看任务详情")
	log.Println("  GET  /resources              - 列出所有资源")
	log.Println("  POST /resources              - 创建资源")
	log.Println("  GET  /resources/{id}         - 查看资源详情")
	log.Println("  GET  /resources/{id}/jobs    - 查看资源的任务汇总")
	log.Println("  POST /relations              - 创建资源关联关系")

	if err := http.ListenAndServe(":8500", mux); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
