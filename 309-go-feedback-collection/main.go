package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"feedback-collection/database"
	"feedback-collection/handler"
	"feedback-collection/models"
	"feedback-collection/service"
)

func main() {
	execDir, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable directory: %v", err)
	}
	execDir = filepath.Dir(execDir)

	dbPath := filepath.Join(execDir, "feedback.db")

	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/feedback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.CreateFeedbackHandler(w, r)
		} else {
			handler.WriteErrorResponse(w, http.StatusMethodNotAllowed, nil)
		}
	})

	mux.HandleFunc("/api/admin/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.LoginHandler(w, r)
		} else {
			handler.WriteErrorResponse(w, http.StatusMethodNotAllowed, nil)
		}
	})

	mux.HandleFunc("/api/admin/feedbacks", handler.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetFeedbackListHandler(w, r)
		} else {
			handler.WriteErrorResponse(w, http.StatusMethodNotAllowed, nil)
		}
	}))

	mux.HandleFunc("/api/admin/feedbacks/", handler.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/admin/feedbacks/")
		path = strings.TrimSuffix(path, "/")
		parts := strings.Split(path, "/")

		if len(parts) == 1 && parts[0] != "" {
			if r.Method == http.MethodGet {
				id, err := strconv.ParseInt(parts[0], 10, 64)
				if err != nil {
					handler.WriteErrorResponse(w, http.StatusBadRequest, service.ErrFeedbackNotFound)
					return
				}
				feedback, err := service.GetFeedbackByID(id)
				if err != nil {
					if err == service.ErrFeedbackNotFound {
						handler.WriteErrorResponse(w, http.StatusNotFound, err)
					} else {
						handler.WriteErrorResponse(w, http.StatusInternalServerError, err)
					}
					return
				}
				handler.WriteJSONResponse(w, http.StatusOK, "success", feedback)
				return
			}
		}

		if len(parts) == 2 {
			idStr := parts[0]
			action := parts[1]

			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				handler.WriteErrorResponse(w, http.StatusBadRequest, service.ErrFeedbackNotFound)
				return
			}

			if action == "status" && r.Method == http.MethodPut {
				var req models.UpdateFeedbackStatusRequest
				if err := handler.DecodeJSON(r, &req); err != nil {
					handler.WriteErrorResponse(w, http.StatusBadRequest, err)
					return
				}

				err = service.UpdateFeedbackStatus(id, &req)
				if err != nil {
					switch err {
					case service.ErrInvalidStatus, service.ErrProcessingNoteRequired:
						handler.WriteErrorResponse(w, http.StatusBadRequest, err)
					case service.ErrFeedbackNotFound:
						handler.WriteErrorResponse(w, http.StatusNotFound, err)
					default:
						handler.WriteErrorResponse(w, http.StatusInternalServerError, err)
					}
					return
				}

				handler.WriteJSONResponse(w, http.StatusOK, "status updated successfully", nil)
				return
			}

			if action == "note" && r.Method == http.MethodPost {
				var req models.AddInternalNoteRequest
				if err := handler.DecodeJSON(r, &req); err != nil {
					handler.WriteErrorResponse(w, http.StatusBadRequest, err)
					return
				}

				err = service.AddInternalNote(id, &req)
				if err != nil {
					switch err {
					case service.ErrNoteTooLong:
						handler.WriteErrorResponse(w, http.StatusBadRequest, err)
					case service.ErrFeedbackNotFound:
						handler.WriteErrorResponse(w, http.StatusNotFound, err)
					default:
						handler.WriteErrorResponse(w, http.StatusInternalServerError, err)
					}
					return
				}

				handler.WriteJSONResponse(w, http.StatusOK, "note added successfully", nil)
				return
			}
		}

		handler.WriteErrorResponse(w, http.StatusMethodNotAllowed, nil)
	}))

	mux.HandleFunc("/api/admin/statistics", handler.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetStatisticsHandler(w, r)
		} else {
			handler.WriteErrorResponse(w, http.StatusMethodNotAllowed, nil)
		}
	}))

	port := ":8080"
	log.Printf("Server starting on port %s...", port)
	log.Printf("Default admin: username=admin, password=admin123")
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
