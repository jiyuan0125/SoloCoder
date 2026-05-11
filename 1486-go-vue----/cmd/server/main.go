package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

	"locker/pkg/api"
	"locker/pkg/core"
)

type Server struct {
	system *core.System
}

func NewServer() *Server {
	sys := core.NewSystem()
	sys.StartOverdueScanner()
	return &Server{system: sys}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (s *Server) writeError(w http.ResponseWriter, status int, err error) {
	s.writeJSON(w, status, api.ErrorResponse{Error: err.Error()})
}

func (s *Server) handleCreateLocker(w http.ResponseWriter, r *http.Request) {
	var req api.CreateLockerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.system.CreateLocker(req); err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeJSON(w, http.StatusCreated, api.SuccessResponse{Success: true})
}

func (s *Server) handleListCompartments(w http.ResponseWriter, r *http.Request) {
	var req api.ListCompartmentsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	resp, err := s.system.ListCompartments(req)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err)
		return
	}
	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleCreateCourier(w http.ResponseWriter, r *http.Request) {
	var req api.CreateCourierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.system.CreateCourier(req); err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeJSON(w, http.StatusCreated, api.SuccessResponse{Success: true})
}

func (s *Server) handleGetCourier(w http.ResponseWriter, r *http.Request) {
	var req api.GetCourierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	resp, err := s.system.GetCourier(req)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err)
		return
	}
	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleRechargeCourier(w http.ResponseWriter, r *http.Request) {
	var req api.RechargeCourierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	resp, err := s.system.RechargeCourier(req)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleStorePackage(w http.ResponseWriter, r *http.Request) {
	var req api.StorePackageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	resp, err := s.system.StorePackage(req)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handlePickupPackage(w http.ResponseWriter, r *http.Request) {
	var req api.PickupPackageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	resp, err := s.system.PickupPackage(req)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleListOverdue(w http.ResponseWriter, r *http.Request) {
	resp, err := s.system.ListOverdue()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleHandleOverdue(w http.ResponseWriter, r *http.Request) {
	var req api.HandleOverdueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.system.HandleOverdue(req); err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeJSON(w, http.StatusOK, api.SuccessResponse{Success: true})
}

func main() {
	var port int
	flag.IntVar(&port, "port", 9006, "Server port")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	server := NewServer()
	mux := http.NewServeMux()

	mux.HandleFunc("/locker/create", server.handleCreateLocker)
	mux.HandleFunc("/locker/compartments", server.handleListCompartments)
	mux.HandleFunc("/courier/create", server.handleCreateCourier)
	mux.HandleFunc("/courier/get", server.handleGetCourier)
	mux.HandleFunc("/courier/recharge", server.handleRechargeCourier)
	mux.HandleFunc("/package/store", server.handleStorePackage)
	mux.HandleFunc("/package/pickup", server.handlePickupPackage)
	mux.HandleFunc("/overdue/list", server.handleListOverdue)
	mux.HandleFunc("/overdue/handle", server.handleHandleOverdue)

	addr := ":" + strconv.Itoa(port)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
