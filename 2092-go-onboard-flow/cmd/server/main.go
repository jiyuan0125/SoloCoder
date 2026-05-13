package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"onboard-flow/internal/db"
	"onboard-flow/internal/handler"
	"onboard-flow/internal/scheduler"
	"onboard-flow/internal/service"
)

func main() {
	dbPath := "./onboard.db"
	if envPath := os.Getenv("DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	if err := db.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	flowService := service.NewFlowService()
	h := handler.NewHandler(flowService)

	sched := scheduler.NewScheduler()
	sched.Start()
	defer sched.Stop()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/employees", h.CreateEmployee)
	mux.HandleFunc("GET /api/employees/{id}", h.GetEmployee)

	mux.HandleFunc("POST /api/employees/{id}/steps/start", h.StartStep)
	mux.HandleFunc("POST /api/employees/{id}/steps/complete", h.CompleteStep)

	mux.HandleFunc("POST /api/employees/{id}/background-check", h.ProcessBackgroundCheck)
	mux.HandleFunc("POST /api/employees/{id}/account-setup", h.UpdateAccountSetup)

	mux.HandleFunc("POST /api/mentors", h.CreateMentor)
	mux.HandleFunc("POST /api/employees/{id}/assign-mentor", h.AssignMentor)

	mux.HandleFunc("POST /api/employees/{id}/exams", h.SubmitExam)

	mux.HandleFunc("POST /api/employees/{id}/mentor-review", h.SubmitMentorReview)
	mux.HandleFunc("POST /api/employees/{id}/manager-review", h.SubmitManagerReview)

	mux.HandleFunc("POST /api/employees/{id}/delays", h.RecordDelay)
	mux.HandleFunc("GET /api/employees/{id}/delays", h.GetDelayRecords)

	mux.HandleFunc("POST /api/employees/{id}/update-amount", h.UpdateTotalAmount)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("Server starting on port 8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
}
