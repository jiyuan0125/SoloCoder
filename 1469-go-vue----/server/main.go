package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"supervision-log-system/core"
)

func main() {
	port := getPort()

	store := core.NewStore()
	router := NewRouter(store)

	log.Printf("监理日志系统服务端启动，监听端口: %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func getPort() string {
	port := "8080"

	if flagPort := flag.String("port", "", "监听端口"); flagPort != nil {
		flag.Parse()
		if *flagPort != "" {
			port = *flagPort
		}
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	if strings.TrimSpace(port) == "" {
		port = "8080"
	}

	return port
}
