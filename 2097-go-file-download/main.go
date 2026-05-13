package main

import (
	"log"
	"net/http"
	"os"

	"filedownload/internal/handlers"
	"filedownload/internal/middleware"
	"filedownload/internal/storage"
)

func main() {
	dbPath := "./data/files.db"
	s, err := storage.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer s.Close()
	storage.GlobalStorage = s

	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/upload", middleware.Auth(handlers.UploadFile))
	mux.HandleFunc("POST /api/generate-link", middleware.Auth(handlers.GenerateDownloadLink))
	mux.HandleFunc("GET /api/download/{linkID}", handlers.DownloadFile)
	mux.HandleFunc("GET /api/files", middleware.Auth(handlers.GetUserFiles))
	mux.HandleFunc("GET /api/files/{fileID}", middleware.Auth(handlers.GetFileInfo))
	mux.HandleFunc("GET /api/files/path/{path...}", middleware.Auth(handlers.DownloadByPath))

	mux.HandleFunc("GET /api/admin/files", middleware.Auth(middleware.RequireAdmin(handlers.AdminGetAllFiles)))
	mux.HandleFunc("GET /api/admin/download-records", middleware.Auth(middleware.RequireAdmin(handlers.AdminGetAllDownloadRecords)))
	mux.HandleFunc("GET /api/admin/users/{userID}/files", middleware.Auth(middleware.RequireAdmin(handlers.AdminGetUserFiles)))
	mux.HandleFunc("GET /api/admin/users/{userID}/download-records", middleware.Auth(middleware.RequireAdmin(handlers.AdminGetUserDownloadRecords)))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Server starting on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
