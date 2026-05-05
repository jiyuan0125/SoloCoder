package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

var (
	port    = flag.Int("port", 8080, "Server port")
	address = flag.String("addr", "0.0.0.0", "Server address")
)

func main() {
	flag.Parse()

	metricManager := NewMetricManager()
	alertManager := metricManager.alertManager
	connectionManager := NewConnectionManager(metricManager, alertManager)
	layoutManager := NewLayoutManager()

	handler := NewAPIHandler(metricManager, connectionManager, layoutManager)

	r := mux.NewRouter()

	r.HandleFunc("/ws", handler.HandleWebSocket).Methods("GET")

	r.HandleFunc("/api/status", handler.GetServerStatus).Methods("GET")

	r.HandleFunc("/api/metrics", handler.GetMetrics).Methods("GET")
	r.HandleFunc("/api/metrics", handler.CreateMetric).Methods("POST")
	r.HandleFunc("/api/metrics/{key}", handler.GetMetric).Methods("GET")
	r.HandleFunc("/api/metrics/{key}/value", handler.UpdateMetricValue).Methods("PUT", "POST")
	r.HandleFunc("/api/metrics/{key}/threshold", handler.SetMetricThreshold).Methods("PUT", "POST")
	r.HandleFunc("/api/metrics/{key}/history", handler.GetMetricHistory).Methods("GET")
	r.HandleFunc("/api/metrics/{key}/trend", handler.GetMetricTrend).Methods("GET")
	r.HandleFunc("/api/metrics/{key}/comparison", handler.GetMetricComparison).Methods("GET")

	r.HandleFunc("/api/dashboard/layout", handler.GetDashboardLayout).Methods("GET")
	r.HandleFunc("/api/dashboard/layout", handler.SaveDashboardLayout).Methods("POST", "PUT")

	r.HandleFunc("/api/delay/stats", handler.GetDelayStats).Methods("GET")
	r.HandleFunc("/api/alerts", handler.GetAlerts).Methods("GET")

	r.HandleFunc("/api/admin/cleanup", handler.CleanupOldData).Methods("POST")

	stopChan := make(chan struct{})
	go handler.RunBackgroundTasks(stopChan)

	addr := fmt.Sprintf("%s:%d", *address, *port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Server starting on %s...", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down server...")

	close(stopChan)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server shutdown complete")
}
