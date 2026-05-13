package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"router/internal/api"
	"router/internal/proxy"
	"router/internal/router"
	"router/internal/store"
)

func main() {
	dbPath := "./routes.db"
	if envDB := os.Getenv("DB_PATH"); envDB != "" {
		dbPath = envDB
	}

	port := "8105"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	st, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer st.Close()

	matcher := router.NewMatcher()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	routes, err := st.GetAll(ctx)
	cancel()
	if err != nil {
		log.Fatalf("Failed to load routes: %v", err)
	}
	matcher.Update(routes)

	p := proxy.New(matcher)
	apiHandler := api.New(st, matcher)

	mux := http.NewServeMux()
	apiHandler.Register(mux)
	mux.Handle("/", p)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server stopped")
}
