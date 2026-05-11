package server

import (
	"net/http"
	"strings"
)

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	switch {
	case path == "/users" && method == http.MethodPost:
		h.CreateUser(w, r)
	case path == "/users" && method == http.MethodGet:
		h.ListUsers(w, r)
	case strings.HasPrefix(path, "/users/") && method == http.MethodGet:
		h.GetUser(w, r)

	case path == "/repairs" && method == http.MethodPost:
		h.CreateRepair(w, r)
	case path == "/repairs" && method == http.MethodGet:
		h.ListRepairs(w, r)
	case strings.HasPrefix(path, "/repairs/"):
		id := getIDFromPath(path, "/repairs/")
		if id != "" {
			action := strings.TrimPrefix(path, "/repairs/"+id)
			switch {
			case action == "/assign" && method == http.MethodPost:
				h.AssignRepair(w, r)
			case action == "/start" && method == http.MethodPost:
				h.StartRepair(w, r)
			case action == "/complete" && method == http.MethodPost:
				h.CompleteRepair(w, r)
			case action == "/confirm" && method == http.MethodPost:
				h.ConfirmRepair(w, r)
			case method == http.MethodGet:
				h.GetRepair(w, r)
			default:
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		} else {
			writeError(w, http.StatusBadRequest, "invalid repair id")
		}

	case path == "/bills" && method == http.MethodPost:
		h.CreateBill(w, r)
	case path == "/bills/generate" && method == http.MethodPost:
		h.GenerateMonthlyBills(w, r)
	case path == "/bills" && method == http.MethodGet:
		h.ListBills(w, r)
	case strings.HasPrefix(path, "/bills/"):
		id := getIDFromPath(path, "/bills/")
		if id != "" {
			action := strings.TrimPrefix(path, "/bills/"+id)
			switch {
			case action == "/pay" && method == http.MethodPost:
				h.PayBill(w, r)
			case action == "/penalty" && method == http.MethodGet:
				h.CalculatePenalty(w, r)
			case method == http.MethodGet:
				h.GetBill(w, r)
			default:
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		} else {
			writeError(w, http.StatusBadRequest, "invalid bill id")
		}

	case path == "/announcements" && method == http.MethodPost:
		h.CreateAnnouncement(w, r)
	case path == "/announcements" && method == http.MethodGet:
		h.ListAnnouncements(w, r)
	case strings.HasPrefix(path, "/announcements/"):
		switch method {
		case http.MethodPut:
			h.UpdateAnnouncement(w, r)
		case http.MethodDelete:
			h.DeleteAnnouncement(w, r)
		case http.MethodGet:
			h.GetAnnouncement(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}

	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}
