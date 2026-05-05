package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"data-export/server/internal/data"
	"data-export/server/internal/handler"
	"data-export/server/internal/stats"
	"data-export/server/internal/task"
	"data-export/server/internal/template"
)

var (
	port = flag.Int("port", 8080, "HTTP server port")
)

func main() {
	flag.Parse()

	templates := template.NewTemplateManager()
	statsCollector := stats.NewStatsCollector()
	dataSource := data.NewDataSource()
	scheduler := task.NewTaskScheduler(templates, dataSource, statsCollector)

	h := handler.NewHandler(templates, scheduler, statsCollector, dataSource)

	mux := http.NewServeMux()
	mux.Handle("/api/", h)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Printf("Data export server starting on port %d...\n", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	<-stop
	fmt.Println("\nShutting down server...")

	scheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server shutdown error: %v\n", err)
	}

	fmt.Println("Server stopped successfully")
}
