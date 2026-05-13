package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"connpool/internal/api"
	"connpool/internal/pool"
	"connpool/internal/stats"
	"connpool/internal/storage"
)

const (
	defaultDBPath = "connpool.db"
	defaultPort   = ":8080"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	log.Printf("Starting connection pool manager...")
	log.Printf("DB Path: %s", dbPath)
	log.Printf("Port: %s", port)

	store, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	poolMgr := pool.NewManager()
	statsMgr := stats.NewManager(store, poolMgr)

	ctx := context.Background()
	configs, err := store.ListPools(ctx)
	if err != nil {
		log.Fatalf("Failed to load pools from storage: %v", err)
	}

	for _, cfg := range configs {
		if err := poolMgr.CreatePool(cfg); err != nil {
			log.Printf("Warning: Failed to create pool %s: %v", cfg.ID, err)
		}
	}

	log.Printf("Loaded %d pools from storage", len(configs))

	handler := api.NewHandler(store, poolMgr, statsMgr)
	mux := http.NewServeMux()
	handler.Register(mux)

	server := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	healthCheckCtx, healthCheckCancel := context.WithCancel(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ids := poolMgr.ListPoolIDs()
				for _, id := range ids {
					if p, err := poolMgr.GetPool(id); err == nil {
						p.CheckHealth()
					}
				}
				statsMgr.UpdateAndSaveStats(healthCheckCtx)
			case <-healthCheckCtx.Done():
				return
			}
		}
	}()

	go func() {
		log.Printf("Server listening on %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-stop
	log.Printf("Shutting down...")

	healthCheckCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	wg.Wait()
	log.Printf("Goodbye")
}
