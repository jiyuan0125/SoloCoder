package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP server address")
	flag.Parse()

	fmt.Printf("Starting age calculator server on %s\n", *addr)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /api/age         - Calculate age and related information")
	fmt.Println("  POST /api/days-between - Calculate days between two dates")

	srv := NewServer(*addr)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
