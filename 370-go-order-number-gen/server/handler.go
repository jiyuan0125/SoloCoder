package main

import (
	"encoding/json"
	"net/http"

	"ordergen/ordergen"
	"ordergen/protocol"
)

type Handler struct {
	generator *ordergen.OrderGenerator
}

func NewHandler(generator *ordergen.OrderGenerator) *Handler {
	return &Handler{generator: generator}
}

func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body")
		return
	}

	var orderNo string
	var err error

	if req.Prefix != "" {
		orderNo, err = h.generator.GenerateWithPrefix(req.Prefix)
	} else {
		orderNo, err = h.generator.Generate()
	}

	if err != nil {
		h.sendError(w, err.Error())
		return
	}

	resp := protocol.GenerateResponse{
		Success: true,
		OrderNo: orderNo,
	}

	h.sendJSON(w, resp)
}

func (h *Handler) Parse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body")
		return
	}

	if req.OrderNo == "" {
		h.sendError(w, "Order number is required")
		return
	}

	info, err := ordergen.ParseOrderNo(req.OrderNo)
	if err != nil {
		h.sendError(w, err.Error())
		return
	}

	resp := protocol.ParseResponse{
		Success:   true,
		Prefix:    info.Prefix,
		Timestamp: info.Timestamp.Format(ordergen.TimeFormat),
		SerialNum: info.SerialNum,
		RandomNum: info.RandomNum,
	}

	h.sendJSON(w, resp)
}

func (h *Handler) Prefix(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getPrefix(w, r)
	case http.MethodPut:
		h.setPrefix(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) getPrefix(w http.ResponseWriter, r *http.Request) {
	prefix := h.generator.GetPrefix()
	resp := protocol.GetPrefixResponse{
		Success: true,
		Prefix:  prefix,
	}
	h.sendJSON(w, resp)
}

func (h *Handler) setPrefix(w http.ResponseWriter, r *http.Request) {
	var req protocol.SetPrefixRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body")
		return
	}

	h.generator.SetPrefix(req.Prefix)

	resp := protocol.SetPrefixResponse{
		Success: true,
	}
	h.sendJSON(w, resp)
}

func (h *Handler) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) sendError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
