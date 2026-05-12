package main

import (
	"flag"
	"fmt"
	"os"
	"research-collaboration/src/routes"
	"research-collaboration/src/storage"
)

func main() {
	defaultPort := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		defaultPort = envPort
	}

	port := flag.String("port", defaultPort, "Server port")
	flag.Parse()

	store := storage.NewStorage()
	router := routes.SetupRouter(store)

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Server starting on %s\n", addr)
	router.Run(addr)
}
