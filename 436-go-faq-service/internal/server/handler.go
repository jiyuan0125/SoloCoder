package server

import (
	"encoding/json"
	"go-faq-service/pkg/protocol"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func jsonResponse(w http.ResponseWriter, status int, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := protocol.Response{
		Success: success,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) CreateFAQ(w http.ResponseWriter, r *http.Request) {
	var req protocol.CreateFAQRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	faq, err := h.service.CreateFAQ(&req)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	jsonResponse(w, http.StatusCreated, true, "FAQ created successfully", faq)
}

func (h *Handler) GetFAQ(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "zh"
	}

	faq, err := h.service.GetFAQ(id, protocol.Language(lang))
	if err != nil {
		jsonResponse(w, http.StatusNotFound, false, err.Error(), nil)
		return
	}

	jsonResponse(w, http.StatusOK, true, "", faq)
}

func (h *Handler) UpdateFAQ(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req protocol.UpdateFAQRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	faq, err := h.service.UpdateFAQ(id, &req)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	jsonResponse(w, http.StatusOK, true, "FAQ updated successfully", faq)
}

func (h *Handler) DeleteFAQ(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteFAQ(id); err != nil {
		jsonResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	jsonResponse(w, http.StatusOK, true, "FAQ deleted successfully", nil)
}

func (h *Handler) SetFAQEnabled(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req protocol.EnableFAQRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	faq, err := h.service.SetFAQEnabled(id, req.IsEnabled)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	action := "enabled"
	if !req.IsEnabled {
		action = "disabled"
	}
	jsonResponse(w, http.StatusOK, true, "FAQ "+action+" successfully", faq)
}

func (h *Handler) SetFAQPinned(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req protocol.PinFAQRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	faq, err := h.service.SetFAQPinned(id, req.IsPinned)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	action := "pinned"
	if !req.IsPinned {
		action = "unpinned"
	}
	jsonResponse(w, http.StatusOK, true, "FAQ "+action+" successfully", faq)
}

func (h *Handler) SearchFAQ(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "zh"
	}

	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page <= 0 {
		page = 1
	}

	pageSizeStr := r.URL.Query().Get("page_size")
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize <= 0 {
		pageSize = 10
	}

	req := protocol.SearchFAQRequest{
		Keyword:  keyword,
		Language: protocol.Language(lang),
		Page:     page,
		PageSize: pageSize,
	}

	result, err := h.service.SearchFAQ(&req)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	jsonResponse(w, http.StatusOK, true, "", result)
}

func (h *Handler) BatchImportFAQ(w http.ResponseWriter, r *http.Request) {
	var req protocol.BatchImportFAQRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	result, err := h.service.BatchImportFAQ(&req)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	jsonResponse(w, http.StatusOK, true, "batch import completed", result)
}

func (h *Handler) RecordClick(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req protocol.RecordClickRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	if err := h.service.RecordClick(id, &req); err != nil {
		jsonResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	jsonResponse(w, http.StatusOK, true, "click recorded", nil)
}

func (h *Handler) GetFAQHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	result, err := h.service.GetFAQHistory(id)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	jsonResponse(w, http.StatusOK, true, "", result)
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req protocol.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	category, err := h.service.CreateCategory(&req)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	jsonResponse(w, http.StatusCreated, true, "category created successfully", category)
}

func (h *Handler) GetCategoryTree(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetCategoryTree()
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	jsonResponse(w, http.StatusOK, true, "", result)
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req protocol.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	category, err := h.service.UpdateCategory(id, &req)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	jsonResponse(w, http.StatusOK, true, "category updated successfully", category)
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteCategory(id); err != nil {
		jsonResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	jsonResponse(w, http.StatusOK, true, "category deleted successfully", nil)
}

func (h *Handler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStatistics()
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	jsonResponse(w, http.StatusOK, true, "", stats)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, true, "service is running", map[string]string{"status": "healthy"})
}

func (h *Handler) ListAllFAQs(w http.ResponseWriter, r *http.Request) {
	allFAQs := h.service.storage.ListAllFAQs()
	faqs := make([]protocol.FAQ, len(allFAQs))
	for i, f := range allFAQs {
		faqs[i] = *f
	}
	jsonResponse(w, http.StatusOK, true, "", faqs)
}

func (h *Handler) ListAllCategories(w http.ResponseWriter, r *http.Request) {
	allCats := h.service.storage.ListAllCategories()
	cats := make([]protocol.Category, len(allCats))
	for i, c := range allCats {
		cats[i] = *c
	}
	jsonResponse(w, http.StatusOK, true, "", cats)
}

func (h *Handler) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
