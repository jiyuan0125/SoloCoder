package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"qrcode-service/internal/qrcode"
	"qrcode-service/internal/storage"
	"qrcode-service/internal/types"
	"qrcode-service/internal/utils"
)

type Handler struct {
	db *storage.DB
}

func NewHandler(db *storage.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/qrcode", h.GenerateQRCode).Methods("POST")
	r.HandleFunc("/api/batch", h.BatchGenerate).Methods("POST")
	r.HandleFunc("/api/resources", h.CreateResource).Methods("POST")
	r.HandleFunc("/api/resources", h.ListResources).Methods("GET")
	r.HandleFunc("/api/resources/{id}", h.GetResource).Methods("GET")
	r.HandleFunc("/api/resources/{id}/summary", h.GetResourceSummary).Methods("GET")
	r.HandleFunc("/api/summaries", h.ListAllSummaries).Methods("GET")
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) GenerateQRCode(w http.ResponseWriter, r *http.Request) {
	req, err := parseQRRequest(r, false)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}
	
	data, err := qrcode.GenerateQR(req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	h.db.LogOperation(req.ResourceID, "single", strings.ToUpper(req.Format), 1)
	
	contentType := "image/png"
	ext := "png"
	switch strings.ToUpper(req.Format) {
	case "JPEG", "JPG":
		contentType = "image/jpeg"
		ext = "jpg"
	case "SVG":
		contentType = "image/svg+xml"
		ext = "svg"
	}
	
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "inline; filename=qrcode."+ext)
	w.Write(data)
}

func (h *Handler) BatchGenerate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse form")
		return
	}
	
	file, _, err := r.FormFile("csv")
	if err != nil {
		writeError(w, http.StatusBadRequest, "csv file is required")
		return
	}
	defer file.Close()
	
	csvData, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read csv file")
		return
	}
	
	lines, skipped, err := utils.ParseCSV(csvData)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse csv: "+err.Error())
		return
	}
	
	if len(lines) > 500 {
		writeError(w, http.StatusBadRequest, "batch limit exceeded: max 500 items")
		return
	}
	
	req, err := parseQRRequest(r, true)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	ext := "png"
	switch strings.ToUpper(req.Format) {
	case "JPEG", "JPG":
		ext = "jpg"
	case "SVG":
		ext = "svg"
	}
	
	zipFiles := make(map[string][]byte)
	results := make([]types.BatchResult, 0)
	successCount := 0
	
	for i, content := range lines {
		result := types.BatchResult{
			LineNumber: i + 1,
			Content:    content,
		}
		
		reqCopy := *req
		reqCopy.Content = content
		
		data, genErr := qrcode.GenerateQR(&reqCopy)
		if genErr != nil {
			result.Error = genErr.Error()
		} else {
			filename := utils.FormatFilename(i, content, ext)
			result.Filename = filename
			zipFiles[filename] = data
			successCount++
		}
		
		results = append(results, result)
	}
	
	zipData, err := utils.CreateZip(zipFiles)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create zip")
		return
	}
	
	h.db.LogOperation(req.ResourceID, "batch", strings.ToUpper(req.Format), len(lines))
	
	response := types.BatchResponse{
		ZipFile: zipData,
		Results: results,
		Total:   len(lines),
		Success: successCount,
		Skipped: skipped,
	}
	
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=qrcodes.zip")
	
	meta, _ := json.Marshal(map[string]interface{}{
		"total":   response.Total,
		"success": response.Success,
		"results": response.Results,
		"skipped": response.Skipped,
	})
	w.Header().Set("X-Batch-Meta", string(meta))
	
	w.Write(zipData)
}

func parseQRRequest(r *http.Request, isBatch bool) (*types.QRCodeRequest, error) {
	req := &types.QRCodeRequest{
		Content:      r.FormValue("content"),
		Format:       r.FormValue("format"),
		Foreground: r.FormValue("foreground"),
		Background: r.FormValue("background"),
		ErrorLevel: r.FormValue("error_level"),
		Title:       r.FormValue("title"),
	}
	
	if req.Format == "" {
		req.Format = "PNG"
	}
	
	if !utils.ValidateFormat(req.Format) {
		return nil, &validationError{Msg: "unsupported format"}
	}
	
	if req.Foreground == "" {
		req.Foreground = "#000000"
	}
	if req.Background == "" {
		req.Background = "#FFFFFF"
	}
	
	if _, err := utils.ParseColor(req.Foreground); err != nil {
		return nil, &validationError{Msg: "invalid foreground color"}
	}
	if _, err := utils.ParseColor(req.Background); err != nil {
		return nil, &validationError{Msg: "invalid background color"}
	}
	
	if sizeStr := r.FormValue("size"); sizeStr != "" {
		size, err := strconv.Atoi(sizeStr)
		if err != nil || size <= 0 {
			return nil, &validationError{Msg: "invalid size"}
		}
		req.Size = size
	} else {
		req.Size = 256
	}
	
	if borderStr := r.FormValue("border"); borderStr != "" {
		border, err := strconv.Atoi(borderStr)
		if err != nil {
			return nil, &validationError{Msg: "invalid border"}
		}
		req.Border = border
	}
	
	if resIDStr := r.FormValue("resource_id"); resIDStr != "" {
		resID, err := strconv.ParseInt(resIDStr, 10, 64)
		if err == nil {
			req.ResourceID = resID
		}
	}
	
	if !isBatch {
		if file, header, err := r.FormFile("logo"); err == nil {
			defer file.Close()
			if !utils.ValidateLogoFormat(header.Filename) {
				return nil, &validationError{Msg: "unsupported logo format"}
			}
			data, readErr := io.ReadAll(file)
			if readErr != nil {
				return nil, readErr
			}
			req.LogoFile = data
			req.LogoFilename = header.Filename
		}
	}
	
	return req, nil
}

type validationError struct {
	Msg string
}

func (e *validationError) Error() string {
	return e.Msg
}

func (h *Handler) CreateResource(w http.ResponseWriter, r *http.Request) {
	var res types.Resource
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	if res.Type == "" || res.Name == "" {
		writeError(w, http.StatusBadRequest, "type and name are required")
		return
	}
	
	id, err := h.db.CreateResource(&res)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	res.ID = id
	writeJSON(w, http.StatusCreated, res)
}

func (h *Handler) ListResources(w http.ResponseWriter, r *http.Request) {
	resources, err := h.db.ListResources()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resources)
}

func (h *Handler) GetResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid resource id")
		return
	}
	
	res, err := h.db.GetResource(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if res == nil {
		writeError(w, http.StatusNotFound, "resource not found")
		return
	}
	
	ops, _ := h.db.GetOperations(id, 100)
	
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"resource": res,
		"operations": ops,
	})
}

func (h *Handler) GetResourceSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid resource id")
		return
	}
	
	summary, err := h.db.GetResourceSummary(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) ListAllSummaries(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.db.ListAllSummaries()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summaries)
}
