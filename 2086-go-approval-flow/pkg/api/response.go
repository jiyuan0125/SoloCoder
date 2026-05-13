package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"approval-flow/pkg/service"
	"approval-flow/pkg/store"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func WriteError(w http.ResponseWriter, err error) {
	var status int
	var message string

	switch {
	case errors.Is(err, store.ErrUserNotFound),
		errors.Is(err, store.ErrChainNotFound),
		errors.Is(err, store.ErrApplicationNotFound):
		status = http.StatusNotFound
		message = err.Error()
	case errors.Is(err, service.ErrChainEmpty),
		errors.Is(err, service.ErrInvalidCondition),
		errors.Is(err, service.ErrRejectReasonRequired):
		status = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, service.ErrSelfApproval),
		errors.Is(err, service.ErrNotApprover):
		status = http.StatusForbidden
		message = err.Error()
	default:
		status = http.StatusInternalServerError
		message = "Internal server error"
	}

	WriteJSON(w, status, &ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}
