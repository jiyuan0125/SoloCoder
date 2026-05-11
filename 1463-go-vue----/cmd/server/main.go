package main

import (
	"flag"
	"net/http"
	"os"

	"safetymanager/internal/core"
)

func getPort() string {
	port := os.Getenv("SAFETY_MANAGER_PORT")
	if port == "" {
		port = "8080"
	}

	flagPort := flag.String("port", "", "Server port")
	flag.Parse()

	if *flagPort != "" {
		port = *flagPort
	}

	return ":" + port
}

func main() {
	port := getPort()
	service := core.NewService()
	server := &Server{service: service}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /zones", server.handleCreateZone)
	mux.HandleFunc("GET /zones", server.handleListZones)

	mux.HandleFunc("POST /plans", server.handleCreatePlan)
	mux.HandleFunc("GET /plans", server.handleListPlans)

	mux.HandleFunc("POST /tasks/generate", server.handleGenerateTasks)
	mux.HandleFunc("GET /tasks", server.handleListTasks)
	mux.HandleFunc("POST /inspections", server.handleSubmitInspection)

	mux.HandleFunc("GET /hazards", server.handleListHazards)
	mux.HandleFunc("POST /hazards/remediate", server.handleSubmitRemediation)
	mux.HandleFunc("POST /hazards/review", server.handleReviewRemediation)

	mux.HandleFunc("POST /hazards/level-change/request", server.handleRequestLevelChange)
	mux.HandleFunc("POST /hazards/level-change/review", server.handleReviewLevelChange)
	mux.HandleFunc("POST /hazards/escalations/check", server.handleCheckEscalations)

	mux.HandleFunc("GET /audit-logs", server.handleListAuditLogs)
	mux.HandleFunc("GET /audit-logs/critical", server.handleExportCriticalLogs)

	http.ListenAndServe(port, mux)
}
