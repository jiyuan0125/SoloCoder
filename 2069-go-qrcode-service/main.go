package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"qrcode-service/internal/handlers"
	"qrcode-service/internal/storage"
)

func main() {
	db, err := storage.NewDB("qrcode.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	
	h := handlers.NewHandler(db)
	
	r := mux.NewRouter()
	h.RegisterRoutes(r)
	
	log.Println("QR Code service running on port 8080")
	if err := http.ListenAndServe(":9403", r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
