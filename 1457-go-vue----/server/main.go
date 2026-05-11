package main

import (
	"complaint-system/common"
	"complaint-system/core"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Server struct {
	service *core.Service
}

func NewServer() *Server {
	return &Server{
		service: core.NewService(),
	}
}

func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) respondError(w http.ResponseWriter, status int, err error) {
	s.respondJSON(w, status, common.ErrorResponse{Error: err.Error()})
}

func (s *Server) handleCreateTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req common.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, err)
		return
	}

	ticketNo, err := s.service.CreateTicket(&req)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)
		return
	}

	s.respondJSON(w, http.StatusCreated, common.CreateTicketResponse{TicketNo: ticketNo})
}

func (s *Server) handleGetTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	ticketNo := strings.TrimPrefix(r.URL.Path, "/api/tickets/")
	if ticketNo == "" || ticketNo == "/" {
		s.respondError(w, http.StatusBadRequest, fmt.Errorf("ticket no is required"))
		return
	}

	ticket, err := s.service.GetTicket(ticketNo)
	if err != nil {
		if err == core.ErrTicketNotFound {
			s.respondError(w, http.StatusNotFound, err)
		} else {
			s.respondError(w, http.StatusInternalServerError, err)
		}
		return
	}

	s.respondJSON(w, http.StatusOK, ticket)
}

func (s *Server) handleListTickets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	tickets := s.service.ListTickets()
	s.respondJSON(w, http.StatusOK, common.ListTicketsResponse{Tickets: tickets})
}

func (s *Server) handleDispatchTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	ticketNo := strings.TrimPrefix(r.URL.Path, "/api/tickets/")
	ticketNo = strings.TrimSuffix(ticketNo, "/dispatch")
	if ticketNo == "" {
		s.respondError(w, http.StatusBadRequest, fmt.Errorf("ticket no is required"))
		return
	}

	var req common.DispatchTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.service.DispatchTicket(ticketNo, &req); err != nil {
		if err == core.ErrTicketNotFound {
			s.respondError(w, http.StatusNotFound, err)
		} else if err == core.ErrInvalidStatus {
			s.respondError(w, http.StatusBadRequest, err)
		} else {
			s.respondError(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleStartProcessing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	ticketNo := strings.TrimPrefix(r.URL.Path, "/api/tickets/")
	ticketNo = strings.TrimSuffix(ticketNo, "/start-processing")
	if ticketNo == "" {
		s.respondError(w, http.StatusBadRequest, fmt.Errorf("ticket no is required"))
		return
	}

	if err := s.service.StartProcessing(ticketNo); err != nil {
		if err == core.ErrTicketNotFound {
			s.respondError(w, http.StatusNotFound, err)
		} else if err == core.ErrInvalidStatus {
			s.respondError(w, http.StatusBadRequest, err)
		} else {
			s.respondError(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleCompleteProcessing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	ticketNo := strings.TrimPrefix(r.URL.Path, "/api/tickets/")
	ticketNo = strings.TrimSuffix(ticketNo, "/complete-processing")
	if ticketNo == "" {
		s.respondError(w, http.StatusBadRequest, fmt.Errorf("ticket no is required"))
		return
	}

	var req common.CompleteProcessingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.service.CompleteProcessing(ticketNo, &req); err != nil {
		if err == core.ErrTicketNotFound {
			s.respondError(w, http.StatusNotFound, err)
		} else if err == core.ErrInvalidStatus {
			s.respondError(w, http.StatusBadRequest, err)
		} else {
			s.respondError(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleReviewTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	ticketNo := strings.TrimPrefix(r.URL.Path, "/api/tickets/")
	ticketNo = strings.TrimSuffix(ticketNo, "/review")
	if ticketNo == "" {
		s.respondError(w, http.StatusBadRequest, fmt.Errorf("ticket no is required"))
		return
	}

	var req common.ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.service.ReviewTicket(ticketNo, &req); err != nil {
		if err == core.ErrTicketNotFound {
			s.respondError(w, http.StatusNotFound, err)
		} else if err == core.ErrInvalidStatus {
			s.respondError(w, http.StatusBadRequest, err)
		} else {
			s.respondError(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	stats := s.service.GetStatistics()
	s.respondJSON(w, http.StatusOK, stats)
}

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "server port")
	flag.Parse()

	if port == 0 {
		envPort := os.Getenv("COMPLAINT_SERVER_PORT")
		if envPort != "" {
			fmt.Sscanf(envPort, "%d", &port)
		}
	}

	if port == 0 {
		port = 8080
	}

	server := NewServer()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/tickets", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tickets" {
			if strings.HasPrefix(r.URL.Path, "/api/tickets/") {
				if strings.Contains(r.URL.Path, "/dispatch") {
					server.handleDispatchTicket(w, r)
				} else if strings.Contains(r.URL.Path, "/start-processing") {
					server.handleStartProcessing(w, r)
				} else if strings.Contains(r.URL.Path, "/complete-processing") {
					server.handleCompleteProcessing(w, r)
				} else if strings.Contains(r.URL.Path, "/review") {
					server.handleReviewTicket(w, r)
				} else {
					server.handleGetTicket(w, r)
				}
			} else {
				http.NotFound(w, r)
			}
			return
		}
		if r.Method == http.MethodPost {
			server.handleCreateTicket(w, r)
		} else if r.Method == http.MethodGet {
			server.handleListTickets(w, r)
		} else {
			server.respondError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		}
	})

	mux.HandleFunc("/api/statistics", server.handleStatistics)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server starting on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
