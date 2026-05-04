package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	port := flag.Int("port", 8080, "Server port")
	flag.Parse()

	server := NewServer(*port)
	
	fmt.Printf("Barcode encoder server starting on port %d...\n", *port)
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
