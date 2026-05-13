package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"log-rotate/internal/config"
	"log-rotate/internal/handler"
	"log-rotate/internal/logger"
	"log-rotate/internal/repository"
)

func main() {
	repo, err := repository.NewRepository(config.DefaultDBPath)
	if err != nil {
		log.Fatalf("init repository failed: %v", err)
	}
	defer repo.Close()

	writer, err := logger.NewLogWriter(repo)
	if err != nil {
		log.Fatalf("init log writer failed: %v", err)
	}
	defer writer.Close()

	querier := logger.NewLogQuerier(repo, writer.GetConfig().LogDir)

	h := handler.NewHandler(repo, writer, querier)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	h.SetupRoutes(r)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("log rotate server starting on port %s", config.ServerPort)
		if err := r.Run(config.ServerPort); err != nil {
			log.Printf("server run error: %v", err)
		}
	}()

	<-quit
	log.Println("server shutting down...")
}
