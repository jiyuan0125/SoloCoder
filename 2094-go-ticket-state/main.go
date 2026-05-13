package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"ticket-system/database"
	"ticket-system/handler"
	"ticket-system/scheduler"
	"ticket-system/service"
)

const (
	DefaultPort = "8080"
	DBPath      = "./tickets.db"
)

func getPort() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return DefaultPort
}

type route struct {
	pattern *regexp.Regexp
	method  string
	handler http.HandlerFunc
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.Open(DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	log.Println("database initialized")

	ticketService := service.NewTicketService(db)
	h := handler.NewHandler(ticketService, db)

	routes := []route{
		{regexp.MustCompile(`^/health$`), http.MethodGet, h.Health},
		{regexp.MustCompile(`^/tickets$`), http.MethodPost, h.CreateTicket},
		{regexp.MustCompile(`^/tickets$`), http.MethodGet, h.ListTickets},
		{regexp.MustCompile(`^/tickets/(\d+)$`), http.MethodGet, h.GetTicket},
		{regexp.MustCompile(`^/tickets/(\d+)/assign$`), http.MethodPost, h.AssignTicket},
		{regexp.MustCompile(`^/tickets/(\d+)/status$`), http.MethodPost, h.UpdateStatus},
		{regexp.MustCompile(`^/tickets/(\d+)/status$`), http.MethodPut, h.UpdateStatus},
		{regexp.MustCompile(`^/tickets/(\d+)/history$`), http.MethodGet, h.GetTicketHistory},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		for _, rt := range routes {
			if rt.method != method {
				continue
			}
			if rt.pattern.MatchString(path) {
				rt.handler(w, r)
				return
			}
		}

		http.NotFound(w, r)
	})

	port := getPort()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	sched := scheduler.NewScheduler(db)
	go sched.Run(ctx)

	go func() {
		log.Printf("server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	cancel()
	log.Println("shutdown complete")
}

func pathMatches(pattern, path string) bool {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, p := range patternParts {
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			continue
		}
		if p != pathParts[i] {
			return false
		}
	}
	return true
}
