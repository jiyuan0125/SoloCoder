package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP server address")
	flag.Parse()

	handler := NewHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/tasks", handler.ListTasks)
	mux.HandleFunc("/tasks/add", handler.AddTask)
	mux.HandleFunc("/tasks/batch-add", handler.BatchAddTasks)
	mux.HandleFunc("/tasks/remove", handler.RemoveTask)
	mux.HandleFunc("/tasks/update-deps", handler.UpdateDependencies)
	mux.HandleFunc("/tasks/clear", handler.ClearTasks)
	mux.HandleFunc("/sort", handler.Sort)
	mux.HandleFunc("/dot", handler.GenerateDot)
	mux.HandleFunc("/dot-critical", handler.GenerateDotWithCriticalPath)

	server := &http.Server{
		Addr:    *addr,
		Handler: mux,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on %s...", *addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	sig := <-sigChan
	log.Printf("Received signal: %v, shutting down...", sig)
}
