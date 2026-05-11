package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"piperepair/core"
)

func main() {
	port := getPort()

	service := core.NewPipeRepairService()
	handler := NewHandler(service)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/orders/create", handler.CreateOrder)
	mux.HandleFunc("/api/orders/get", handler.GetOrder)
	mux.HandleFunc("/api/orders/list", handler.ListOrders)
	mux.HandleFunc("/api/orders/start", handler.StartProcessing)
	mux.HandleFunc("/api/orders/submit", handler.SubmitRepairRecord)
	mux.HandleFunc("/api/orders/accept", handler.AcceptOrder)
	mux.HandleFunc("/api/masters/info", handler.GetMasterInfo)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("管道疏通报修管理系统服务端已启动，监听端口: %s", port)
	log.Printf("可用区域代码: 01001-01009 (朝阳区01001-01003, 海淀区01004-01006, 丰台区01007-01009)")
	log.Printf("可用师傅ID: m001-张师傅, m002-李师傅, m003-王师傅, m004-赵师傅, m005-刘师傅")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func getPort() string {
	port := "8080"

	flagPort := flag.String("port", "", "服务端监听端口")
	flag.Parse()

	if *flagPort != "" {
		port = *flagPort
	} else if envPort := os.Getenv("PIPEREPAIR_PORT"); envPort != "" {
		port = envPort
	}

	return port
}
