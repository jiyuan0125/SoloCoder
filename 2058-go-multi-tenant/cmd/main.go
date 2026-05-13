package main

import (
	"log"
	"multitenant/internal/database"
	"multitenant/internal/server"
	"net/http"
	"os"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./multitenant.db"
	}

	db, err := database.NewDB(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := database.InitSchema(db); err != nil {
		log.Fatalf("failed to init schema: %v", err)
	}

	mux := http.NewServeMux()
	handler := server.NewHandler(db)
	handler.RegisterRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("server starting on %s", addr)
	log.Printf("database: %s", dbPath)
	log.Fatal(http.ListenAndServe(addr, mux))
}
