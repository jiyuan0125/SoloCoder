package main

import (
	"flag"
	"log"
	"os"
	"telemedicine/config"
	"telemedicine/router"
)

func main() {
	port := flag.String("port", "", "Server port")
	flag.Parse()

	if *port == "" {
		if envPort := os.Getenv("PORT"); envPort != "" {
			*port = envPort
		} else {
			*port = "8080"
		}
	}

	config.InitDB()
	r := router.SetupRouter()

	log.Printf("Server starting on port %s...", *port)
	if err := r.Run(":" + *port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
