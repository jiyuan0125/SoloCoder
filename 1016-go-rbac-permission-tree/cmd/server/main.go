package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/example/rbac-permission-tree/internal/api"
	"github.com/example/rbac-permission-tree/internal/rbac"
)

type Server struct {
	store *rbac.Store
	mux   *http.ServeMux
}

func NewServer() *Server {
	s := &Server{
		store: rbac.NewStore(),
		mux:   http.NewServeMux(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/roles", s.handleRoles)
	s.mux.HandleFunc("/roles/", s.handleRoleByID)
	s.mux.HandleFunc("/check", s.handleCheck)
	s.mux.HandleFunc("/users/", s.handleUserByID)
}

func (s *Server) handleRoles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listRoles(w, r)
	case http.MethodPost:
		s.createRole(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listRoles(w http.ResponseWriter, r *http.Request) {
	roles := s.store.ListRoles()
	apiRoles := make([]api.Role, 0, len(roles))
	for _, role := range roles {
		apiRoles = append(apiRoles, api.Role{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		})
	}
	writeJSON(w, http.StatusOK, api.ListRolesResponse{Roles: apiRoles})
}

func (s *Server) createRole(w http.ResponseWriter, r *http.Request) {
	var req api.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	role, err := s.store.CreateRole(req.ID, req.Name, req.Description)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleAlreadyExists) {
			writeError(w, http.StatusConflict, err)
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusCreated, api.Role{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	})
}

func (s *Server) handleRoleByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/roles/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "role id required", http.StatusBadRequest)
		return
	}

	roleID := parts[0]

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			s.getRole(w, r, roleID)
		case http.MethodPut:
			s.updateRole(w, r, roleID)
		case http.MethodDelete:
			s.deleteRole(w, r, roleID)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	subPath := parts[1]
	switch {
	case subPath == "permissions":
		s.handleRolePermissions(w, r, roleID)
	case subPath == "parents":
		s.handleRoleParents(w, r, roleID)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (s *Server) getRole(w http.ResponseWriter, r *http.Request, roleID string) {
	role, err := s.store.GetRole(roleID)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	writeJSON(w, http.StatusOK, api.Role{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	})
}

func (s *Server) updateRole(w http.ResponseWriter, r *http.Request, roleID string) {
	var req api.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	role, err := s.store.UpdateRole(roleID, req.Name, req.Description)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	writeJSON(w, http.StatusOK, api.Role{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	})
}

func (s *Server) deleteRole(w http.ResponseWriter, r *http.Request, roleID string) {
	err := s.store.DeleteRole(roleID)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRolePermissions(w http.ResponseWriter, r *http.Request, roleID string) {
	switch r.Method {
	case http.MethodGet:
		s.getRolePermissions(w, r, roleID)
	case http.MethodPost:
		s.addPermissionToRole(w, r, roleID)
	case http.MethodDelete:
		s.removePermissionFromRole(w, r, roleID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getRolePermissions(w http.ResponseWriter, r *http.Request, roleID string) {
	perms, err := s.store.GetRolePermissions(roleID)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	apiPerms := make([]api.Permission, 0, len(perms))
	for _, perm := range perms {
		apiPerms = append(apiPerms, api.Permission{
			Resource: perm.Resource,
			Action:   perm.Action,
			Scope:    api.Scope(perm.Scope),
		})
	}
	writeJSON(w, http.StatusOK, api.GetRolePermissionsResponse{Permissions: apiPerms})
}

func (s *Server) addPermissionToRole(w http.ResponseWriter, r *http.Request, roleID string) {
	var req api.AddPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	for _, perm := range req.Permissions {
		err := s.store.AddPermissionToRole(roleID, rbac.Permission{
			Resource: perm.Resource,
			Action:   perm.Action,
			Scope:    rbac.Scope(perm.Scope),
		})
		if err != nil {
			if errors.Is(err, rbac.ErrRoleNotFound) {
				writeError(w, http.StatusNotFound, err)
			} else {
				writeError(w, http.StatusInternalServerError, err)
			}
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) removePermissionFromRole(w http.ResponseWriter, r *http.Request, roleID string) {
	var req api.RemovePermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	err := s.store.RemovePermissionFromRole(roleID, req.Resource, req.Action)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
		} else if errors.Is(err, rbac.ErrPermissionNotFound) {
			writeError(w, http.StatusNotFound, err)
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRoleParents(w http.ResponseWriter, r *http.Request, roleID string) {
	switch r.Method {
	case http.MethodGet:
		s.getRoleParents(w, r, roleID)
	case http.MethodPost:
		s.addParent(w, r, roleID)
	case http.MethodDelete:
		s.removeParent(w, r, roleID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getRoleParents(w http.ResponseWriter, r *http.Request, roleID string) {
	parents, err := s.store.GetRoleParents(roleID)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	writeJSON(w, http.StatusOK, api.GetRoleParentsResponse{Parents: parents})
}

func (s *Server) addParent(w http.ResponseWriter, r *http.Request, roleID string) {
	var req api.AddParentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	err := s.store.AddParent(roleID, req.ParentID)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
		} else if errors.Is(err, rbac.ErrInheritanceCycle) {
			writeError(w, http.StatusConflict, err)
		} else if errors.Is(err, rbac.ErrInheritanceConflict) {
			writeError(w, http.StatusConflict, err)
		} else if errors.Is(err, rbac.ErrSelfInheritance) {
			writeError(w, http.StatusBadRequest, err)
		} else if errors.Is(err, rbac.ErrInheritanceExists) {
			writeError(w, http.StatusConflict, err)
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) removeParent(w http.ResponseWriter, r *http.Request, roleID string) {
	var req api.RemoveParentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	err := s.store.RemoveParent(roleID, req.ParentID)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUserByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "user id required", http.StatusBadRequest)
		return
	}

	userID := parts[0]

	if len(parts) == 1 || parts[1] != "roles" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getUserRoles(w, r, userID)
	case http.MethodPost:
		s.addRoleToUser(w, r, userID)
	case http.MethodDelete:
		s.removeRoleFromUser(w, r, userID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getUserRoles(w http.ResponseWriter, r *http.Request, userID string) {
	roles := s.store.GetUserRoles(userID)
	writeJSON(w, http.StatusOK, api.GetUserRolesResponse{Roles: roles})
}

func (s *Server) addRoleToUser(w http.ResponseWriter, r *http.Request, userID string) {
	var req api.AddRoleToUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	err := s.store.AddRoleToUser(userID, req.RoleID)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
		} else if errors.Is(err, rbac.ErrMutexRoleConflict) {
			writeError(w, http.StatusConflict, err)
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) removeRoleFromUser(w http.ResponseWriter, r *http.Request, userID string) {
	var req api.RemoveRoleFromUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	err := s.store.RemoveRoleFromUser(userID, req.RoleID)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			writeError(w, http.StatusNotFound, err)
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.CheckPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result := s.store.CheckPermission(req.UserID, req.Resource, req.Action)
	writeJSON(w, http.StatusOK, api.CheckPermissionResponse{
		Allowed: result.Allowed,
		Scope:   api.Scope(result.Scope),
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, api.ErrorResponse{Error: err.Error()})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func main() {
	server := NewServer()
	port := ":8080"
	fmt.Printf("RBAC server running on %s\n", port)
	log.Fatal(http.ListenAndServe(port, server))
}
