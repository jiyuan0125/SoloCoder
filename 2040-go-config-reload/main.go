package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"config-reload/api"
	"config-reload/config"
	"config-reload/storage"
)

const (
	port           = "8080"
	configFilePath = "config.yaml"
	dbFilePath     = "config_history.db"
	reloadInterval = 10 * time.Second
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("Starting config reload service...")

	cm := config.NewConfigManager(configFilePath)

	if err := cm.LoadInitial(); err != nil {
		log.Fatalf("Failed to load initial config: %v", err)
	}
	log.Println("Initial config loaded successfully")

	store, err := storage.NewStorage(dbFilePath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()
	log.Println("Storage initialized successfully")

	notifier := make(chan config.ConfigChangeNotification, 100)
	cm.RegisterNotifier(notifier)

	var wg sync.WaitGroup
	wg.Add(1)
	go handleConfigChanges(notifier, store, &wg)

	cm.StartReloadTicker(reloadInterval)
	log.Printf("Reload ticker started with interval: %v", reloadInterval)

	apiServer := api.NewAPI(cm, store)
	mux := http.NewServeMux()
	apiServer.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("HTTP server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case <-shutdown:
		log.Println("Shutdown signal received")
	case err := <-serverErrors:
		log.Printf("Server error: %v", err)
	}

	log.Println("Initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	cm.StopReloadTicker()
	close(notifier)

	wg.Wait()

	log.Println("Service stopped gracefully")
}

func handleConfigChanges(notifier <-chan config.ConfigChangeNotification, store *storage.Storage, wg *sync.WaitGroup) {
	defer wg.Done()

	for notification := range notifier {
		change := notification.Change
		if change == nil {
			continue
		}

		log.Printf("Config change detected at %v", change.Timestamp)
		log.Printf("  Old hash: %s", change.OldHash)
		log.Printf("  New hash: %s", change.Hash)

		if len(change.Added) > 0 {
			log.Printf("  Added services: %v", change.Added)
		}
		if len(change.Removed) > 0 {
			log.Printf("  Removed services: %v", change.Removed)
		}
		if len(change.Modified) > 0 {
			log.Printf("  Modified services: %v", change.Modified)
		}

		message := fmt.Sprintf("Config reloaded at %v", change.Timestamp)
		if err := store.RecordChange(change, storage.ConfigToJSON(notification.Config), message, true); err != nil {
			log.Printf("WARNING: Failed to record config change: %v", err)
		}
	}

	log.Println("Config change handler stopped")
}
