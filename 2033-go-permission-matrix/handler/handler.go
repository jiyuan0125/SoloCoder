package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"permission-matrix/service"
	"strconv"
	"strings"
)

type Handler struct {
	svc *service.PermissionService
}

func NewHandler(svc *service.PermissionService) *Handler {
	return &Handler{svc: svc}
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{Success: true, Data: data})
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{Success: false, Error: err.Error()})
}

func parseID(r *http.Request, key string) (int64, error) {
	parts := strings.Split(r.URL.Path, "/")
	for i, part := range parts {
		if part == key && i+1 < len(parts) {
			return strconv.ParseInt(parts[i+1], 10, 64)
		}
	}
	return 0, errors.New("id not found")
}

func getUserID(r *http.Request) string {
	parts := strings.Split(r.URL.Path, "/")
	for i, part := range parts {
		if part == "users" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func getLastPathSegment(r *http.Request) string {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

type createRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req createRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, errors.New("角色名称不能为空"))
		return
	}

	role, err := h.svc.CreateRole(req.Name, req.Description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, role)
}

func (h *Handler) GetRole(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "roles")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	role, err := h.svc.GetRole(id)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, role)
}

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.svc.ListRoles()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, roles)
}

func (h *Handler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "roles")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.svc.DeleteRole(id); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, nil)
}

type createPermissionRequest struct {
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

func (h *Handler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var req createPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Resource == "" || req.Action == "" {
		writeError(w, http.StatusBadRequest, errors.New("资源和操作不能为空"))
		return
	}

	perm, err := h.svc.CreatePermission(req.Resource, req.Action, req.Description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, perm)
}

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	perms, err := h.svc.ListPermissions(resource)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, perms)
}

type assignPermissionRequest struct {
	PermissionID int64 `json:"permission_id"`
	IsAllowed    bool  `json:"is_allowed"`
}

func (h *Handler) AssignPermissionToRole(w http.ResponseWriter, r *http.Request) {
	roleID, err := parseID(r, "roles")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req assignPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.svc.AssignPermissionToRole(roleID, req.PermissionID, req.IsAllowed); err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, nil)
}

func (h *Handler) RemovePermissionFromRole(w http.ResponseWriter, r *http.Request) {
	roleID, err := parseID(r, "roles")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	permID, err := parseID(r, "permissions")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.svc.RemovePermissionFromRole(roleID, permID); err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, nil)
}

func (h *Handler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	roleID, err := parseID(r, "roles")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	perms, err := h.svc.GetRolePermissions(roleID)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, perms)
}

type addInheritanceRequest struct {
	InheritsFrom int64 `json:"inherits_from"`
}

func (h *Handler) AddRoleInheritance(w http.ResponseWriter, r *http.Request) {
	roleID, err := parseID(r, "roles")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req addInheritanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.svc.AddRoleInheritance(roleID, req.InheritsFrom); err != nil {
		if errors.Is(err, service.ErrCircularInheritance) {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if errors.Is(err, service.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, nil)
}

func (h *Handler) RemoveRoleInheritance(w http.ResponseWriter, r *http.Request) {
	roleID, err := parseID(r, "roles")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	parentID, err := parseID(r, "parents")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.svc.RemoveRoleInheritance(roleID, parentID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, nil)
}

func (h *Handler) GetRoleParents(w http.ResponseWriter, r *http.Request) {
	roleID, err := parseID(r, "roles")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	parents, err := h.svc.GetRoleParents(roleID)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, parents)
}

type assignRoleRequest struct {
	RoleID int64 `json:"role_id"`
}

func (h *Handler) AssignRoleToUser(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	if userID == "" {
		writeError(w, http.StatusBadRequest, errors.New("用户ID不能为空"))
		return
	}

	var req assignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.svc.AssignRoleToUser(userID, req.RoleID); err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, nil)
}

func (h *Handler) RemoveRoleFromUser(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	if userID == "" {
		writeError(w, http.StatusBadRequest, errors.New("用户ID不能为空"))
		return
	}

	roleID, err := parseID(r, "roles")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.svc.RemoveRoleFromUser(userID, roleID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, nil)
}

func (h *Handler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	if userID == "" {
		writeError(w, http.StatusBadRequest, errors.New("用户ID不能为空"))
		return
	}

	roles, err := h.svc.GetUserRoles(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, roles)
}

func (h *Handler) GetUserEffectivePermissions(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	if userID == "" {
		writeError(w, http.StatusBadRequest, errors.New("用户ID不能为空"))
		return
	}

	perms, err := h.svc.GetEffectivePermissions(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, perms)
}

func (h *Handler) CheckUserPermission(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	action := r.URL.Query().Get("action")
	if resource == "" || action == "" {
		writeError(w, http.StatusBadRequest, errors.New("resource 和 action 参数不能为空"))
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	var userID string
	for i, part := range parts {
		if part == "users" && i+1 < len(parts) {
			userID = parts[i+1]
			break
		}
	}
	if userID == "" {
		writeError(w, http.StatusBadRequest, errors.New("用户ID不能为空"))
		return
	}

	result, err := h.svc.CheckUserPermission(userID, resource, action)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}
