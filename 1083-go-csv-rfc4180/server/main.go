package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/parse", handleParse)
	mux.HandleFunc("/serialize", handleSerialize)
	mux.HandleFunc("/validate", handleValidate)

	addr := ":" + port
	fmt.Printf("CSV RFC 4180 server starting on %s\n", addr)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /parse     - Parse CSV text")
	fmt.Println("  POST /serialize - Serialize data to CSV")
	fmt.Println("  POST /validate  - Validate CSV format")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
