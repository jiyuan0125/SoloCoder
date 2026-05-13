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

	"votingsystem/internal/db"
	"votingsystem/internal/handler"
	"votingsystem/internal/service"
)

func main() {
	store, err := db.NewStore("./vote.db")
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	defer store.Close()

	svc := service.NewService(store)
	h := handler.NewHandler(svc)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/owners", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListOwners(w, r)
		case http.MethodPost:
			h.CreateOwner(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/votings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListVotings(w, r)
		case http.MethodPost:
			h.CreateVoting(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/votings/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/votings/")
		parts := strings.Split(path, "/")

		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		votingID := parts[0]

		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.GetVoting(w, r, votingID)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		subResource := parts[1]
		switch subResource {
		case "votes":
			if r.Method == http.MethodGet {
				h.ListVotes(w, r, votingID)
			} else if r.Method == http.MethodPost {
				h.Vote(w, r, votingID)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		case "proxies":
			if r.Method == http.MethodGet {
				h.ListProxies(w, r, votingID)
			} else if r.Method == http.MethodPost {
				h.Delegate(w, r, votingID)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		case "logs":
			if r.Method == http.MethodGet {
				h.ListAuditLogs(w, r, votingID)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	server := &http.Server{
		Addr:    ":8301",
		Handler: mux,
	}

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			if err := svc.ProcessExpiredVotings(); err != nil {
				log.Printf("error processing expired votings: %v", err)
			}
		}
	}()

	go func() {
		log.Printf("server starting on :8301")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	log.Println("server stopped")
}
