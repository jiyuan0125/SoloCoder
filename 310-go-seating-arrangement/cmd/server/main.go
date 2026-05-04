package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"seating-arrangement/internal/server"
)

const (
	storageFile = "data.json"
	port        = ":8080"
)

func main() {
	log.Println("启动场馆座位选座系统服务端...")

	venueManager := server.NewVenueManager()
	sessionManager := server.NewSessionManager(venueManager)
	orderManager := server.NewOrderManager(sessionManager)

	storage := server.NewStorage(storageFile, venueManager, sessionManager, orderManager)

	log.Println("加载持久化数据...")
	if err := storage.Load(); err != nil {
		log.Printf("警告: 加载数据失败: %v", err)
	}

	handler := server.NewHandler(venueManager, sessionManager, orderManager, storage)

	handler.StartLockCleaner(1 * time.Minute)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/venues", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.CreateVenue(w, r)
		} else if r.Method == http.MethodGet {
			handler.ListVenues(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.CreateSession(w, r)
		} else if r.Method == http.MethodGet {
			handler.ListSessions(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/seats/lock", handler.LockSeat)
	mux.HandleFunc("/api/orders/confirm", handler.ConfirmOrder)
	mux.HandleFunc("/api/orders/cancel", handler.CancelOrder)
	mux.HandleFunc("/api/seats/release", handler.ReleaseSeat)
	mux.HandleFunc("/api/seats/available", handler.GetAvailableSeats)

	mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path[len(path)-6:] == "/stats" {
			handler.GetSessionStats(w, r)
		} else if path[len(path)-5:] == "/sold" {
			handler.ExportSoldSeats(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	log.Printf("服务端已启动，监听端口 %s", port)
	log.Printf("数据文件: %s", storageFile)
	
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
		os.Exit(1)
	}
}
