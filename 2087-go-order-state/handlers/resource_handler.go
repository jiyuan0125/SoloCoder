package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"order-state-machine/models"
	"strconv"
	"strings"
	"time"
)

type ResourceHandler struct {
	db *sql.DB
}

type CreateResourceRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func NewResourceHandler(db *sql.DB) *ResourceHandler {
	return &ResourceHandler{db: db}
}

func (h *ResourceHandler) HandleResources(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listResources(w, r)
	case http.MethodPost:
		h.createResource(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ResourceHandler) listResources(w http.ResponseWriter, r *http.Request) {
	resourceType := r.URL.Query().Get("type")

	var rows *sql.Rows
	var err error

	if resourceType != "" {
		rows, err = h.db.Query(`
			SELECT id, name, type, created_at
			FROM resources
			WHERE type = ?
			ORDER BY id DESC
		`, resourceType)
	} else {
		rows, err = h.db.Query(`
			SELECT id, name, type, created_at
			FROM resources
			ORDER BY id DESC
		`)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	resources := []*models.Resource{}
	for rows.Next() {
		resource := &models.Resource{}
		err := rows.Scan(&resource.ID, &resource.Name, &resource.Type, &resource.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resources = append(resources, resource)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}

func (h *ResourceHandler) createResource(w http.ResponseWriter, r *http.Request) {
	var req CreateResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "资源名称不能为空", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		http.Error(w, "资源类型不能为空", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(`
		INSERT INTO resources (name, type)
		VALUES (?, ?)
	`, req.Name, req.Type)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()

	resource := &models.Resource{
		ID:        int(id),
		Name:      req.Name,
		Type:      req.Type,
		CreatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resource)
}

func (h *ResourceHandler) HandleResourceByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/resources/")
	idStr := strings.TrimSuffix(path, "/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的资源ID", http.StatusBadRequest)
		return
	}

	resource, err := h.getResourceByID(id)
	if err == sql.ErrNoRows {
		http.Error(w, "资源不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resource)
}

func (h *ResourceHandler) getResourceByID(id int) (*models.Resource, error) {
	resource := &models.Resource{}
	err := h.db.QueryRow(`
		SELECT id, name, type, created_at
		FROM resources WHERE id = ?
	`, id).Scan(&resource.ID, &resource.Name, &resource.Type, &resource.CreatedAt)

	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (h *ResourceHandler) HandleResourceSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/resources/summary/")
	idStr := strings.TrimSuffix(path, "/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的资源ID", http.StatusBadRequest)
		return
	}

	resource, err := h.getResourceByID(id)
	if err == sql.ErrNoRows {
		http.Error(w, "资源不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var orderCount int
	var totalAmount float64
	err = h.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(amount), 0)
		FROM orders WHERE resource_id = ?
	`, id).Scan(&orderCount, &totalAmount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var orderOpCount int
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM order_history WHERE resource_id = ?
	`, id).Scan(&orderOpCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var asOpCount int
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM aftersales_history WHERE resource_id = ?
	`, id).Scan(&asOpCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	summary := &models.ResourceSummary{
		ResourceID:     resource.ID,
		ResourceName:   resource.Name,
		ResourceType:   resource.Type,
		OrderCount:     orderCount,
		TotalAmount:    totalAmount,
		OperationCount: orderOpCount + asOpCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
