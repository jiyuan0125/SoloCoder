package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"email-template/internal/email"
	"email-template/internal/template"
)

type Handler struct {
	tplSvc *template.Service
	mailSvc *email.Service
}

func NewHandler(tplSvc *template.Service, mailSvc *email.Service) *Handler {
	return &Handler{
		tplSvc:  tplSvc,
		mailSvc: mailSvc,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/templates", h.handleTemplates)
	mux.HandleFunc("/api/templates/", h.handleTemplateByName)
	mux.HandleFunc("/api/templates/validate", h.handleValidate)
	mux.HandleFunc("/api/send", h.handleSend)
	mux.HandleFunc("/api/send/batch", h.handleBatchSend)
	mux.HandleFunc("/api/status/", h.handleStatus)
	mux.HandleFunc("/api/batch/", h.handleBatchStatus)
	mux.HandleFunc("/api/images/upload", h.handleImageUpload)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) handleTemplates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		templates, err := h.tplSvc.List(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, templates)
	case http.MethodPost:
		var req template.CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		tpl, err := h.tplSvc.Create(r.Context(), &req)
		if err != nil {
			if errors.Is(err, template.ErrTemplateExists) {
				writeError(w, http.StatusConflict, "template name already exists")
				return
			}
			var htmlErr *template.HTMLSyntaxError
			if errors.As(err, &htmlErr) {
				writeJSON(w, http.StatusBadRequest, map[string]interface{}{
					"error":    "invalid HTML template",
					"position": htmlErr.Position,
					"details":  htmlErr.Err.Error(),
				})
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, tpl)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleTemplateByName(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/templates/")
	if name == "" {
		writeError(w, http.StatusBadRequest, "template name required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		tpl, err := h.tplSvc.Get(r.Context(), name)
		if err != nil {
			if errors.Is(err, template.ErrTemplateNotFound) {
				writeError(w, http.StatusNotFound, "template not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, tpl)
	case http.MethodPut:
		existing, err := h.tplSvc.Get(r.Context(), name)
		if err != nil {
			if errors.Is(err, template.ErrTemplateNotFound) {
				writeError(w, http.StatusNotFound, "template not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		var req template.UpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		req.ID = existing.ID

		tpl, err := h.tplSvc.Update(r.Context(), &req)
		if err != nil {
			var htmlErr *template.HTMLSyntaxError
			if errors.As(err, &htmlErr) {
				writeJSON(w, http.StatusBadRequest, map[string]interface{}{
					"error":    "invalid HTML template",
					"position": htmlErr.Position,
					"details":  htmlErr.Err.Error(),
				})
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, tpl)
	case http.MethodDelete:
		if err := h.tplSvc.Delete(r.Context(), name); err != nil {
			if errors.Is(err, template.ErrTemplateNotFound) {
				writeError(w, http.StatusNotFound, "template not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		HTML string `json:"html"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result := h.tplSvc.ValidateTemplate(r.Context(), req.HTML)
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		TemplateName string                 `json:"template_name"`
		Recipient    string                 `json:"recipient"`
		Variables    map[string]interface{} `json:"variables"`
		ImageURLs    []string               `json:"image_urls"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.TemplateName == "" || req.Recipient == "" {
		writeError(w, http.StatusBadRequest, "template_name and recipient required")
		return
	}

	var images []email.InlineImage
	for _, imgURL := range req.ImageURLs {
		img, err := h.mailSvc.UploadImageFromURL(r.Context(), imgURL)
		if err == nil {
			images = append(images, *img)
		}
	}

	resp, err := h.mailSvc.Send(r.Context(), &email.SendRequest{
		TemplateName: req.TemplateName,
		Recipient:    req.Recipient,
		Variables:    req.Variables,
		InlineImages: images,
	})
	if err != nil {
		if errors.Is(err, template.ErrTemplateNotFound) {
			writeError(w, http.StatusNotFound, "template not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, resp)
}

func (h *Handler) handleBatchSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	templateName := r.FormValue("template_name")
	if templateName == "" {
		writeError(w, http.StatusBadRequest, "template_name required")
		return
	}

	file, _, err := r.FormFile("csv_file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "csv_file required")
		return
	}
	defer file.Close()

	csvData, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read CSV file")
		return
	}

	variablesStr := r.FormValue("variables")
	var variables map[string]interface{}
	if variablesStr != "" {
		json.Unmarshal([]byte(variablesStr), &variables)
	}

	imageURLsStr := r.FormValue("image_urls")
	var imageURLs []string
	if imageURLsStr != "" {
		json.Unmarshal([]byte(imageURLsStr), &imageURLs)
	}

	var images []email.InlineImage
	for _, imgURL := range imageURLs {
		img, err := h.mailSvc.UploadImageFromURL(r.Context(), imgURL)
		if err == nil {
			images = append(images, *img)
		}
	}

	resp, err := h.mailSvc.BatchSend(r.Context(), &email.BatchSendRequest{
		TemplateName: templateName,
		CSVData:      csvData,
		Variables:    variables,
		InlineImages: images,
	})
	if err != nil {
		if strings.Contains(err.Error(), "exceeds") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, resp)
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/status/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "record id required")
		return
	}

	record, err := h.mailSvc.GetSendStatus(r.Context(), id)
	if err != nil {
		if err.Error() == "not found" {
			writeError(w, http.StatusNotFound, "record not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, record)
}

func (h *Handler) handleBatchStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	batchID := strings.TrimPrefix(r.URL.Path, "/api/batch/")
	if batchID == "" {
		writeError(w, http.StatusBadRequest, "batch id required")
		return
	}

	records, err := h.mailSvc.GetBatchStatus(r.Context(), batchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	total := len(records)
	sent := 0
	failed := 0
	pending := 0
	for _, r := range records {
		switch r.Status {
		case "sent":
			sent++
		case "failed":
			failed++
		default:
			pending++
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"batch_id": batchID,
		"total":    total,
		"sent":     sent,
		"failed":   failed,
		"pending":  pending,
		"records":  records,
	})
}

func (h *Handler) handleImageUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if _, err := url.Parse(req.URL); err != nil {
		writeError(w, http.StatusBadRequest, "invalid URL")
		return
	}

	img, err := h.mailSvc.UploadImageFromURL(r.Context(), req.URL)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to fetch image: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"cid":      img.CID,
		"filename": img.Filename,
		"size":     fmt.Sprintf("%d bytes", len(img.Data)),
	})
}
