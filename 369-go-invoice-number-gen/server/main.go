package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"invoice-number-gen/invoicegen"
	"invoice-number-gen/protocol"
)

var (
	generator *invoicegen.Generator
)

func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req protocol.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := protocol.GenerateResponse{
			Error: "invalid request body",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	invoiceNumber, err := generator.Generate(req.Prefix)
	if err != nil {
		resp := protocol.GenerateResponse{
			Error: err.Error(),
		}
		if err == invoicegen.ErrSequenceExceeded {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := protocol.GenerateResponse{
		InvoiceNumber: invoiceNumber,
	}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP server address")
	storagePath := flag.String("storage", "./invoice_state.json", "Path to storage file")
	flag.Parse()

	var err error
	generator, err = invoicegen.NewGenerator(*storagePath)
	if err != nil {
		log.Fatalf("Failed to initialize generator: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		generator.Sync()
		os.Exit(0)
	}()

	http.HandleFunc("/generate", generateHandler)

	log.Printf("Server starting on %s", *addr)
	log.Printf("Storage file: %s", *storagePath)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
