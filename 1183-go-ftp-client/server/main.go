package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := flag.String("port", "", "Port to listen on")
	flag.Parse()

	listenPort := getPort(*port)
	if listenPort == "" {
		log.Fatal("No port specified")
	}

	tm := NewTransferManager()
	handler := NewHandler(tm)

	http.HandleFunc("/api/config", handler.HandleConfig)
	http.HandleFunc("/api/operation", handler.HandleOperation)
	http.HandleFunc("/api/queue", handler.HandleQueue)
	http.HandleFunc("/api/cancel", handler.HandleCancel)

	addr := ":" + listenPort
	fmt.Printf("Server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func getPort(argPort string) string {
	if argPort != "" {
		return argPort
	}

	if envPort := os.Getenv("FTP_SERVER_PORT"); envPort != "" {
		return envPort
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}

	return "8202"
}
