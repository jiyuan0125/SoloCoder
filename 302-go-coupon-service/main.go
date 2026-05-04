package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	storage := NewStorage()
	if err := storage.Load(); err != nil {
		log.Fatalf("Failed to load storage: %v", err)
	}
	log.Println("Storage loaded successfully")

	batchHandler := NewBatchHandler(storage)
	claimHandler := NewClaimHandler(storage)
	redeemHandler := NewRedeemHandler(storage)

	mux := http.NewServeMux()

	// Batch management routes
	mux.HandleFunc("/batches", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			batchHandler.CreateBatch(w, r)
		} else if r.Method == http.MethodGet {
			batchHandler.ListBatches(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/batches/detail", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			batchHandler.GetBatch(w, r)
		} else if r.Method == http.MethodPut {
			batchHandler.UpdateBatch(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/batches/progress", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			batchHandler.GetBatchProgress(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Claim management routes
	mux.HandleFunc("/claims", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			claimHandler.ClaimCoupon(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/claims/user", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			claimHandler.GetUserCoupons(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/claims/return", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			claimHandler.ReturnCoupon(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Redeem management routes
	mux.HandleFunc("/redeems", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			redeemHandler.RedeemCoupon(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/redeems/detail", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			redeemHandler.GetRedeemRecord(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/coupons/validate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			redeemHandler.ValidateCoupon(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		if err := storage.Save(); err != nil {
			log.Printf("Failed to save storage: %v", err)
		} else {
			log.Println("Storage saved successfully")
		}
		os.Exit(0)
	}()

	addr := ":8080"
	log.Printf("Server starting on %s...", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
