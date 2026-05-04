package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"ordergen/ordergen"
)

var (
	port        = flag.String("port", "8080", "HTTP server port")
	storagePath = flag.String("storage", "./ordergen_state.json", "Path to storage file")
	prefix      = flag.String("prefix", "", "Default business prefix")
)

func main() {
	flag.Parse()

	config := ordergen.OrderGeneratorConfig{
		Prefix:      *prefix,
		StoragePath: *storagePath,
	}

	generator, err := ordergen.NewOrderGenerator(config)
	if err != nil {
		log.Fatalf("Failed to create order generator: %v", err)
	}

	handler := NewHandler(generator)

	mux := http.NewServeMux()
	mux.HandleFunc("/generate", handler.Generate)
	mux.HandleFunc("/parse", handler.Parse)
	mux.HandleFunc("/prefix", handler.Prefix)

	log.Printf("Server starting on port %s...", *port)
	log.Printf("Storage path: %s", *storagePath)
	if *prefix != "" {
		log.Printf("Default prefix: %s", *prefix)
	}

	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
		os.Exit(1)
	}
}
