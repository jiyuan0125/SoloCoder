package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	port := flag.String("port", ":8080", "HTTP服务监听端口")
	dataFile := flag.String("data", "enrollment_data.json", "数据持久化文件路径")
	flag.Parse()

	store := NewStore(*dataFile)
	handler := NewHandler(store)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /plans", handler.CreatePlan)
	mux.HandleFunc("GET /plans", handler.ListPlans)
	mux.HandleFunc("PUT /plans", handler.UpdatePlan)
	mux.HandleFunc("POST /plans/close", handler.ClosePlan)

	mux.HandleFunc("POST /registrations", handler.SubmitRegistration)
	mux.HandleFunc("GET /registrations", handler.ListRegistrations)
	mux.HandleFunc("GET /registrations/query", handler.QueryRegistration)
	mux.HandleFunc("POST /registrations/review", handler.ReviewRegistration)

	log.Printf("招生报名服务启动，监听端口 %s", *port)
	log.Printf("数据文件: %s", *dataFile)
	if err := http.ListenAndServe(*port, mux); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
