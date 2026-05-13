package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"email-template/internal/config"
	"email-template/internal/database"
	"email-template/internal/email"
	"email-template/internal/handler"
	"email-template/internal/template"
)

func main() {
	configPath := "config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	tplSvc := template.NewService(db)
	mailSvc := email.NewService(cfg, db, tplSvc)
	h := handler.NewHandler(tplSvc, mailSvc)

	mux := http.NewServeMux()
	h.Register(mux)

	server := &http.Server{
		Addr:    ":8402",
		Handler: mux,
	}

	mailSvc.StartRetryWorker()

	go func() {
		log.Printf("Server starting on :8402")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	mailSvc.Stop()
	log.Println("Server stopped")
}
