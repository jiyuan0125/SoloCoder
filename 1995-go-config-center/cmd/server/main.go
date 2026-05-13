package main

import (
	"config-center/internal/server"
	"config-center/internal/store"
	"config-center/internal/watch"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10007"
	}

	s := store.New()
	h := server.NewHandler(s)
	wm := watch.NewManager(s)

	s.AddChangeListener(wm.HandleChange)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := ":" + port
	log.Printf("config center server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
