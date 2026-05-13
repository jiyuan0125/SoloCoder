package router

import (
	"net/http"
	"release-mgmt/internal/handler"
	"strings"
)

func matchesPattern(path, pattern string) bool {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")

	if len(pathParts) != len(patternParts) {
		return false
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, "{") {
			continue
		}
		if part != pathParts[i] {
			return false
		}
	}
	return true
}

func New() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.Health)

	mux.HandleFunc("/api/releases", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.ListReleases(w, r)
		} else if r.Method == http.MethodPost {
			handler.CreateRelease(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/releases/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch {
		case matchesPattern(path, "/api/releases/{version}"):
			if r.Method == http.MethodGet && !strings.Contains(strings.TrimPrefix(path, "/api/releases/"), "/") {
				handler.GetRelease(w, r)
				return
			}
			http.NotFound(w, r)

		case matchesPattern(path, "/api/releases/{id}/submit"):
			handler.SubmitForReview(w, r)

		case matchesPattern(path, "/api/releases/{id}/changes"):
			if r.Method == http.MethodGet {
				handler.GetChanges(w, r)
			} else if r.Method == http.MethodPost {
				handler.AddChange(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}

		case matchesPattern(path, "/api/releases/{id}/approvals"):
			if r.Method == http.MethodGet {
				handler.GetApprovals(w, r)
			} else if r.Method == http.MethodPost {
				handler.Approve(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}

		case matchesPattern(path, "/api/releases/{id}/deploy/staging"):
			handler.DeployToStaging(w, r)

		case matchesPattern(path, "/api/releases/{id}/deploy/production"):
			handler.DeployToProduction(w, r)

		case matchesPattern(path, "/api/releases/{id}/rollback"):
			handler.RollbackProduction(w, r)

		case matchesPattern(path, "/api/releases/{id}/history"):
			handler.GetOperationHistory(w, r)

		case matchesPattern(path, "/api/releases/{id}/history/{hid}/notes"):
			handler.AddNoteToHistory(w, r)

		default:
			http.NotFound(w, r)
		}
	})

	return mux
}
