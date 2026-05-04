package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type APIHandler struct {
	applicationService *ApplicationService
	statusService      *StatusService
	adminService       *AdminService
	userService        *UserService
}

func NewAPIHandler(
	applicationService *ApplicationService,
	statusService *StatusService,
	adminService *AdminService,
	userService *UserService,
) *APIHandler {
	return &APIHandler{
		applicationService: applicationService,
		statusService:      statusService,
		adminService:       adminService,
		userService:        userService,
	}
}

func (h *APIHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *APIHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

func (h *APIHandler) CreateClaimHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req CreateClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	claim, err := h.applicationService.CreateClaim(req)
	if err != nil {
		if ve, ok := err.(*ValidationError); ok {
			h.respondError(w, http.StatusBadRequest, ve.Message)
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, claim)
}

func (h *APIHandler) GetUserClaimsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID := strings.TrimPrefix(r.URL.Path, "/api/user/claims/")
	if userID == "" {
		h.respondError(w, http.StatusBadRequest, "User ID required")
		return
	}

	claims := h.userService.GetUserClaims(userID)
	if claims == nil {
		claims = []*WarrantyClaim{}
	}
	h.respondJSON(w, http.StatusOK, claims)
}

func (h *APIHandler) GetPendingClaimsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := h.adminService.GetPendingClaims()
	if claims == nil {
		claims = []*WarrantyClaim{}
	}
	h.respondJSON(w, http.StatusOK, claims)
}

func (h *APIHandler) GetAllClaimsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := h.adminService.GetAllClaims()
	if claims == nil {
		claims = []*WarrantyClaim{}
	}
	h.respondJSON(w, http.StatusOK, claims)
}

func (h *APIHandler) ReviewClaimHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claimID := strings.TrimPrefix(r.URL.Path, "/api/admin/review/")
	if claimID == "" {
		h.respondError(w, http.StatusBadRequest, "Claim ID required")
		return
	}

	var req ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	claim, err := h.statusService.ReviewClaim(claimID, req.Approved, req.Reason)
	if err != nil {
		if ve, ok := err.(*ValidationError); ok {
			h.respondError(w, http.StatusBadRequest, ve.Message)
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, claim)
}

func (h *APIHandler) SubmitAppealHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claimID := strings.TrimPrefix(r.URL.Path, "/api/user/appeal/")
	if claimID == "" {
		h.respondError(w, http.StatusBadRequest, "Claim ID required")
		return
	}

	var req AppealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	appeal, err := h.statusService.SubmitAppeal(claimID, req.Reason)
	if err != nil {
		if ve, ok := err.(*ValidationError); ok {
			h.respondError(w, http.StatusBadRequest, ve.Message)
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, appeal)
}

func (h *APIHandler) GetStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	stats := h.adminService.GetStatistics()
	h.respondJSON(w, http.StatusOK, stats)
}

func main() {
	storage := NewStorage("warranty_claims.json")
	
	applicationService := NewApplicationService(storage)
	statusService := NewStatusService(storage)
	adminService := NewAdminService(storage)
	userService := NewUserService(storage)
	
	handler := NewAPIHandler(applicationService, statusService, adminService, userService)
	
	http.HandleFunc("/api/claims", handler.CreateClaimHandler)
	http.HandleFunc("/api/user/claims/", handler.GetUserClaimsHandler)
	http.HandleFunc("/api/user/appeal/", handler.SubmitAppealHandler)
	http.HandleFunc("/api/admin/pending", handler.GetPendingClaimsHandler)
	http.HandleFunc("/api/admin/all", handler.GetAllClaimsHandler)
	http.HandleFunc("/api/admin/review/", handler.ReviewClaimHandler)
	http.HandleFunc("/api/admin/statistics", handler.GetStatisticsHandler)
	
	fmt.Println("Warranty Claim Service starting on :8080...")
	fmt.Println("Endpoints:")
	fmt.Println("  POST   /api/claims              - Create warranty claim")
	fmt.Println("  GET    /api/user/claims/{id}    - Get user's claims")
	fmt.Println("  POST   /api/user/appeal/{id}    - Submit appeal")
	fmt.Println("  GET    /api/admin/pending        - Get pending claims")
	fmt.Println("  GET    /api/admin/all            - Get all claims")
	fmt.Println("  POST   /api/admin/review/{id}    - Review claim")
	fmt.Println("  GET    /api/admin/statistics     - Get statistics")
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
