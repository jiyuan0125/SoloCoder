package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"feedback-collection/models"
	"feedback-collection/service"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	resp, err := service.Login(&req)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			writeErrorResponse(w, http.StatusUnauthorized, err)
		} else {
			writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSONResponse(w, http.StatusOK, "login successful", resp)
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeErrorResponse(w, http.StatusUnauthorized, service.ErrTokenInvalid)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeErrorResponse(w, http.StatusUnauthorized, service.ErrTokenInvalid)
			return
		}

		token := parts[1]
		valid, err := service.ValidateToken(token)
		if err != nil || !valid {
			writeErrorResponse(w, http.StatusUnauthorized, service.ErrTokenInvalid)
			return
		}

		next.ServeHTTP(w, r)
	}
}
