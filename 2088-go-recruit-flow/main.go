package main

import (
	"fmt"
	"net/http"
	"recruit-flow/database"
	"recruit-flow/handlers"
	"recruit-flow/service"
	"strings"
	"time"
)

func main() {
	if err := database.InitDB("recruit.db"); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}
	fmt.Println("Database initialized successfully")

	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/stages", handlers.StageInfoHandler)
	http.HandleFunc("/candidates", handlers.CreateCandidateHandler)
	http.HandleFunc("/candidates/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/advance") {
			handlers.AdvanceStageHandler(w, r)
			return
		}
		if strings.HasSuffix(path, "/reject") {
			handlers.RejectCandidateHandler(w, r)
			return
		}
		if strings.HasSuffix(path, "/tech-score") {
			handlers.SubmitTechScoreHandler(w, r)
			return
		}
		if strings.HasSuffix(path, "/history") {
			handlers.GetHistoryHandler(w, r)
			return
		}
		handlers.GetCandidateHandler(w, r)
	})

	go startScheduledTasks()

	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func startScheduledTasks() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := service.ProcessExpiredOffers(); err != nil {
			fmt.Printf("Error processing expired offers: %v\n", err)
		}
		if err := service.CheckOverdueStages(); err != nil {
			fmt.Printf("Error checking overdue stages: %v\n", err)
		}
	}
}
