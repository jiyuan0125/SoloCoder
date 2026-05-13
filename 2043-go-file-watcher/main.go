package main

import (
	"file-watcher/handlers"
	"file-watcher/repository"
	"file-watcher/watcher"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := repository.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	mgr := watcher.GetManager()
	if err := mgr.LoadExistingWatches(); err != nil {
		log.Printf("Warning: failed to load existing watches: %v", err)
	}

	router := gin.Default()

	router.POST("/api/watches", handlers.AddWatch)
	router.DELETE("/api/watches/:id", handlers.RemoveWatch)
	router.GET("/api/watches", handlers.ListWatches)
	router.GET("/api/watches/:id", handlers.GetWatch)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8101"
	}
	
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	mgr.Stop()
	log.Println("Server stopped")
}
