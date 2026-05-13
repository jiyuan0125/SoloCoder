package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"multitenant/internal/customfield"
	"multitenant/internal/feature"
	"multitenant/internal/plan"
	"multitenant/internal/tenant"
	"multitenant/internal/usage"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	planSvc        *plan.Service
	tenantSvc      *tenant.Service
	featureSvc     *feature.Service
	customFieldSvc *customfield.Service
	usageSvc       *usage.Service
}

var validTenantID = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		planSvc:        plan.NewService(db),
		tenantSvc:      tenant.NewService(db),
		featureSvc:     feature.NewService(db),
		customFieldSvc: customfield.NewService(db),
		usageSvc:       usage.NewService(db),
	}
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

func validateTenantID(id string) bool {
	return validTenantID.MatchString(id)
}

func parseTenantIDFromPath(r *http.Request, prefix string) (string, bool) {
	path := r.URL.Path
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	remaining := path[len(prefix):]
	idx := strings.Index(remaining, "/")
	if idx == -1 {
		idx = len(remaining)
	}
	if idx == 0 {
		return "", false
	}
	return remaining[:idx], true
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/", h.Dispatch)
}

func (h *Handler) Dispatch(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api")
	if path == "" || path[0] != '/' {
		jsonError(w, http.StatusNotFound, "not found")
		return
	}
	path = path[1:]
	parts := strings.SplitN(path, "/", 4)

	if len(parts) == 0 {
		jsonError(w, http.StatusNotFound, "not found")
		return
	}

	switch parts[0] {
	case "plans":
		h.handlePlans(w, r, parts[1:])
	case "tenants":
		h.handleTenants(w, r, parts[1:])
	case "fields":
		h.handleFields(w, r, parts[1:])
	case "records":
		h.handleRecords(w, r, parts[1:])
	case "admin":
		h.handleAdmin(w, r, parts[1:])
	default:
		jsonError(w, http.StatusNotFound, "not found")
	}
}

func (h *Handler) handlePlans(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 0 || parts[0] == "" {
		h.PlansHandler(w, r)
		return
	}
	r.URL.Path = "/api/plans/" + parts[0]
	h.PlansDetailHandler(w, r)
}

func (h *Handler) handleTenants(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 0 || parts[0] == "" {
		h.TenantsHandler(w, r)
		return
	}
	if len(parts) == 1 {
		r.URL.Path = "/api/tenants/" + parts[0]
		h.TenantsDetailHandler(w, r)
		return
	}
	switch parts[1] {
	case "features":
		r.URL.Path = "/api/tenants/" + parts[0] + "/features"
		h.TenantFeaturesHandler(w, r)
	case "fields":
		r.URL.Path = "/api/tenants/" + parts[0] + "/fields"
		h.TenantFieldsHandler(w, r)
	case "usage":
		r.URL.Path = "/api/tenants/" + parts[0] + "/usage"
		h.TenantUsageHandler(w, r)
	case "notifications":
		r.URL.Path = "/api/tenants/" + parts[0] + "/notifications"
		h.TenantNotificationsHandler(w, r)
	case "change-plan":
		r.URL.Path = "/api/tenants/" + parts[0] + "/change-plan"
		h.TenantPlanChangeHandler(w, r)
	default:
		jsonError(w, http.StatusNotFound, "not found")
	}
}

func (h *Handler) handleFields(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 0 || parts[0] == "" {
		jsonError(w, http.StatusNotFound, "field id required")
		return
	}
	r.URL.Path = "/api/fields/" + parts[0]
	h.FieldDetailHandler(w, r)
}

func (h *Handler) handleRecords(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) < 2 || parts[0] == "" {
		jsonError(w, http.StatusNotFound, "record id required")
		return
	}
	r.URL.Path = "/api/records/" + parts[0] + "/" + parts[1]
	h.RecordValuesHandler(w, r)
}

func (h *Handler) handleAdmin(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) > 0 && parts[0] == "apply-pending-plans" {
		h.ApplyPendingPlansHandler(w, r)
		return
	}
	jsonError(w, http.StatusNotFound, "not found")
}

func (h *Handler) PlansHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		plans, err := h.planSvc.List()
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, plans)

	case http.MethodPost:
		var p plan.Plan
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		if p.ID == "" || p.Name == "" {
			jsonError(w, http.StatusBadRequest, "id and name required")
			return
		}
		if err := h.planSvc.Create(&p); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusCreated, p)

	default:
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) PlansDetailHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/plans/")
	if id == "" {
		jsonError(w, http.StatusBadRequest, "plan id required")
		return
	}
	switch r.Method {
	case http.MethodGet:
		p, err := h.planSvc.Get(id)
		if err == sql.ErrNoRows {
			jsonError(w, http.StatusNotFound, "plan not found")
			return
		}
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, p)

	case http.MethodPut:
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		p, err := h.planSvc.Update(id, updates)
		if err == sql.ErrNoRows {
			jsonError(w, http.StatusNotFound, "plan not found")
			return
		}
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, p)

	case http.MethodDelete:
		if err := h.planSvc.Delete(id); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})

	default:
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) TenantsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tenants, err := h.tenantSvc.List()
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, tenants)

	case http.MethodPost:
		var t tenant.Tenant
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		if t.ID == "" || t.Name == "" || t.PlanID == "" {
			jsonError(w, http.StatusBadRequest, "id, name, plan_id required")
			return
		}
		if !validateTenantID(t.ID) {
			jsonError(w, http.StatusBadRequest, "tenant id must be alphanumeric and underscore only")
			return
		}
		if _, err := h.planSvc.Get(t.PlanID); err == sql.ErrNoRows {
			jsonError(w, http.StatusNotFound, "plan not found")
			return
		}
		if err := h.tenantSvc.Create(&t); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusCreated, t)

	default:
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) TenantsDetailHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tenants/")
	parts := strings.SplitN(path, "/", 2)
	tenantID := parts[0]
	if tenantID == "" {
		jsonError(w, http.StatusBadRequest, "tenant id required")
		return
	}
	if !validateTenantID(tenantID) {
		jsonError(w, http.StatusBadRequest, "tenant id must be alphanumeric and underscore only")
		return
	}
	exists, err := h.tenantSvc.Exists(tenantID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		jsonError(w, http.StatusNotFound, "tenant not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		t, err := h.tenantSvc.Get(tenantID)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, t)

	case http.MethodPut:
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		t, err := h.tenantSvc.Update(tenantID, updates)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, t)

	case http.MethodDelete:
		if err := h.tenantSvc.Delete(tenantID); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})

	default:
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) TenantFeaturesHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tenants/")
	parts := strings.SplitN(path, "/", 3)
	tenantID := parts[0]
	if !validateTenantID(tenantID) {
		jsonError(w, http.StatusBadRequest, "tenant id must be alphanumeric and underscore only")
		return
	}
	exists, err := h.tenantSvc.Exists(tenantID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		jsonError(w, http.StatusNotFound, "tenant not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		features, err := h.featureSvc.List(tenantID)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, features)

	case http.MethodPost, http.MethodPut:
		var req struct {
			FeatureKey string `json:"feature_key"`
			Enabled    bool   `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.FeatureKey == "" {
			jsonError(w, http.StatusBadRequest, "feature_key required")
			return
		}
		if err := h.featureSvc.Set(tenantID, req.FeatureKey, req.Enabled); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"tenant_id":   tenantID,
			"feature_key": req.FeatureKey,
			"enabled":     req.Enabled,
		})

	default:
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) TenantFieldsHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tenants/")
	parts := strings.SplitN(path, "/", 3)
	tenantID := parts[0]
	if !validateTenantID(tenantID) {
		jsonError(w, http.StatusBadRequest, "tenant id must be alphanumeric and underscore only")
		return
	}
	exists, err := h.tenantSvc.Exists(tenantID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		jsonError(w, http.StatusNotFound, "tenant not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		includeDisabled := r.URL.Query().Get("include_disabled") == "true"
		fields, err := h.customFieldSvc.List(tenantID, includeDisabled)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, fields)

	case http.MethodPost:
		var f customfield.CustomField
		if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		f.TenantID = tenantID
		if f.FieldName == "" || f.FieldType == "" {
			jsonError(w, http.StatusBadRequest, "field_name and field_type required")
			return
		}
		err := h.customFieldSvc.Create(&f)
		if err == customfield.ErrInvalidFieldType {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err != nil && strings.Contains(err.Error(), "UNIQUE") {
			jsonError(w, http.StatusConflict, "field name already exists")
			return
		}
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusCreated, f)

	default:
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) FieldDetailHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/fields/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid field id")
		return
	}
	switch r.Method {
	case http.MethodGet:
		f, err := h.customFieldSvc.Get(id)
		if err == sql.ErrNoRows {
			jsonError(w, http.StatusNotFound, "field not found")
			return
		}
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, f)

	case http.MethodPut:
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		f, err := h.customFieldSvc.Update(id, updates)
		if err == sql.ErrNoRows {
			jsonError(w, http.StatusNotFound, "field not found")
			return
		}
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, f)

	case http.MethodDelete:
		err := h.customFieldSvc.Delete(id)
		if err == customfield.ErrFieldHasData {
			jsonError(w, http.StatusConflict, "field has data and cannot be deleted, disable it instead")
			return
		}
		if err == sql.ErrNoRows {
			jsonError(w, http.StatusNotFound, "field not found")
			return
		}
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})

	default:
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) RecordValuesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		recordID := strings.TrimPrefix(r.URL.Path, "/api/records/")
		idx := strings.Index(recordID, "/values")
		if idx != -1 {
			recordID = recordID[:idx]
		}
		if recordID == "" {
			jsonError(w, http.StatusBadRequest, "record id required")
			return
		}
		values, err := h.customFieldSvc.GetValues(recordID)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, values)
		return
	}

	if r.Method == http.MethodPost {
		recordID := strings.TrimPrefix(r.URL.Path, "/api/records/")
		idx := strings.Index(recordID, "/values")
		if idx != -1 {
			recordID = recordID[:idx]
		}
		if recordID == "" {
			jsonError(w, http.StatusBadRequest, "record id required")
			return
		}
		var req struct {
			FieldID int64  `json:"field_id"`
			Value   string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.FieldID == 0 {
			jsonError(w, http.StatusBadRequest, "field_id required")
			return
		}
		f, err := h.customFieldSvc.Get(req.FieldID)
		if err == sql.ErrNoRows {
			jsonError(w, http.StatusNotFound, "field not found")
			return
		}
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !f.Enabled {
			jsonError(w, http.StatusConflict, "field is disabled")
			return
		}
		if err := h.customFieldSvc.SetValue(req.FieldID, recordID, req.Value); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"record_id": recordID,
			"field_id":  req.FieldID,
			"value":     req.Value,
		})
		return
	}

	jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (h *Handler) TenantUsageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/tenants/")
	parts := strings.SplitN(path, "/", 3)
	tenantID := parts[0]
	if !validateTenantID(tenantID) {
		jsonError(w, http.StatusBadRequest, "tenant id must be alphanumeric and underscore only")
		return
	}
	exists, err := h.tenantSvc.Exists(tenantID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		jsonError(w, http.StatusNotFound, "tenant not found")
		return
	}

	t, err := h.tenantSvc.Get(tenantID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if r.Method == http.MethodGet {
		list, err := h.usageSvc.List(tenantID)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, list)
		return
	}

	var req struct {
		ResourceType string `json:"resource_type"`
		Amount       int    `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.ResourceType == "" {
		jsonError(w, http.StatusBadRequest, "resource_type required")
		return
	}
	if req.Amount <= 0 {
		jsonError(w, http.StatusBadRequest, "amount must be positive")
		return
	}

	p, err := h.planSvc.Get(t.PlanID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var maxLimit int
	switch req.ResourceType {
	case usage.ResourceUser:
		maxLimit = p.MaxUsers
	case usage.ResourceStorage:
		maxLimit = p.MaxStorageMB
	case usage.ResourceAPICall:
		maxLimit = p.MaxAPICalls
	default:
		jsonError(w, http.StatusBadRequest, "unknown resource type")
		return
	}

	newAmount, remaining, err := h.usageSvc.Add(tenantID, req.ResourceType, req.Amount, maxLimit, t.CycleStartAt, t.CycleEndAt)
	if err == usage.ErrQuotaExceeded {
		jsonResponse(w, http.StatusForbidden, map[string]interface{}{
			"error":     "quota exceeded",
			"used":      newAmount,
			"limit":     maxLimit,
			"remaining": remaining,
		})
		return
	}
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"resource_type": req.ResourceType,
		"used":          newAmount,
		"limit":         maxLimit,
		"remaining":     remaining,
	})
}

func (h *Handler) TenantNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/tenants/")
	parts := strings.SplitN(path, "/", 3)
	tenantID := parts[0]
	if !validateTenantID(tenantID) {
		jsonError(w, http.StatusBadRequest, "tenant id must be alphanumeric and underscore only")
		return
	}
	exists, err := h.tenantSvc.Exists(tenantID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		jsonError(w, http.StatusNotFound, "tenant not found")
		return
	}
	unreadOnly := r.URL.Query().Get("unread") == "true"
	list, err := h.usageSvc.ListNotifications(tenantID, unreadOnly)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, list)
}

func (h *Handler) ApplyPendingPlansHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	count, err := h.tenantSvc.ApplyPendingPlans()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"status":   "ok",
		"applied":  count,
		"message":  fmt.Sprintf("applied %d pending plan changes", count),
	})
}

func (h *Handler) TenantPlanChangeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/tenants/")
	parts := strings.SplitN(path, "/", 3)
	tenantID := parts[0]
	if !validateTenantID(tenantID) {
		jsonError(w, http.StatusBadRequest, "tenant id must be alphanumeric and underscore only")
		return
	}
	exists, err := h.tenantSvc.Exists(tenantID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		jsonError(w, http.StatusNotFound, "tenant not found")
		return
	}
	var req struct {
		PlanID string `json:"plan_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.PlanID == "" {
		jsonError(w, http.StatusBadRequest, "plan_id required")
		return
	}
	if _, err := h.planSvc.Get(req.PlanID); err == sql.ErrNoRows {
		jsonError(w, http.StatusNotFound, "plan not found")
		return
	}
	if err := h.tenantSvc.ChangePlan(tenantID, req.PlanID); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "pending",
		"message": "plan change scheduled for next billing cycle",
	})
}

func (h *Handler) WithTenantValidation(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			jsonError(w, http.StatusBadRequest, "X-Tenant-ID header required")
			return
		}
		if !validateTenantID(tenantID) {
			jsonError(w, http.StatusBadRequest, "tenant id must be alphanumeric and underscore only")
			return
		}
		exists, err := h.tenantSvc.Exists(tenantID)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !exists {
			jsonError(w, http.StatusNotFound, "tenant not found")
			return
		}
		next.ServeHTTP(w, r)
	}
}

func timePtr(t time.Time) *time.Time { return &t }
