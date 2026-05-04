package main

import (
	"billing/internal/server/handler"
	"billing/internal/server/model"
	"billing/internal/server/service"
	"billing/internal/server/store"
	"billing/pkg/api"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	port     = flag.Int("port", 8080, "HTTP server port")
	dataDir  = flag.String("data", "./data", "Data directory for persistence")
)

func main() {
	flag.Parse()

	log.Printf("Starting billing server on port %d...", *port)
	log.Printf("Data directory: %s", *dataDir)

	dataStore := model.NewDataStore()

	persistentStore, err := store.NewPersistentStore(*dataDir)
	if err != nil {
		log.Printf("Warning: failed to create persistent store: %v", err)
	} else {
		storedData, err := persistentStore.Load()
		if err != nil {
			log.Printf("Warning: failed to load existing data: %v", err)
		} else {
			dataStore.FromStoredData(storedData)
			log.Println("Loaded existing data from disk")
		}
	}

	_ = initializeDefaultPlans(dataStore, persistentStore)

	billingService := service.NewBillingService(dataStore, persistentStore)
	billingHandler := handler.NewBillingHandler(billingService)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/plans", billingHandler.ListPlans)
	mux.HandleFunc("POST /api/plans", billingHandler.CreatePlan)
	mux.HandleFunc("GET /api/plans/{id}", billingHandler.GetPlan)

	mux.HandleFunc("GET /api/customers", billingHandler.ListCustomers)
	mux.HandleFunc("POST /api/customers", billingHandler.CreateCustomer)
	mux.HandleFunc("GET /api/customers/{id}", billingHandler.GetCustomer)

	mux.HandleFunc("POST /api/plan-changes", billingHandler.ChangePlan)

	mux.HandleFunc("POST /api/usage", billingHandler.RecordUsage)
	mux.HandleFunc("GET /api/usage/{customer_id}", billingHandler.GetUsage)

	mux.HandleFunc("POST /api/bills/generate", billingHandler.GenerateBill)
	mux.HandleFunc("GET /api/bills", billingHandler.ListAllBills)
	mux.HandleFunc("GET /api/bills/{id}", billingHandler.GetBill)
	mux.HandleFunc("GET /api/customers/{customer_id}/bills", billingHandler.ListCustomerBills)
	mux.HandleFunc("POST /api/bills/mark-paid", billingHandler.MarkPaid)

	mux.HandleFunc("GET /api/pricing", billingHandler.GetPricing)
	mux.HandleFunc("PUT /api/pricing", billingHandler.UpdatePricing)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	log.Printf("Billing server is running on http://localhost:%d", *port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")

	if persistentStore != nil {
		data := dataStore.ToStoredData()
		if err := persistentStore.Save(data); err != nil {
			log.Printf("Warning: failed to save data: %v", err)
		} else {
			log.Println("Data saved successfully")
		}
	}
}

func initializeDefaultPlans(dataStore *model.DataStore, persistentStore *store.PersistentStore) error {
	existingPlans := dataStore.ListPlans()
	if len(existingPlans) > 0 {
		return nil
	}

	defaultPlans := []*struct {
		Name           string
		MonthlyFee     float64
		SmsQuota       int
		StorageQuotaGB float64
		Description    string
	}{
		{"基础版", 99.0, 1000, 10.0, "适合小型企业，包含基础功能"},
		{"专业版", 299.0, 5000, 50.0, "适合成长型企业，更多资源配额"},
		{"企业版", 899.0, 50000, 500.0, "适合大型企业，高级功能支持"},
	}

	for _, p := range defaultPlans {
		plan := &api.Plan{
			Name:           p.Name,
			MonthlyFee:     p.MonthlyFee,
			SmsQuota:       p.SmsQuota,
			StorageQuotaGB: p.StorageQuotaGB,
			Description:    p.Description,
		}
		if _, err := dataStore.CreatePlan(plan); err != nil {
			log.Printf("Warning: failed to create default plan %s: %v", p.Name, err)
		}
	}

	log.Println("Initialized default plans")

	if persistentStore != nil {
		data := dataStore.ToStoredData()
		if err := persistentStore.Save(data); err != nil {
			log.Printf("Warning: failed to save default plans: %v", err)
		}
	}

	return nil
}
