package main

import (
	"announcement-board/common"
	"net/http"
	"strings"
)

func setupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/departments", handleListDepartments)
	mux.HandleFunc("/api/user/info", handleGetUserInfo)

	mux.HandleFunc("/api/announcements", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleCreateAnnouncement(w, r)
		} else if r.Method == http.MethodGet {
			userID := getUserID(r)
			user, ok := globalStore.GetUser(userID)
			if ok && user.Role == common.RoleAdmin {
				handleListAdminAnnouncements(w, r)
			} else {
				handleListEmployeeAnnouncements(w, r)
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/announcements/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) <= len("/api/announcements/") {
			http.NotFound(w, r)
			return
		}

		remaining := path[len("/api/announcements/"):]

		if strings.HasSuffix(remaining, "/toggle-pin") {
			if r.Method == http.MethodPost {
				handleTogglePin(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if r.Method == http.MethodGet {
			handleGetAnnouncement(w, r)
		} else if r.Method == http.MethodPut {
			handleUpdateAnnouncement(w, r)
		} else if r.Method == http.MethodDelete {
			handleDeleteDraft(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/admin/announcements", handleListAdminAnnouncements)

	mux.HandleFunc("/api/approvals/submit", handleSubmitApproval)
	mux.HandleFunc("/api/approvals/approve", handleApproveAnnouncement)
	mux.HandleFunc("/api/approvals/reject", handleRejectAnnouncement)
	mux.HandleFunc("/api/approvals/pending", handleListPendingApprovals)

	return mux
}
