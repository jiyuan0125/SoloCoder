package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ssh-tunnel-manager/api"
	"ssh-tunnel-manager/database"
	"ssh-tunnel-manager/manager"
	"ssh-tunnel-manager/workflow"
)

func main() {
	port := flag.Int("port", 8080, "Server port")
	dataDir := flag.String("data", "./data", "Data directory")
	flag.Parse()

	db, err := database.Init(*dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	mgr := manager.New(db)
	wf := workflow.NewService(db)

	apiServer := api.NewServer(db, mgr, wf)

	mux := http.NewServeMux()
	apiServer.Register(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
	}

	go func() {
		log.Printf("SSH Tunnel Manager server starting on port %d...", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
}
