package main

import (
	"encoding/json"
	"net/http"

	"supervision-log-system/common"
	"supervision-log-system/core"
)

type Router struct {
	store               *Store
	supervisoryService  *core.SupervisoryService
	acceptanceService   *core.AcceptanceService
	issueService        *core.IssueService
	dailyLogService     *core.DailyLogService
}

type Store struct {
	*core.Store
}

func NewRouter(store *core.Store) *Router {
	return &Router{
		store:              &Store{Store: store},
		supervisoryService: core.NewSupervisoryService(store),
		acceptanceService:  core.NewAcceptanceService(store),
		issueService:       core.NewIssueService(store),
		dailyLogService:    core.NewDailyLogService(
			store,
			core.NewSupervisoryService(store),
			core.NewAcceptanceService(store),
		),
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch req.URL.Path {
	case "/api/supervisory":
		if req.Method == http.MethodPost {
			r.createSupervisoryRecord(w, req)
		} else if req.Method == http.MethodGet {
			r.listSupervisoryRecords(w)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/api/supervisory/submit":
		if req.Method == http.MethodPost {
			r.submitSupervisoryRecord(w, req)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/api/supervisory/supplement":
		if req.Method == http.MethodPost {
			r.addSupervisorySupplement(w, req)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/api/acceptance":
		if req.Method == http.MethodPost {
			r.createAcceptanceRecord(w, req)
		} else if req.Method == http.MethodGet {
			r.listAcceptanceRecords(w)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/api/acceptance/rectification":
		if req.Method == http.MethodPost {
			r.createRectificationNotice(w, req)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/api/acceptance/rectification/complete":
		if req.Method == http.MethodPost {
			r.completeRectification(w, req)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/api/acceptance/recheck":
		if req.Method == http.MethodPost {
			r.recheckAcceptance(w, req)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/api/issues":
		if req.Method == http.MethodPost {
			r.createIssue(w, req)
		} else if req.Method == http.MethodGet {
			r.listIssues(w)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/api/issues/status":
		if req.Method == http.MethodPut {
			r.updateIssueStatus(w, req)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/api/issues/overdue":
		if req.Method == http.MethodPost {
			r.checkOverdueIssues(w)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/api/dailylog":
		if req.Method == http.MethodPost {
			r.submitDailyLog(w, req)
		} else if req.Method == http.MethodGet {
			r.getDailyLog(w, req)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/api/dailylog/draft":
		if req.Method == http.MethodPost {
			r.generateDailyLogDraft(w, req)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func sendError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.ErrorResponse{Error: message})
}

func sendSuccess(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.SuccessResponse{
		Success: true,
		Data:    data,
	})
}
