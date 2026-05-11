package server

import (
	"net/http"
	"strings"
)

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"Contract Management API"}`))
			return
		}

		path := r.URL.Path
		parts := strings.Split(strings.Trim(path, "/"), "/")

		if len(parts) == 0 {
			http.NotFound(w, r)
			return
		}

		switch parts[0] {
		case "contracts":
			handleContracts(w, r, h, parts)
		default:
			http.NotFound(w, r)
		}
	})

	return mux
}

func handleContracts(w http.ResponseWriter, r *http.Request, h *Handler, parts []string) {
	if len(parts) == 1 {
		if r.Method == http.MethodPost {
			h.CreateContract(w, r)
			return
		}
		if r.Method == http.MethodGet {
			h.ListContracts(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(parts) == 2 {
		if r.Method == http.MethodGet {
			h.GetContract(w, r)
			return
		}
		if r.Method == http.MethodPut {
			h.UpdateContractAmount(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(parts) >= 3 {
		switch parts[2] {
		case "progress":
			h.GetProgress(w, r)
		case "audit-logs":
			h.GetAuditLogs(w, r)
		case "milestones":
			if len(parts) >= 4 {
				if len(parts) >= 5 && parts[4] == "pay" {
					h.PayMilestone(w, r)
					return
				}
				h.CompleteMilestone(w, r)
				return
			}
			http.Error(w, "invalid path", http.StatusBadRequest)
		default:
			http.NotFound(w, r)
		}
	}
}
