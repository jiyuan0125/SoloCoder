package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flashsale/internal/config"
	"flashsale/internal/db"
	"flashsale/internal/repository"
	"flashsale/internal/service"
	"flashsale/internal/handler"
	"flashsale/internal/scheduler"
)

func main() {
	cfg := config.Load()

	database, err := db.Init(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	activityRepo := repository.NewActivityRepository(database)
	orderRepo := repository.NewOrderRepository(database)
	stockRepo := repository.NewStockRepository(database)
	reportRepo := repository.NewReportRepository(database)
	cacheRepo := repository.NewCacheRepository(database)

	activityService := service.NewActivityService(activityRepo, stockRepo, orderRepo)
	orderService := service.NewOrderService(orderRepo, stockRepo, activityRepo)
	reportService := service.NewReportService(reportRepo, orderRepo, activityRepo)
	cacheService := service.NewCacheService(cacheRepo, activityRepo, stockRepo, reportRepo)

	activityHandler := handler.NewActivityHandler(activityService)
	orderHandler := handler.NewOrderHandler(orderService)
	reportHandler := handler.NewReportHandler(reportService)
	cacheHandler := handler.NewCacheHandler(cacheService)

	mux := http.NewServeMux()
	activityHandler.RegisterRoutes(mux)
	orderHandler.RegisterRoutes(mux)
	reportHandler.RegisterRoutes(mux)
	cacheHandler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: mux,
	}

	sched := scheduler.NewScheduler(orderService, reportService)
	sched.Start()
	defer sched.Stop()

	go func() {
		log.Printf("Server starting on %s", cfg.ServerAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
