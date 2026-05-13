package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"poolapp/api"
	"poolapp/pool"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	manager := pool.NewManager()
	registry := pool.NewFactoryRegistry()
	handler := api.NewHandler(manager, registry)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/pools", handler.GetPools)
	r.POST("/pools", handler.CreatePool)
	r.GET("/pools/:name/stats", handler.GetPoolStats)
	r.GET("/factory-types", handler.GetFactoryTypes)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s...", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %s", err)
	}

	log.Println("Closing all connection pools...")
	manager.CloseAll()

	log.Println("Server exiting")
}
