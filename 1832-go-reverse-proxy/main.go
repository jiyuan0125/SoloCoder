package main

import (
	"os"

	"github.com/gin-gonic/gin"

	"reverse-proxy/api"
	"reverse-proxy/healthcheck"
	"reverse-proxy/proxy"
	"reverse-proxy/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := store.New()

	hc := healthcheck.New(s)
	go hc.Start()
	defer hc.Stop()

	p := proxy.New(s)
	handlers := api.New(s)

	r := gin.Default()

	admin := r.Group("/")
	{
		admin.GET("/backends", handlers.GetBackends)
		admin.POST("/backends", handlers.AddBackend)
		admin.DELETE("/backends/:id", handlers.DeleteBackend)
	}

	r.NoRoute(p.Handler())

	r.Run(":" + port)
}
