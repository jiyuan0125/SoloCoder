package main

import (
	"fmt"
	"log"
	"net/http"
	"userbehavior/internal/server"
)

func main() {
	store := server.NewStore()
	analytics := server.NewAnalyticsEngine(store)
	qualityMon := server.NewQualityMonitor()
	handler := server.NewHandler(store, analytics, qualityMon)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/track", handler.TrackBehavior)
	mux.HandleFunc("/analytics/funnel", handler.FunnelAnalysis)
	mux.HandleFunc("/analytics/retention", handler.RetentionAnalysis)
	mux.HandleFunc("/analytics/paths", handler.PathAnalysis)
	mux.HandleFunc("/user/profile", handler.GetUserProfile)
	mux.HandleFunc("/realtime/stats", handler.GetRealtimeStats)

	port := 8080
	log.Printf("Server starting on port %d...", port)
	log.Printf("Available endpoints:")
	log.Printf("  GET  /health")
	log.Printf("  POST /track")
	log.Printf("  POST /analytics/funnel")
	log.Printf("  POST /analytics/retention")
	log.Printf("  POST /analytics/paths")
	log.Printf("  GET  /user/profile?user_id=<id>")
	log.Printf("  GET  /realtime/stats")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
