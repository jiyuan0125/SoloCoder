package main

import (
	"flag"
	"log"
)

func main() {
	host := flag.String("host", "127.0.0.1", "Server host")
	port := flag.Int("port", 8080, "Server port")
	flag.Parse()

	srv := NewServer(*host, *port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
