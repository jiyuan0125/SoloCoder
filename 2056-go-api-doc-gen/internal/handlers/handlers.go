package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-api-doc-gen/internal/models"
	"go-api-doc-gen/internal/parser"
	"go-api-doc-gen/internal/renderer"
	"go-api-doc-gen/internal/storage"
)

type Handlers struct {
	store    *storage.Storage
	renderer *renderer.Renderer
}

func New(store *storage.Storage) *Handlers {
	return &Handlers{
		store:    store,
		renderer: renderer.New(),
	}
}

func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/versions", h.GetVersions)
	mux.HandleFunc("POST /api/versions", h.CreateVersion)
	mux.HandleFunc("GET /api/versions/{id}", h.GetVersionDetail)
	mux.HandleFunc("GET /api/versions/{id}/modules", h.GetModules)
	mux.HandleFunc("GET /api/versions/{id}/apis", h.GetAPIs)
	mux.HandleFunc("GET /api/versions/{id}/modules/{module}/apis", h.GetAPIsByModule)
	mux.HandleFunc("GET /api/versions/{id}/search", h.SearchAPIs)
	mux.HandleFunc("GET /api/versions/{id}/export/markdown", h.ExportMarkdown)
	mux.HandleFunc("GET /api/versions/{id}/export/openapi", h.ExportOpenAPI)
	mux.HandleFunc("GET /api/compare", h.CompareVersions)
	mux.HandleFunc("POST /api/test", h.TestAPI)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func (h *Handlers) GetVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := h.store.GetVersions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, versions)
}

type CreateVersionRequest struct {
	Name       string `json:"name"`
	Note       string `json:"note"`
	ModuleName string `json:"module_name"`
	SourceDir  string `json:"source_dir"`
}

func (h *Handlers) CreateVersion(w http.ResponseWriter, r *http.Request) {
	var req CreateVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.SourceDir == "" {
		writeError(w, http.StatusBadRequest, "source_dir 不能为空")
		return
	}

	if req.Name == "" {
		req.Name = fmt.Sprintf("v%s", time.Now().Format("20060102150405"))
	}

	moduleName := req.ModuleName
	if moduleName == "" {
		moduleName = "默认模块"
	}

	p := parser.New(moduleName)
	apis, err := p.Parse(req.SourceDir)
	if err != nil {
		if strings.Contains(err.Error(), "目录不存在") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	versionID, err := h.store.CreateVersion(req.Name, req.Note, req.SourceDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(apis) > 0 {
		if err := h.store.SaveAPIs(versionID, apis); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"version_id": versionID,
		"name":       req.Name,
		"api_count":  len(apis),
	})
}

func (h *Handlers) GetVersionDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的版本ID")
		return
	}

	version, err := h.store.GetVersion(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if version == nil {
		writeError(w, http.StatusNotFound, "版本不存在")
		return
	}

	apis, err := h.store.GetAPIs(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"version": version,
		"apis":    apis,
	})
}

func (h *Handlers) GetModules(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的版本ID")
		return
	}

	modules, err := h.store.GetModules(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if modules == nil {
		modules = []string{}
	}

	writeJSON(w, http.StatusOK, modules)
}

func (h *Handlers) GetAPIs(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的版本ID")
		return
	}

	apis, err := h.store.GetAPIs(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if apis == nil {
		apis = []models.API{}
	}

	writeJSON(w, http.StatusOK, apis)
}

func (h *Handlers) GetAPIsByModule(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的版本ID")
		return
	}

	module := r.PathValue("module")

	apis, err := h.store.GetAPIsByModule(id, module)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if apis == nil {
		apis = []models.API{}
	}

	writeJSON(w, http.StatusOK, apis)
}

func (h *Handlers) SearchAPIs(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的版本ID")
		return
	}

	keyword := r.URL.Query().Get("q")

	var apis []models.API
	if keyword == "" {
		apis, err = h.store.GetAPIs(id)
	} else {
		apis, err = h.store.SearchAPIs(id, keyword)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if apis == nil {
		apis = []models.API{}
	}

	writeJSON(w, http.StatusOK, apis)
}

func (h *Handlers) ExportMarkdown(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的版本ID")
		return
	}

	version, err := h.store.GetVersion(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if version == nil {
		writeError(w, http.StatusNotFound, "版本不存在")
		return
	}

	apis, err := h.store.GetAPIs(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	md := h.renderer.RenderMarkdown(apis, version.Name)

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"api-docs-%s.md\"", version.Name))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(md))
}

func (h *Handlers) ExportOpenAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的版本ID")
		return
	}

	version, err := h.store.GetVersion(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if version == nil {
		writeError(w, http.StatusNotFound, "版本不存在")
		return
	}

	apis, err := h.store.GetAPIs(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	openapi := h.renderer.RenderOpenAPI(apis, version.Name, "API Documentation")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"openapi-%s.json\"", version.Name))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(openapi))
}

func (h *Handlers) CompareVersions(w http.ResponseWriter, r *http.Request) {
	oldIDStr := r.URL.Query().Get("old")
	newIDStr := r.URL.Query().Get("new")

	if oldIDStr == "" || newIDStr == "" {
		writeError(w, http.StatusBadRequest, "需要提供 old 和 new 参数")
		return
	}

	oldID, err := strconv.ParseInt(oldIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 old 版本ID")
		return
	}

	newID, err := strconv.ParseInt(newIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 new 版本ID")
		return
	}

	oldAPIs, err := h.store.GetAPIs(oldID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	newAPIs, err := h.store.GetAPIs(newID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	diff := h.renderer.CompareVersions(oldAPIs, newAPIs)
	writeJSON(w, http.StatusOK, diff)
}

type TestAPIRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type TestAPIResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Duration   string            `json:"duration"`
}

func (h *Handlers) TestAPI(w http.ResponseWriter, r *http.Request) {
	var req TestAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "URL 不能为空")
		return
	}

	if req.Method == "" {
		req.Method = "GET"
	}

	start := time.Now()

	httpReq, err := http.NewRequest(req.Method, req.URL, bytes.NewBufferString(req.Body))
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("创建请求失败: %v", err))
		return
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("请求失败: %v", err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("读取响应失败: %v", err))
		return
	}

	duration := time.Since(start)

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	writeJSON(w, http.StatusOK, TestAPIResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(body),
		Duration:   duration.String(),
	})
}
