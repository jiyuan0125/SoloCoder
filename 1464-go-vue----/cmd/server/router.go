package main

import (
	"net/http"
)

func (h *Handler) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/devices", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.HandleCreateDevice(w, r)
		case http.MethodGet:
			h.HandleListDevices(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/devices/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.HandleGetDevice(w, r)
		case http.MethodPut, http.MethodPatch:
			h.HandleUpdateDevice(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/inspection/points", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.HandleCreatePoint(w, r)
		case http.MethodGet:
			h.HandleListPoints(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/inspection/routes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.HandleCreateRoute(w, r)
		case http.MethodGet:
			h.HandleListRoutes(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/inspection/plans", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.HandleCreatePlan(w, r)
		case http.MethodGet:
			h.HandleListPlans(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/inspection/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.HandleListTasks(w, r)
	})

	mux.HandleFunc("/api/inspection/tasks/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/inspection/tasks/check" {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.HandleCheckTaskPoint(w, r)
			return
		}

		path := r.URL.Path
		if len(path) > len("/api/inspection/tasks/") && path[len("/api/inspection/tasks/"):] == "check" {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.HandleCheckTaskPoint(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if len(path) > len("/api/inspection/tasks/") && path[len(path)-8:] == "/summary" {
				h.HandleTaskSummary(w, r)
			} else {
				h.HandleGetTask(w, r)
			}
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/inspection/tasks/check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.HandleCheckTaskPoint(w, r)
	})

	mux.HandleFunc("/api/drills/plans", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.HandleCreateDrillPlan(w, r)
		case http.MethodGet:
			h.HandleListDrillPlans(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/drills/plans/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > len("/api/drills/plans/") && path[len(path)-9:] == "/complete" {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.HandleCompleteDrill(w, r)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	})

	mux.HandleFunc("/api/drills/records", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.HandleListDrillRecords(w, r)
	})

	mux.HandleFunc("/api/reminders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.HandleListReminders(w, r)
	})

	mux.HandleFunc("/api/reminders/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > len("/api/reminders/") && path[len(path)-5:] == "/read" {
			if r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.HandleMarkReminderRead(w, r)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	})

	return mux
}
