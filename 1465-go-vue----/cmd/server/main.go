package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"envmonitor/internal/core"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "server port")
	flag.Parse()

	if port == 0 {
		envPort := os.Getenv("ENV_MONITOR_PORT")
		if envPort != "" {
			if p, err := strconv.Atoi(envPort); err == nil {
				port = p
			}
		}
	}

	if port == 0 {
		port = 8080
	}

	service := core.NewService()
	handler := NewHandler(service)

	fmt.Printf("Starting server on port %d...\n", port)
	if err := handler.Start(fmt.Sprintf(":%d", port)); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
