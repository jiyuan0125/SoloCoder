package main

import (
	"dns-cache-service/common"
	"dns-cache-service/dnscache"
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	service *dnscache.Service
}

func NewHandler(service *dnscache.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) parseRecordTypes(typesStr string) []common.RecordType {
	if typesStr == "" {
		return nil
	}

	parts := strings.Split(typesStr, ",")
	types := make([]common.RecordType, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(strings.ToUpper(part))
		switch part {
		case "A":
			types = append(types, common.TypeA)
		case "AAAA":
			types = append(types, common.TypeAAAA)
		case "CNAME":
			types = append(types, common.TypeCNAME)
		case "MX":
			types = append(types, common.TypeMX)
		case "TXT":
			types = append(types, common.TypeTXT)
		}
	}

	return types
}

func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	domain := r.URL.Query().Get("domain")
	if domain == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain parameter is required"})
		return
	}

	typesStr := r.URL.Query().Get("types")
	types := h.parseRecordTypes(typesStr)

	response := h.service.Query(domain, types)
	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req common.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Domain == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain is required"})
		return
	}

	response := h.service.Refresh(req.Domain, req.Types)
	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	stats := h.service.GetStats()
	h.writeJSON(w, http.StatusOK, stats)
}

func (h *Handler) Preload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req common.PreloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if len(req.Domains) == 0 {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domains list is required"})
		return
	}

	response := h.service.PreloadDomains(req.Domains, []common.RecordType{common.TypeA, common.TypeAAAA})
	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) Entries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	entries := h.service.GetAllEntries()
	h.writeJSON(w, http.StatusOK, entries)
}
