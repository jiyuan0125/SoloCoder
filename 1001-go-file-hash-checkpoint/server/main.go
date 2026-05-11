package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"hash-checkpoint/core"
)

func main() {
	addr := flag.String("addr", ":8100", "HTTP server address")
	flag.Parse()

	manager := core.NewCheckpointManager()
	handler := NewHandler(manager)

	mux := http.NewServeMux()
	mux.HandleFunc("/check/start", handler.StartCheck)
	mux.HandleFunc("/check/chunk", handler.SubmitChunk)
	mux.HandleFunc("/check/progress", handler.GetProgress)
	mux.HandleFunc("/check/complete", handler.CompleteCheck)

	server := &http.Server{
		Addr:         *addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Printf("hash-checkpoint server listening on %s\n", *addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
