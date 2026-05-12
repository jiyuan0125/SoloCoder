package main

import (
	"credit-system/internal/handler"
	"credit-system/internal/store"
	"flag"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "Server port")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	s := store.New()
	h := handler.New(s)
	h.RegisterRoutes(r)

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			h.CheckCourseStartReminders()
		}
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		h.CheckCourseStartReminders()
	}()

	r.Run(":" + port)
}
