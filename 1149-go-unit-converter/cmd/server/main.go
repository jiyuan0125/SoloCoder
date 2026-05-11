package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/convert", ConvertHandler)

	fmt.Printf("Unit Converter Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
