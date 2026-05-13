package main

import (
	"log"
	"net/http"

	"filedownload/internal/handler"
	"filedownload/internal/service"
	"filedownload/internal/store"
)

func main() {
	db, err := store.NewSQLiteStore("./data/files.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	storage, err := service.NewStorageService("./data/storage")
	if err != nil {
		log.Fatalf("failed to create storage: %v", err)
	}

	fileService := service.NewFileService(db, storage)
	h := handler.NewHandler(fileService)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/upload", h.Upload)
	mux.HandleFunc("/api/download/", h.DownloadByID)
	mux.HandleFunc("/api/files/", h.DownloadByPath)
	mux.HandleFunc("/api/files", h.ListFiles)
	mux.HandleFunc("/api/admin/files", h.AdminListAllFiles)
	mux.HandleFunc("/api/admin/downloads", h.AdminListDownloadRecords)

	log.Println("server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
