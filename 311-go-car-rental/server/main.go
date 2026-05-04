package main

import (
	"log"
	"net/http"

	"carrental/server/handler"
	"carrental/server/service"
	"carrental/server/storage"
)

func main() {
	dataFile := "carrental_data.json"

	store, err := storage.NewJSONStorage(dataFile)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	carService := service.NewCarService(store)
	orderService := service.NewOrderService(store, carService)

	h := handler.NewHandler(carService, orderService)

	http.HandleFunc("/cars", h.HandleCars)
	http.HandleFunc("/cars/status", h.HandleUpdateCarStatus)
	http.HandleFunc("/cars/availability", h.HandleCheckAvailability)
	http.HandleFunc("/orders", h.HandleOrders)
	http.HandleFunc("/orders/return", h.HandleReturnCar)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
