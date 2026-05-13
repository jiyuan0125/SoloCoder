package router

import (
	"net/http"
	"permission-matrix/handler"
	"strings"
)

type routeInfo struct {
	userID  string
	roleID  string
	permID  string
	action  string
}

func parseUserRoute(path string) routeInfo {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	info := routeInfo{}

	if len(parts) >= 2 && parts[0] == "users" {
		if len(parts) >= 2 {
			info.userID = parts[1]
		}
		if len(parts) >= 3 {
			info.action = parts[2]
		}
		if len(parts) >= 4 {
			if parts[2] == "roles" {
				info.roleID = parts[3]
			}
		}
	}

	return info
}

func NewRouter(h *handler.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/roles", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListRoles(w, r)
		case http.MethodPost:
			h.CreateRole(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/roles/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		parts := strings.Split(path, "/")

		if len(parts) == 3 {
			switch r.Method {
			case http.MethodGet:
				h.GetRole(w, r)
			case http.MethodDelete:
				h.DeleteRole(w, r)
			default:
				http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			}
			return
		}

		if len(parts) >= 4 {
			switch parts[3] {
			case "permissions":
				handleRolePermissions(w, r, h, parts)
			case "parents":
				handleRoleParents(w, r, h, parts)
			default:
				http.Error(w, "路径不存在", http.StatusNotFound)
			}
			return
		}

		http.Error(w, "路径不存在", http.StatusNotFound)
	})

	mux.HandleFunc("/permissions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListPermissions(w, r)
		case http.MethodPost:
			h.CreatePermission(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		info := parseUserRoute(r.URL.Path)

		if info.userID == "" {
			http.Error(w, "路径不存在", http.StatusNotFound)
			return
		}

		switch info.action {
		case "":
			if r.Method == http.MethodGet {
				h.GetUserRoles(w, r)
				return
			}
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)

		case "roles":
			handleUserRoles(w, r, h, info)

		case "permissions":
			if r.Method == http.MethodGet {
				h.GetUserEffectivePermissions(w, r)
				return
			}
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)

		case "check":
			if r.Method == http.MethodGet {
				h.CheckUserPermission(w, r)
				return
			}
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)

		default:
			http.Error(w, "路径不存在", http.StatusNotFound)
		}
	})

	return mux
}

func handleRolePermissions(w http.ResponseWriter, r *http.Request, h *handler.Handler, parts []string) {
	switch r.Method {
	case http.MethodGet:
		h.GetRolePermissions(w, r)
	case http.MethodPost:
		h.AssignPermissionToRole(w, r)
	case http.MethodDelete:
		if len(parts) >= 5 {
			h.RemovePermissionFromRole(w, r)
			return
		}
		http.Error(w, "路径不存在", http.StatusNotFound)
	default:
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
	}
}

func handleRoleParents(w http.ResponseWriter, r *http.Request, h *handler.Handler, parts []string) {
	switch r.Method {
	case http.MethodGet:
		h.GetRoleParents(w, r)
	case http.MethodPost:
		h.AddRoleInheritance(w, r)
	case http.MethodDelete:
		if len(parts) >= 5 {
			h.RemoveRoleInheritance(w, r)
			http.Error(w, "路径不存在", http.StatusNotFound)
		}
		http.Error(w, "路径不存在", http.StatusNotFound)
	default:
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
	}
}

func handleUserRoles(w http.ResponseWriter, r *http.Request, h *handler.Handler, info routeInfo) {
	switch r.Method {
	case http.MethodGet:
		h.GetUserRoles(w, r)
	case http.MethodPost:
		h.AssignRoleToUser(w, r)
	case http.MethodDelete:
		if info.roleID != "" {
			h.RemoveRoleFromUser(w, r)
			return
		}
		http.Error(w, "路径不存在", http.StatusNotFound)
	default:
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
	}
}
