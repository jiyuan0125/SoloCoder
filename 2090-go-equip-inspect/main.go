package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"equip-inspect/database"
	"equip-inspect/handler"
	"equip-inspect/service"
)

func main() {
	dbPath := "./data/inspect.db"
	if envPath := os.Getenv("DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	log.Println("Database initialized successfully")

	go startScheduler()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/tasks", handler.ListTasksHandler)
	mux.HandleFunc("GET /api/tasks/{id}", handler.GetTaskHandler)
	mux.HandleFunc("POST /api/tasks/{id}/assign", handler.AssignTaskHandler)
	mux.HandleFunc("GET /api/tasks/{id}/points", handler.GetTaskPointsHandler)
	mux.HandleFunc("POST /api/tasks/{id}/checkin", handler.CheckInPointHandler)
	mux.HandleFunc("POST /api/tasks/{id}/submit", handler.SubmitInspectionHandler)
	mux.HandleFunc("POST /api/tasks/{id}/transition", handler.TransitionTaskHandler)

	mux.HandleFunc("GET /api/plans", handler.ListPlansHandler)
	mux.HandleFunc("POST /api/plans", handler.CreatePlanHandler)
	mux.HandleFunc("POST /api/plans/{id}/generate", handler.GenerateTasksHandler)

	mux.HandleFunc("GET /api/repairs", handler.ListRepairOrdersHandler)
	mux.HandleFunc("POST /api/repairs/{id}/assign", handler.AssignRepairHandler)
	mux.HandleFunc("POST /api/repairs/{id}/complete", handler.CompleteRepairHandler)

	mux.HandleFunc("GET /api/statistics", handler.GetStatisticsHandler)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := ":8080"
	if envAddr := os.Getenv("PORT"); envAddr != "" {
		addr = ":" + envAddr
	}

	log.Printf("Server starting on port %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func startScheduler() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Println("Scheduler started")
	for range ticker.C {
		generateScheduledTasks()
	}
}

func generateScheduledTasks() {
	rows, err := database.DB.Query(`SELECT id FROM inspection_plans WHERE active = 1`)
	if err != nil {
		log.Printf("Scheduler error: %v", err)
		return
	}
	defer rows.Close()

	var planIDs []int64
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			continue
		}
		planIDs = append(planIDs, id)
	}

	for _, planID := range planIDs {
		tasks, err := service.GenerateTasksFromPlan(planID)
		if err != nil {
			log.Printf("Failed to generate tasks for plan %d: %v", planID, err)
			continue
		}
		if tasks != nil && len(tasks) > 0 {
			log.Printf("Generated %d tasks for plan %d", len(tasks), planID)
		}
	}
}
