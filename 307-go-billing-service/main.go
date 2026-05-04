package main

import (
	"billing-service/handler"
	"billing-service/repository"
	"billing-service/service"
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	dbPath := flag.String("db", "./data/billing.db", "Path to SQLite database file")
	port := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	absDBPath, err := filepath.Abs(*dbPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	fmt.Printf("Initializing database at: %s\n", absDBPath)

	db, err := repository.InitDB(absDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	fmt.Println("Database initialized successfully")

	packageRepo := repository.NewPackageRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	usageRepo := repository.NewUsageRepository(db)
	billRepo := repository.NewBillRepository(db)
	configRepo := repository.NewConfigRepository(db)

	customerService := service.NewCustomerService(customerRepo, packageRepo)
	usageService := service.NewUsageService(usageRepo)
	billService := service.NewBillService(billRepo, customerRepo, packageRepo, usageRepo, configRepo)
	configService := service.NewConfigService(configRepo)

	router := handler.NewRouter(billService, customerService, usageService, configService)

	mux := http.NewServeMux()
	router.SetupRoutes(mux)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Starting server on %s\n", addr)
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /api/customers          - List all customers")
	fmt.Println("  POST /api/customers          - Create new customer")
	fmt.Println("  GET  /api/customers/:id      - Get customer by ID")
	fmt.Println("  POST /api/customers/:id/change-package - Change customer package")
	fmt.Println("  GET  /api/bills              - List all bills")
	fmt.Println("  GET  /api/bills?customer_id=:id - List customer bills")
	fmt.Println("  POST /api/bills              - Generate monthly bill")
	fmt.Println("  GET  /api/bills/:id          - Get bill by ID")
	fmt.Println("  GET  /api/bills/:id/detail   - Get bill detail")
	fmt.Println("  GET  /api/bills/:id/payments - Get bill payment history")
	fmt.Println("  POST /api/bills/:id/mark-paid - Mark bill as paid")
	fmt.Println("  GET  /api/usages?customer_id=:id&year=:year&month=:month - Get usage")
	fmt.Println("  POST /api/usages?action=sms  - Record SMS usage")
	fmt.Println("  POST /api/usages?action=storage - Record storage usage")
	fmt.Println("  GET  /api/configs/:key       - Get config by key")
	fmt.Println("  PUT  /api/configs/:key       - Update config")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
