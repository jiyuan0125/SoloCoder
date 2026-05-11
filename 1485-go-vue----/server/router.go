package main

import (
	"net/http"
	"strings"
)

func NewRouter(h *Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateTask(w, r)
		case http.MethodGet:
			h.ListTasks(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/tasks/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch {
		case strings.HasSuffix(path, "/end"):
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.EndTask(w, r)

		case strings.HasSuffix(path, "/data"):
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.GetSensorData(w, r)

		case strings.HasSuffix(path, "/alerts"):
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.GetAlerts(w, r)

		case strings.HasSuffix(path, "/stats"):
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.GetStats(w, r)

		case strings.HasSuffix(path, "/report"):
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.GetReport(w, r)

		case strings.HasSuffix(path, "/export/sensor"),
			strings.HasSuffix(path, "/export/stats"),
			strings.HasSuffix(path, "/export/alerts"):
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.ExportCSV(w, r)

		default:
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.GetTask(w, r)
		}
	})

	mux.HandleFunc("/api/sensor-data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.SubmitSensorData(w, r)
	})

	return mux
}
