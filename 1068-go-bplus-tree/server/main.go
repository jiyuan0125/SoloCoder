package main

import (
	"log"
	"net/http"
)

const defaultPort = ":8080"

func main() {
	handler := NewHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/insert", handler.Insert)
	mux.HandleFunc("/search", handler.Search)
	mux.HandleFunc("/delete", handler.Delete)
	mux.HandleFunc("/range", handler.Range)
	mux.HandleFunc("/scan", handler.Scan)

	log.Printf("B+树索引服务正在启动，监听端口 %s", defaultPort)
	if err := http.ListenAndServe(defaultPort, mux); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
