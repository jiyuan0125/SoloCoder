package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "服务监听端口")
	flag.Parse()

	if port == 0 {
		envPort := os.Getenv("PORT")
		if envPort != "" {
			fmt.Sscanf(envPort, "%d", &port)
		}
	}

	if port == 0 {
		port = 8080
	}

	server := NewServer()
	router := mux.NewRouter()

	router.HandleFunc("/api/v1/vehicles", server.CreateVehicle).Methods("POST")
	router.HandleFunc("/api/v1/vehicles", server.ListVehicles).Methods("GET")
	router.HandleFunc("/api/v1/vehicles/{plate}", server.GetVehicle).Methods("GET")
	router.HandleFunc("/api/v1/vehicles/{plate}/status", server.UpdateVehicleStatus).Methods("PUT")
	router.HandleFunc("/api/v1/vehicles/{plate}/location", server.UpdateVehicleLocation).Methods("PUT")

	router.HandleFunc("/api/v1/orders/estimate", server.EstimateFare).Methods("POST")
	router.HandleFunc("/api/v1/orders", server.CreateOrder).Methods("POST")
	router.HandleFunc("/api/v1/orders", server.ListOrders).Methods("GET")
	router.HandleFunc("/api/v1/orders/{id}", server.GetOrder).Methods("GET")
	router.HandleFunc("/api/v1/orders/{id}/accept", server.AcceptOrder).Methods("POST")
	router.HandleFunc("/api/v1/orders/{id}/start", server.StartTrip).Methods("POST")
	router.HandleFunc("/api/v1/orders/{id}/complete", server.CompleteTrip).Methods("POST")

	router.HandleFunc("/api/v1/orders/{id}/rating", server.SubmitRating).Methods("POST")

	router.HandleFunc("/api/v1/complaints", server.ListComplaints).Methods("GET")
	router.HandleFunc("/api/v1/complaints/{id}/handle", server.HandleComplaint).Methods("POST")

	router.HandleFunc("/api/v1/metrics", server.GetMetrics).Methods("GET")

	addr := fmt.Sprintf(":%d", port)
	log.Printf("出租车调度管理系统服务端启动，监听端口: %d", port)
	log.Fatal(http.ListenAndServe(addr, router))
}
