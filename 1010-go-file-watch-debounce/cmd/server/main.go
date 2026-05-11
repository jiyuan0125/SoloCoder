package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"file-watch-debounce/server"
)

func main() {
	port := flag.String("port", "8101", "HTTP server port")
	flag.Parse()

	srv := server.NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/watch/add", srv.AddWatchHandler)
	mux.HandleFunc("/api/watch/remove", srv.RemoveWatchHandler)
	mux.HandleFunc("/api/watches", srv.ListWatchesHandler)
	mux.HandleFunc("/api/events", srv.ListEventsHandler)

	httpServer := &http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on port %s...", *port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	fmt.Println("\nShutting down server...")
}
