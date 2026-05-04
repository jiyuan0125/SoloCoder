package main

import (
	"fmt"
	"net/http"
)

func main() {
	storage := NewStorage("insurance_data.json")
	service := NewService(storage)
	handler := NewHandler(service)

	fmt.Println("Insurance Claim Server starting on port 8080...")
	fmt.Println("Endpoints:")
	fmt.Println("  POST /policies/create    - Create new policy")
	fmt.Println("  GET  /policies           - List all policies")
	fmt.Println("  GET  /policies/{number}/claims - List claims for a policy")
	fmt.Println("  POST /claims/submit      - Submit a claim")
	fmt.Println("  GET  /claims             - List all claims")
	fmt.Println("  GET  /claims/{id}        - Get claim by ID")
	fmt.Println("  GET  /claims/pending     - List pending claims (admin)")
	fmt.Println("  GET  /claims/approved    - List approved claims (admin)")
	fmt.Println("  POST /claims/review      - Review a claim (admin)")
	fmt.Println("  POST /claims/pay         - Confirm payment (admin)")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
