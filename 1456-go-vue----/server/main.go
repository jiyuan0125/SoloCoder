package main

import (
	"archivesystem/core"
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {
	var port string

	flag.StringVar(&port, "port", "", "服务端监听端口")
	flag.Parse()

	if port == "" {
		port = os.Getenv("ARCHIVE_SERVER_PORT")
	}
	if port == "" {
		port = "8080"
	}

	service := core.NewArchiveService()
	router := setupRouter(service)

	addr := ":" + port
	fmt.Printf("档案管理系统服务端启动，监听端口: %s\n", port)
	if err := http.ListenAndServe(addr, router); err != nil {
		fmt.Printf("服务启动失败: %v\n", err)
		os.Exit(1)
	}
}
