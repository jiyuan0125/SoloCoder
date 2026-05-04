package main

import (
	"batch-processor/batch"
	"flag"
	"log"
	"net/http"
	"time"
)

type SimpleHandler struct {
}

func (h *SimpleHandler) ProcessBatch(items []string) error {
	log.Printf("Processing batch of %d items: %v", len(items), items)
	return nil
}

func main() {
	port := flag.String("port", "8080", "HTTP server port")
	batchSize := flag.Int("batch-size", 10, "Batch size for processing")
	timeout := flag.Duration("timeout", 5*time.Second, "Timeout for batch processing")
	closeTimeout := flag.Duration("close-timeout", 10*time.Second, "Timeout for Close()")
	flag.Parse()

	handler := &SimpleHandler{}

	processor := batch.NewBatchProcessor[string](handler, batch.BatchProcessorConfig{
		BatchSize:    *batchSize,
		Timeout:      *timeout,
		CloseTimeout: *closeTimeout,
	})

	serverHandler := NewServerHandler(processor)

	mux := http.NewServeMux()
	mux.HandleFunc("/submit", serverHandler.SubmitHandler)
	mux.HandleFunc("/flush", serverHandler.FlushHandler)
	mux.HandleFunc("/stats", serverHandler.StatsHandler)
	mux.HandleFunc("/shutdown", serverHandler.ShutdownHandler)

	addr := ":" + *port
	log.Printf("Server starting on %s", addr)
	log.Printf("Configuration: batch-size=%d, timeout=%v, close-timeout=%v", *batchSize, *timeout, *closeTimeout)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
