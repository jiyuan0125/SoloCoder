package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	customerStore := &FileCustomerStore{filePath: dataDir + "/customers.json"}
	if err := customerStore.Load(); err != nil {
		log.Printf("Warning: Failed to load customers: %v", err)
	}

	usageStore := &FileUsageStore{filePath: dataDir + "/usage.json"}
	if err := usageStore.Load(); err != nil {
		log.Printf("Warning: Failed to load usage: %v", err)
	}

	billingStore := &FileBillingStore{filePath: dataDir + "/billing.json"}
	if err := billingStore.Load(); err != nil {
		log.Printf("Warning: Failed to load billing: %v", err)
	}

	customerService := NewCustomerService(customerStore)
	usageService := NewUsageService(usageStore, customerStore)
	billingService := NewBillingService(billingStore, customerStore, usageStore)

	http.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			customerService.CreateCustomer(w, r)
		case http.MethodGet:
			customerService.ListCustomers(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/customers/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			customerService.GetCustomer(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/customers/change-plan", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			customerService.ChangePlan(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/usage/sms", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			usageService.RecordSMSUsage(w, r)
		case http.MethodGet:
			usageService.GetSMSUsage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/usage/storage", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			usageService.RecordStorageUsage(w, r)
		case http.MethodGet:
			usageService.GetStorageUsage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/usage/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			usageService.UpdatePricingConfig(w, r)
		case http.MethodGet:
			usageService.GetPricingConfig(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/bills/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			billingService.GenerateMonthlyBills(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/bills", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			billingService.ListAllBills(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/bills/customer/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			billingService.ListCustomerBills(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/bills/pay", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			billingService.MarkBillAsPaid(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := ":8080"
	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
