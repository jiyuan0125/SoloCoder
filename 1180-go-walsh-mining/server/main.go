package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "监听端口")
	flag.Parse()

	if port == 0 {
		if envPort := os.Getenv("PORT"); envPort != "" {
			fmt.Sscanf(envPort, "%d", &port)
		}
	}
	if port == 0 {
		port = 8404
	}

	http.HandleFunc("/import", importHandler)
	http.HandleFunc("/weight", setWeightsHandler)
	http.HandleFunc("/mine", mineHandler)
	http.HandleFunc("/items", itemsHandler)
	http.HandleFunc("/rules", rulesHandler)
	http.HandleFunc("/stats", statsHandler)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("WALS Mining Server 启动，监听端口 %d\n", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "服务启动失败: %v\n", err)
		os.Exit(1)
	}
}
