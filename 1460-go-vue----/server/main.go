package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"supplier-portal/core"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "server port (overrides env PORT)")
	flag.Parse()

	if port == 0 {
		if envPort := os.Getenv("PORT"); envPort != "" {
			port = 8080
			fmt.Sscanf(envPort, "%d", &port)
		} else {
			port = 8080
		}
	}

	store := core.NewStore()
	server := NewServer(store)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/inquiries", server.CreateInquiryHandler)
	mux.HandleFunc("GET /api/inquiries", server.ListInquiriesHandler)
	mux.HandleFunc("GET /api/inquiry", server.GetInquiryHandler)
	mux.HandleFunc("PUT /api/inquiry/materials", server.UpdateInquiryMaterialsHandler)

	mux.HandleFunc("GET /api/todos", server.GetTodoRemindersHandler)

	mux.HandleFunc("POST /api/quotations", server.SubmitQuotationHandler)
	mux.HandleFunc("GET /api/quotations", server.GetQuotationsHandler)

	mux.HandleFunc("POST /api/comparison", server.GenerateComparisonHandler)

	mux.HandleFunc("POST /api/award", server.AwardAndCreateOrdersHandler)
	mux.HandleFunc("GET /api/orders", server.GetOrdersHandler)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Supplier Portal Server starting on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
