package main

import (
	"audit-log/server/handler"
	"audit-log/server/model"
	"fmt"
	"log"
	"net/http"
)

func main() {
	storage := model.NewInMemoryStorage()
	auditHandler := handler.NewAuditHandler(storage)

	http.HandleFunc("/health", auditHandler.HealthCheck)

	http.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			auditHandler.CreateLog(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/logs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			auditHandler.GetLog(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/logs/query", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			auditHandler.QueryLogs(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			auditHandler.Login(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/user/lock-status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			auditHandler.GetUserLockStatus(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/export/approval", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			auditHandler.RequestExportApproval(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/archive/trigger", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			auditHandler.TriggerArchive(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/archive/list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			auditHandler.GetArchivedList(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/storage/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			auditHandler.GetStorageStatus(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/statistics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			auditHandler.GetStatistics(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := ":8080"
	fmt.Printf("Audit Log Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
