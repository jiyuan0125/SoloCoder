package main

import (
	"flag"
	"net/http"
	"os"
	"parking-system/core"
)

const defaultPort = "8904"

func main() {
	port := getPort()

	service := core.NewParkingService()
	handler := NewHandler(service)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/spots", handler.CreateSpot)
	mux.HandleFunc("GET /api/spots", handler.ListSpots)
	mux.HandleFunc("PUT /api/spots/{id}/status", handler.UpdateSpotStatus)

	mux.HandleFunc("GET /api/guidance", handler.GetGuidance)

	mux.HandleFunc("POST /api/parking/checkin", handler.CheckIn)
	mux.HandleFunc("POST /api/parking/checkout", handler.CheckOut)
	mux.HandleFunc("GET /api/parking/fee", handler.QueryFee)
	mux.HandleFunc("GET /api/parking/records", handler.ListParkingRecords)

	mux.HandleFunc("POST /api/cards", handler.CreateMonthlyCard)
	mux.HandleFunc("POST /api/cards/renew", handler.RenewMonthlyCard)
	mux.HandleFunc("GET /api/cards", handler.ListMonthlyCards)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func getPort() string {
	var port string

	flag.StringVar(&port, "port", "", "Server port")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
	}

	if port == "" {
		port = defaultPort
	}

	return port
}
