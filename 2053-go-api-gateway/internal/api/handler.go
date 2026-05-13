package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/example/apigateway/internal/config"
	"github.com/example/apigateway/internal/middleware"
	"github.com/example/apigateway/internal/model"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: data})
}

func writeError(w http.ResponseWriter, code int, errMsg string) {
	writeJSON(w, code, APIResponse{Success: false, Error: errMsg})
}

func parseIDFromPath(path, prefix string) (int64, error) {
	trimmed := strings.TrimPrefix(path, prefix)
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 {
		return 0, http.ErrNotSupported
	}
	return strconv.ParseInt(parts[0], 10, 64)
}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/routes", handleRoutes)
	mux.HandleFunc("/api/routes/", handleRouteByID)

	mux.HandleFunc("/api/changes", handleChanges)
	mux.HandleFunc("/api/changes/", handleChangeOperations)

	mux.HandleFunc("/api/api-keys", handleAPIKeys)
	mux.HandleFunc("/api/api-keys/", handleAPIKeyByID)

	mux.HandleFunc("/api/jwt-secret", handleJWTSecret)
	mux.HandleFunc("/api/jwt-token", handleJWTToken)
}

func handleRoutes(w http.ResponseWriter, r *http.Request) {
	mgr := config.GetConfigManager()

	switch r.Method {
	case http.MethodGet:
		routes, err := mgr.ListRoutes()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeSuccess(w, routes)

	case http.MethodPost:
		var req struct {
			Name         string            `json:"name"`
			Path         string            `json:"path"`
			MatchType    string            `json:"match_type"`
			Headers      map[string]string `json:"headers"`
			TargetURL    string            `json:"target_url"`
			Timeout      int               `json:"timeout"`
			AuthType     string            `json:"auth_type"`
			RateLimitIP  int               `json:"rate_limit_ip"`
			RateLimitKey int               `json:"rate_limit_key"`
			Applicant    string            `json:"applicant"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		newRule := &model.RouteRule{
			Name:         req.Name,
			Path:         req.Path,
			MatchType:    model.RouteMatchType(req.MatchType),
			Headers:      req.Headers,
			TargetURL:    req.TargetURL,
			Timeout:      time.Duration(req.Timeout) * time.Second,
			AuthType:     model.AuthType(req.AuthType),
			RateLimitIP:  req.RateLimitIP,
			RateLimitKey: req.RateLimitKey,
		}

		newConfig, _ := json.Marshal(newRule)

		change := &model.RouteChange{
			Operation: "create",
			NewConfig: string(newConfig),
			Status:    model.StatusDraft,
			Applicant: req.Applicant,
		}

		id, err := mgr.CreateRouteChange(change)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeSuccess(w, map[string]int64{"change_id": id})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleRouteByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, err := parseIDFromPath(r.URL.Path, "/api/routes/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid route id")
		return
	}

	mgr := config.GetConfigManager()
	route, err := mgr.GetRoute(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if route == nil {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}

	writeSuccess(w, route)
}

func handleChanges(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	mgr := config.GetConfigManager()
	changes, err := mgr.ListRouteChanges()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, changes)
}

func handleChangeOperations(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/api/changes/submit" {
		handleSubmitChange(w, r)
		return
	}
	if path == "/api/changes/approve1" {
		handleApproval1(w, r, true)
		return
	}
	if path == "/api/changes/approve2" {
		handleApproval2(w, r, true)
		return
	}
	if path == "/api/changes/reject1" {
		handleApproval1(w, r, false)
		return
	}
	if path == "/api/changes/reject2" {
		handleApproval2(w, r, false)
		return
	}
	if path == "/api/changes/final-approve" {
		handleFinalApprove(w, r)
		return
	}
	if path == "/api/changes/update" {
		handleUpdateChange(w, r)
		return
	}
	if path == "/api/changes/delete" {
		handleDeleteChange(w, r)
		return
	}

	if strings.HasPrefix(path, "/api/changes/") {
		id, err := parseIDFromPath(path, "/api/changes/")
		if err == nil && r.Method == http.MethodGet {
			mgr := config.GetConfigManager()
			change, err := mgr.GetRouteChange(id)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if change == nil {
				writeError(w, http.StatusNotFound, "change not found")
				return
			}
			writeSuccess(w, change)
			return
		}
	}

	writeError(w, http.StatusNotFound, "operation not found")
}

func handleSubmitChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		ChangeID  int64  `json:"change_id"`
		Applicant string `json:"applicant"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	mgr := config.GetConfigManager()
	if err := mgr.SubmitForApproval(req.ChangeID, req.Applicant); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeSuccess(w, map[string]string{"status": "submitted"})
}

func handleApproval1(w http.ResponseWriter, r *http.Request, approved bool) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		ChangeID int64  `json:"change_id"`
		Approver string `json:"approver"`
		Reason   string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	mgr := config.GetConfigManager()
	if err := mgr.Approval1(req.ChangeID, req.Approver, approved, req.Reason); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if approved {
		writeSuccess(w, map[string]string{"status": "approved_1"})
	} else {
		writeSuccess(w, map[string]string{"status": "rejected_1"})
	}
}

func handleApproval2(w http.ResponseWriter, r *http.Request, approved bool) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		ChangeID int64  `json:"change_id"`
		Approver string `json:"approver"`
		Reason   string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	mgr := config.GetConfigManager()
	if err := mgr.Approval2(req.ChangeID, req.Approver, approved, req.Reason); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if approved {
		writeSuccess(w, map[string]string{"status": "approved_2"})
	} else {
		writeSuccess(w, map[string]string{"status": "rejected_2"})
	}
}

func handleFinalApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		ChangeID int64  `json:"change_id"`
		Approver string `json:"approver"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	mgr := config.GetConfigManager()
	if err := mgr.FinalApproveAndApply(req.ChangeID, req.Approver); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeSuccess(w, map[string]string{"status": "applied"})
}

func handleUpdateChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		ChangeID     int64             `json:"change_id"`
		Name         string            `json:"name"`
		Path         string            `json:"path"`
		MatchType    string            `json:"match_type"`
		Headers      map[string]string `json:"headers"`
		TargetURL    string            `json:"target_url"`
		Timeout      int               `json:"timeout"`
		AuthType     string            `json:"auth_type"`
		RateLimitIP  int               `json:"rate_limit_ip"`
		RateLimitKey int               `json:"rate_limit_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	mgr := config.GetConfigManager()
	change, err := mgr.GetRouteChange(req.ChangeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if change == nil {
		writeError(w, http.StatusNotFound, "change not found")
		return
	}
	if change.Status != model.StatusDraft {
		writeError(w, http.StatusBadRequest, "can only update draft changes")
		return
	}

	newRule := &model.RouteRule{
		Name:         req.Name,
		Path:         req.Path,
		MatchType:    model.RouteMatchType(req.MatchType),
		Headers:      req.Headers,
		TargetURL:    req.TargetURL,
		Timeout:      time.Duration(req.Timeout) * time.Second,
		AuthType:     model.AuthType(req.AuthType),
		RateLimitIP:  req.RateLimitIP,
		RateLimitKey: req.RateLimitKey,
	}
	newConfig, _ := json.Marshal(newRule)
	change.NewConfig = string(newConfig)

	if err := mgr.UpdateRouteChange(change); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, map[string]string{"status": "updated"})
}

func handleDeleteChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		RouteID   int64  `json:"route_id"`
		Applicant string `json:"applicant"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	mgr := config.GetConfigManager()
	oldRoute, err := mgr.GetRoute(req.RouteID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if oldRoute == nil {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}

	oldConfig, _ := json.Marshal(oldRoute)

	change := &model.RouteChange{
		RouteID:   req.RouteID,
		Operation: "delete",
		OldConfig: string(oldConfig),
		Status:    model.StatusDraft,
		Applicant: req.Applicant,
	}

	id, err := mgr.CreateRouteChange(change)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, map[string]int64{"change_id": id})
}

func handleAPIKeys(w http.ResponseWriter, r *http.Request) {
	mgr := config.GetConfigManager()

	switch r.Method {
	case http.MethodGet:
		keys, err := mgr.ListAPIKeys()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeSuccess(w, keys)

	case http.MethodPost:
		var req struct {
			UserID    string  `json:"user_id"`
			ExpiresAt *string `json:"expires_at"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		newKey := middleware.GenerateAPIKey()
		apiKey := &model.APIKey{
			Key:    newKey,
			UserID: req.UserID,
		}

		if req.ExpiresAt != nil {
			if t, err := time.Parse(time.RFC3339, *req.ExpiresAt); err == nil {
				apiKey.ExpiresAt = &t
			}
		}

		_, err := mgr.CreateAPIKey(apiKey)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeSuccess(w, map[string]string{"api_key": newKey})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleAPIKeyByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, err := parseIDFromPath(r.URL.Path, "/api/api-keys/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid api key id")
		return
	}

	mgr := config.GetConfigManager()
	if err := mgr.DeleteAPIKey(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, map[string]string{"status": "deleted"})
}

func handleJWTSecret(w http.ResponseWriter, r *http.Request) {
	mgr := config.GetConfigManager()

	switch r.Method {
	case http.MethodGet:
		secret, err := mgr.GetJWTSecret()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if secret == nil {
			writeError(w, http.StatusNotFound, "jwt secret not configured")
			return
		}
		writeSuccess(w, map[string]string{
			"issuer":     secret.Issuer,
			"created_at": secret.CreatedAt.Format(time.RFC3339),
		})

	case http.MethodPost:
		var req struct {
			Secret string `json:"secret"`
			Issuer string `json:"issuer"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Secret == "" {
			req.Secret = uuid.New().String() + "-" + uuid.New().String()
		}

		_, err := mgr.CreateJWTSecret(&model.JWTSecret{
			Secret: req.Secret,
			Issuer: req.Issuer,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeSuccess(w, map[string]string{
			"secret": req.Secret,
			"issuer": req.Issuer,
		})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleJWTToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		UserID string `json:"user_id"`
		Expire int    `json:"expire_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	mgr := config.GetConfigManager()
	secret, err := mgr.GetJWTSecret()
	if err != nil || secret == nil {
		writeError(w, http.StatusInternalServerError, "jwt secret not configured")
		return
	}

	if req.Expire <= 0 {
		req.Expire = 24
	}

	claims := jwt.MapClaims{
		"sub": req.UserID,
		"iss": secret.Issuer,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Duration(req.Expire) * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret.Secret))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, map[string]string{"token": tokenStr})
}
