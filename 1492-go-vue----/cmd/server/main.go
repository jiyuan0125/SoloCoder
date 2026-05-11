package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"renovation-management/internal/core"
)

var ErrMissingID = errors.New("missing id")

type Server struct {
	store *core.Store
	mux   *http.ServeMux
}

func NewServer() *Server {
	s := &Server{
		store: core.NewStore(),
		mux:   http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/houses", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.handleCreateHouse(w, r)
		case http.MethodGet:
			s.handleListHouses(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	s.mux.HandleFunc("/houses/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleGetHouse(w, r)
	})

	s.mux.HandleFunc("/designs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleCreateDesign(w, r)
	})

	s.mux.HandleFunc("/designs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleGetDesign(w, r)
	})

	s.mux.HandleFunc("/designs/confirm", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleConfirmDesign(w, r)
	})

	s.mux.HandleFunc("/quotations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.handleCreateQuotation(w, r)
		case http.MethodGet:
			s.handleListQuotations(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	s.mux.HandleFunc("/quotations/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleGetQuotation(w, r)
	})

	s.mux.HandleFunc("/workitems/quantity", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleUpdateQuantity(w, r)
	})

	s.mux.HandleFunc("/changes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleCreateChangeOrder(w, r)
	})

	s.mux.HandleFunc("/changes/confirm", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleConfirmChangeOrder(w, r)
	})

	s.mux.HandleFunc("/phases", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleCreatePhase(w, r)
	})

	s.mux.HandleFunc("/phases/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleGetPhase(w, r)
	})

	s.mux.HandleFunc("/phases/progress", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleUpdateProgress(w, r)
	})

	s.mux.HandleFunc("/phases/delayed", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleListDelayedPhases(w, r)
	})

	s.mux.HandleFunc("/settlements/calculate/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleCalculateSettlement(w, r)
	})

	s.mux.HandleFunc("/settlements/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleGetSettlement(w, r)
	})
}

func getPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return ":" + port
	}

	flagPort := flag.String("port", "8080", "server port")
	flag.Parse()

	return ":" + *flagPort
}

func main() {
	server := NewServer()
	addr := getPort()

	fmt.Printf("server starting on %s\n", addr)

	if err := http.ListenAndServe(addr, server.mux); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
