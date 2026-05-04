package main

import (
	"flag"
	"log"
	"net/http"

	"referral-service/internal/server/generator"
	"referral-service/internal/server/handler"
	"referral-service/internal/server/service"
	"referral-service/internal/server/store"
	"referral-service/pkg/api"
)

func main() {
	port := flag.String("port", "8080", "HTTP server port")
	dataFile := flag.String("data", "referral_data.json", "Data file path for persistence")
	flag.Parse()

	s, err := store.NewStore(*dataFile)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	g := generator.NewGenerator()
	svc := service.NewService(s, g)
	h := handler.NewHandler(svc)

	http.HandleFunc(api.EndpointGenerateReferralCode, h.GenerateReferralCode)
	http.HandleFunc(api.EndpointBindReferral, h.BindReferral)
	http.HandleFunc(api.EndpointCompleteFirstOrder, h.CompleteFirstOrder)
	http.HandleFunc(api.EndpointRefundFirstOrder, h.RefundFirstOrder)
	http.HandleFunc(api.EndpointGetUserStats, h.GetUserStats)
	http.HandleFunc(api.EndpointSetRewardPoints, h.HandleRewardPoints)
	http.HandleFunc(api.EndpointGetStats, h.GetStats)

	log.Printf("Server starting on port %s...", *port)
	log.Printf("Data file: %s", *dataFile)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
