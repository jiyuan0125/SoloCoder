package main

import (
	"excel-report-system/database"
	"excel-report-system/handlers"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	workDir, _ := os.Getwd()
	dbPath := filepath.Join(workDir, "reports.db")
	tmplDir := filepath.Join(workDir, "templates")
	exportDir := filepath.Join(workDir, "exports")

	if err := os.MkdirAll(tmplDir, 0755); err != nil {
		log.Fatal("Failed to create templates directory:", err)
	}
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		log.Fatal("Failed to create exports directory:", err)
	}

	db, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	handler := handlers.NewHandler(db, tmplDir, exportDir)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/reports", handler.ListReports)
	mux.HandleFunc("/api/reports/", handler.ReportDetail)
	mux.HandleFunc("/api/reports/workflow", handler.WorkflowAction)

	mux.HandleFunc("/api/templates", handler.ListTemplates)
	mux.HandleFunc("/api/templates/", handler.TemplateDetail)

	mux.HandleFunc("/api/entities", handler.ListEntities)
	mux.HandleFunc("/api/entities/", handler.EntityDetail)

	mux.HandleFunc("/api/export/", handler.ExportReport)

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":9401", mux))
}
