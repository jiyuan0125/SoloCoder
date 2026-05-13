package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"workorder-flow/internal/database"
	"workorder-flow/internal/duty"
	"workorder-flow/internal/engine"
	"workorder-flow/internal/handler"
	"workorder-flow/internal/notification"
)

func main() {
	dbPath := "./workorder.db"
	if envPath := os.Getenv("DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	db, err := database.NewDB(dbPath)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.InitSchema(); err != nil {
		log.Fatalf("failed to initialize schema: %v", err)
	}

	if err := db.SeedInitialData(); err != nil {
		log.Fatalf("failed to seed initial data: %v", err)
	}

	scheduler := duty.NewScheduler(db)
	flowEngine := engine.NewFlowEngine(db, scheduler)
	notifier := notification.NewNotificationService(db)
	h := handler.NewHandler(db, flowEngine, scheduler, notifier)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go h.StartOverdueChecker(ctx, 1*time.Minute)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Work Order Flow System API", "version": "1.0.0"}`))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	mux.HandleFunc("/workorders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreateWorkOrder(w, r)
		} else if r.Method == http.MethodGet {
			h.ListWorkOrders(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/workorders/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/workorders/")

		if path == "" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if strings.HasSuffix(path, "/claim") {
			if r.Method == http.MethodPost {
				h.ClaimWorkOrder(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/advance") {
			if r.Method == http.MethodPost {
				h.AdvanceWorkOrder(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/confirm") {
			if r.Method == http.MethodPost {
				h.ConfirmWorkOrder(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/reassign") {
			if r.Method == http.MethodPost {
				h.ReassignWorkOrder(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/escalate") {
			if r.Method == http.MethodPost {
				h.EscalateWorkOrder(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/history") {
			if r.Method == http.MethodGet {
				h.GetWorkOrderHistory(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/communicate") {
			if r.Method == http.MethodPost {
				h.AddCommunication(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/communications") {
			if r.Method == http.MethodGet {
				h.GetWorkOrderCommunications(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if r.Method == http.MethodGet {
			h.GetWorkOrder(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/resource-types", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.ListResourceTypes(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/resource-types/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/resource-types/")

		if !strings.HasSuffix(path, "/resources") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if r.Method == http.MethodGet {
			h.ListResourcesByType(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/resources/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/resources/")

		if !strings.HasSuffix(path, "/workorders") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if r.Method == http.MethodGet {
			h.GetResourceWorkOrders(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := ":9100"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	go func() {
		log.Printf("server starting on %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("server stopped")
}
