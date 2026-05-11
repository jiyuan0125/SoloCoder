package main

import (
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"os"
	"strings"

	"tender-management/common"
	"tender-management/core"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "Server port (default: 8080)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("TENDER_PORT")
	}
	if port == "" {
		port = "8080"
	}

	service := core.NewTenderService()
	handler := NewHandler(service)

	mux := http.NewServeMux()

	mux.HandleFunc("/projects", handler.ListProjects)
	mux.HandleFunc("/projects/create", handler.CreateProject)
	mux.HandleFunc("/projects/", handler.ProjectHandler)

	mux.HandleFunc("/bids", handler.SubmitBid)
	mux.HandleFunc("/bids/", handler.BidHandler)

	mux.HandleFunc("/opening", handler.OpenProject)
	mux.HandleFunc("/opening/", handler.OpeningHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	println("Server starting on port", port)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

type Handler struct {
	service *core.TenderService
}

func NewHandler(service *core.TenderService) *Handler {
	return &Handler{service: service}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("method not allowed"))
		return
	}

	supplierID := r.URL.Query().Get("supplier_id")
	role := r.URL.Query().Get("role")

	var projects []common.Project
	if role == "supplier" && supplierID != "" {
		projects = h.service.ListProjectsForSupplier(supplierID)
	} else {
		projects = h.service.ListAllProjects()
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(common.ProjectListData{Projects: projects}))
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("method not allowed"))
		return
	}

	var req common.CreateProjectRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("invalid request body"))
		return
	}

	project, err := h.service.CreateProject(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusCreated, common.NewSuccessResponse(common.ProjectDetailData{Project: *project}))
}

func (h *Handler) ProjectHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/projects/")
	parts := strings.Split(path, "/")
	projectID := parts[0]

	if len(parts) == 1 && parts[0] != "" {
		h.handleSingleProject(w, r, projectID)
		return
	}

	if len(parts) == 2 {
		switch parts[1] {
		case "update":
			h.handleUpdateProject(w, r, projectID)
		case "publish":
			h.handlePublishProject(w, r, projectID)
		case "bids":
			h.handleProjectBids(w, r, projectID)
		default:
			writeJSON(w, http.StatusNotFound, common.NewErrorResponse("not found"))
		}
		return
	}

	writeJSON(w, http.StatusNotFound, common.NewErrorResponse("not found"))
}

func (h *Handler) handleSingleProject(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("method not allowed"))
		return
	}

	project, err := h.service.GetProject(projectID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(common.ProjectDetailData{Project: *project}))
}

func (h *Handler) handleUpdateProject(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("method not allowed"))
		return
	}

	var req common.UpdateProjectRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("invalid request body"))
		return
	}

	if err := h.service.UpdateProject(projectID, req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	project, _ := h.service.GetProject(projectID)
	writeJSON(w, http.StatusOK, common.NewSuccessResponse(common.ProjectDetailData{Project: *project}))
}

func (h *Handler) handlePublishProject(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("method not allowed"))
		return
	}

	if err := h.service.PublishProject(projectID); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	project, _ := h.service.GetProject(projectID)
	writeJSON(w, http.StatusOK, common.NewSuccessResponse(common.ProjectDetailData{Project: *project}))
}

func (h *Handler) handleProjectBids(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("method not allowed"))
		return
	}

	bids := h.service.ListBidsByProject(projectID)
	writeJSON(w, http.StatusOK, common.NewSuccessResponse(map[string][]common.Bid{"bids": bids}))
}

func (h *Handler) SubmitBid(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("method not allowed"))
		return
	}

	var req common.SubmitBidRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("invalid request body"))
		return
	}

	bid, err := h.service.SubmitBid(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusCreated, common.NewSuccessResponse(common.BidDetailData{Bid: *bid}))
}

func (h *Handler) BidHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/bids/")
	parts := strings.Split(path, "/")

	if len(parts) == 1 {
		bidID := parts[0]
		h.handleGetBid(w, r, bidID)
		return
	}

	if len(parts) == 2 && parts[1] == "update" {
		bidID := parts[0]
		h.handleUpdateBid(w, r, bidID)
		return
	}

	if len(parts) == 3 && parts[1] == "supplier" {
		projectID := parts[0]
		supplierID := parts[2]
		h.handleGetBidBySupplier(w, r, projectID, supplierID)
		return
	}

	writeJSON(w, http.StatusNotFound, common.NewErrorResponse("not found"))
}

func (h *Handler) handleGetBid(w http.ResponseWriter, r *http.Request, bidID string) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("method not allowed"))
		return
	}

	bid, err := h.service.GetBid(bidID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(common.BidDetailData{Bid: *bid}))
}

func (h *Handler) handleUpdateBid(w http.ResponseWriter, r *http.Request, bidID string) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("method not allowed"))
		return
	}

	var req common.UpdateBidRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("invalid request body"))
		return
	}

	bid, err := h.service.UpdateBid(bidID, req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(common.BidDetailData{Bid: *bid}))
}

func (h *Handler) handleGetBidBySupplier(w http.ResponseWriter, r *http.Request, projectID, supplierID string) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("method not allowed"))
		return
	}

	bid, err := h.service.GetBidBySupplier(projectID, supplierID)
	if err != nil {
		if errors.Is(err, core.ErrBidNotFound) {
			writeJSON(w, http.StatusNotFound, common.NewErrorResponse("bid not found"))
			return
		}
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(common.BidDetailData{Bid: *bid}))
}

func (h *Handler) OpenProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("method not allowed"))
		return
	}

	var req common.OpenProjectRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("invalid request body"))
		return
	}

	result, err := h.service.OpenProject(req.ProjectID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(common.OpeningResultData{Result: *result}))
}

func (h *Handler) OpeningHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.NewErrorResponse("method not allowed"))
		return
	}

	projectID := strings.TrimPrefix(r.URL.Path, "/opening/")
	if projectID == "" {
		writeJSON(w, http.StatusBadRequest, common.NewErrorResponse("project id required"))
		return
	}

	result, err := h.service.GetOpeningResult(projectID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, common.NewErrorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(common.OpeningResultData{Result: *result}))
}
