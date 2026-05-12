package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gateway/pkg/api"
	"gateway/pkg/router"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	mgr := router.NewManager()

	api.RegisterAdmin(r, mgr)

	r.NoRoute(mgr.ProxyHandler())

	log.Printf("gateway starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
