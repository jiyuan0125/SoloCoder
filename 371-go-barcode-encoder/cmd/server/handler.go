package main

import (
	"encoding/json"
	"net/http"

	"barcode-encoder/pkg/barcode"
	"barcode-encoder/pkg/protocol"
)

type Handler struct {
	encoder barcode.Encoder
}

func NewHandler() *Handler {
	return &Handler{
		encoder: barcode.NewEncoder(),
	}
}

func (h *Handler) Encode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		response := protocol.NewErrorResponse("method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	var req protocol.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := protocol.NewErrorResponse("invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	result, err := h.encoder.Encode(req.Input)
	if err != nil {
		response := protocol.NewErrorResponse(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := protocol.NewSuccessResponse(
		result.Pattern,
		result.WidthSequence,
		result.TotalModules,
		result.Input,
	)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
