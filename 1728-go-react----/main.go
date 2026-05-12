package main

import (
	"flag"
	"fmt"
	"lab-safety/api"
	"lab-safety/config"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	var port string

	flag.StringVar(&port, "port", "", "Server port (e.g., 8080)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	config.InitStorage()
	config.StartExpiryChecker()

	router := gin.Default()

	api.SetupRoutes(router)

	fmt.Printf("Server starting on port %s\n", port)
	router.Run(":" + port)
}
