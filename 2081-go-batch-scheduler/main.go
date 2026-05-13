package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-batch-scheduler/api"
	"go-batch-scheduler/db"
	"go-batch-scheduler/models"
	"go-batch-scheduler/scheduler"
)

func main() {
	workerCount := flag.Int("workers", 4, "number of worker goroutines")
	port := flag.String("port", "8080", "server port")
	dbPath := flag.String("db", "data/scheduler.db", "sqlite database path")
	flag.Parse()

	log.Printf("Initializing database at %s", *dbPath)

	database, err := db.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	log.Printf("Creating default resources if not exists...")
	defaultResources := []*models.Resource{
		{ID: "res-ds01", Name: "DataSource-MySQL", Type: "database"},
		{ID: "res-fs01", Name: "FileSystem-S3", Type: "storage"},
		{ID: "res-fs02", Name: "FileSystem-Local", Type: "storage"},
	}
	for _, r := range defaultResources {
		existing, err := database.GetResource(r.ID)
		if err != nil {
			log.Printf("Error checking resource %s: %v", r.ID, err)
			continue
		}
		if existing == nil {
			if err := database.CreateResource(r); err != nil {
				log.Printf("Error creating resource %s: %v", r.ID, err)
			}
		}
	}

	log.Printf("Starting scheduler with %d workers...", *workerCount)
	sched := scheduler.New(database, *workerCount)
	sched.Start()
	defer sched.Stop()

	handler := api.NewHandler(database, sched)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/resources", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetResources(w, r)
		case http.MethodPost:
			handler.CreateResource(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/resources/relations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.CreateResourceRelation(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetAllTasks(w, r)
		case http.MethodPost:
			handler.SubmitTask(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/tasks/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if len(path) > len("/api/tasks/") && path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}

		if len(path) > len("/api/tasks/") {
			remaining := path[len("/api/tasks/"):]

			if remaining == "" {
				if r.Method == http.MethodGet {
					handler.GetAllTasks(w, r)
					return
				}
			}

			if remaining == "stats" {
				if r.Method == http.MethodGet {
					handler.GetStats(w, r)
					return
				}
			}

			if remaining == "summaries" {
				if r.Method == http.MethodGet {
					handler.GetResourceSummaries(w, r)
					return
				}
			}

			if len(remaining) > len("") {
				parts := splitPath(remaining)
				if len(parts) >= 2 {
					if parts[1] == "retry" {
						if r.Method == http.MethodPost {
							handler.RetryTask(w, r)
							return
						}
					} else if parts[1] == "flow" {
						if r.Method == http.MethodPost {
							handler.AdvanceFlow(w, r)
							return
						}
					}
				}

				if len(parts) >= 1 && r.Method == http.MethodGet {
					handler.GetTask(w, r)
					return
				}
			}
		}
	})

	mux.HandleFunc("/api/stats", handler.GetStats)

	mux.HandleFunc("/api/resources/summaries", handler.GetResourceSummaries)

	server := &http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}

	go func() {
		log.Printf("Server starting on :%s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Goodbye")
}

func splitPath(path string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}
