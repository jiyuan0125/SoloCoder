package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"sync"

	"cuckoo/pkg/common"
	"cuckoo/pkg/cuckoo"
)

type Manager struct {
	mu       sync.RWMutex
	filters  map[string]*cuckoo.Filter
}

func NewManager() *Manager {
	return &Manager{
		filters: make(map[string]*cuckoo.Filter),
	}
}

func (m *Manager) getFilter(name string) (*cuckoo.Filter, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	filter, ok := m.filters[name]
	return filter, ok
}

func (m *Manager) setFilter(name string, filter *cuckoo.Filter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.filters[name] = filter
}

func (m *Manager) deleteFilter(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.filters, name)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func main() {
	var port string
	flag.StringVar(&port, "port", "", "HTTP server port (default: 8080)")
	flag.Parse()

	if port == "" {
		if envPort := os.Getenv("PORT"); envPort != "" {
			port = envPort
		} else {
			port = "8214"
		}
	}

	manager := NewManager()

	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, common.CreateResponse{
				Success: false,
				Message: "method not allowed",
			})
			return
		}

		var req common.CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.CreateResponse{
				Success: false,
				Message: "invalid request body",
			})
			return
		}

		if req.Name == "" {
			writeJSON(w, http.StatusBadRequest, common.CreateResponse{
				Success: false,
				Message: "name is required",
			})
			return
		}

		if req.Capacity == 0 {
			writeJSON(w, http.StatusBadRequest, common.CreateResponse{
				Success: false,
				Message: "capacity must be greater than 0",
			})
			return
		}

		if _, exists := manager.getFilter(req.Name); exists {
			writeJSON(w, http.StatusBadRequest, common.CreateResponse{
				Success: false,
				Message: "filter already exists",
			})
			return
		}

		filter := cuckoo.New(req.Capacity)
		manager.setFilter(req.Name, filter)

		writeJSON(w, http.StatusCreated, common.CreateResponse{
			Success: true,
			Message: "filter created",
		})
	})

	http.HandleFunc("/insert", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, common.InsertResponse{
				Success: false,
				Message: "method not allowed",
			})
			return
		}

		var req common.InsertRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.InsertResponse{
				Success: false,
				Message: "invalid request body",
			})
			return
		}

		if req.Name == "" {
			writeJSON(w, http.StatusBadRequest, common.InsertResponse{
				Success: false,
				Message: "name is required",
			})
			return
		}

		filter, exists := manager.getFilter(req.Name)
		if !exists {
			writeJSON(w, http.StatusNotFound, common.InsertResponse{
				Success: false,
				Message: "filter not found",
			})
			return
		}

		elements := req.Elements
		if req.Element != "" && len(elements) == 0 {
			elements = append(elements, req.Element)
		}

		inserted, failed := 0, 0
		for _, elem := range elements {
			result := filter.Insert([]byte(elem))
			if result.Success {
				inserted++
			} else {
				failed++
			}
		}

		writeJSON(w, http.StatusOK, common.InsertResponse{
			Success:  true,
			Inserted: inserted,
			Failed:   failed,
		})
	})

	http.HandleFunc("/lookup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, common.LookupResponse{
				Exists: false,
			})
			return
		}

		var req common.LookupRequest
		if r.Method == http.MethodPost {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, common.LookupResponse{
					Exists: false,
				})
				return
			}
		} else {
			req.Name = r.URL.Query().Get("name")
			req.Element = r.URL.Query().Get("element")
		}

		if req.Name == "" {
			writeJSON(w, http.StatusBadRequest, common.LookupResponse{
				Name:    req.Name,
				Element: req.Element,
				Exists:  false,
			})
			return
		}

		filter, exists := manager.getFilter(req.Name)
		if !exists {
			writeJSON(w, http.StatusNotFound, common.LookupResponse{
				Name:    req.Name,
				Element: req.Element,
				Exists:  false,
			})
			return
		}

		elementExists := filter.Lookup([]byte(req.Element))
		writeJSON(w, http.StatusOK, common.LookupResponse{
			Name:    req.Name,
			Element: req.Element,
			Exists:  elementExists,
		})
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, common.DeleteResponse{
				Success: false,
				Message: "method not allowed",
			})
			return
		}

		var req common.DeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.DeleteResponse{
				Success: false,
				Message: "invalid request body",
			})
			return
		}

		if req.Name == "" {
			writeJSON(w, http.StatusBadRequest, common.DeleteResponse{
				Success: false,
				Message: "name is required",
			})
			return
		}

		filter, exists := manager.getFilter(req.Name)
		if !exists {
			writeJSON(w, http.StatusNotFound, common.DeleteResponse{
				Success: false,
				Message: "filter not found",
			})
			return
		}

		result := filter.Delete([]byte(req.Element))
		if result.Success {
			writeJSON(w, http.StatusOK, common.DeleteResponse{
				Success: true,
				Message: "deleted",
			})
		} else {
			writeJSON(w, http.StatusNotFound, common.DeleteResponse{
				Success: false,
				Message: "element not found",
			})
		}
	})

	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, common.InfoResponse{
				Exists: false,
			})
			return
		}

		var req common.InfoRequest
		if r.Method == http.MethodPost {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, common.InfoResponse{
					Exists: false,
				})
				return
			}
		} else {
			req.Name = r.URL.Query().Get("name")
		}

		if req.Name == "" {
			writeJSON(w, http.StatusBadRequest, common.InfoResponse{
				Name:   req.Name,
				Exists: false,
			})
			return
		}

		filter, exists := manager.getFilter(req.Name)
		if !exists {
			writeJSON(w, http.StatusNotFound, common.InfoResponse{
				Name:   req.Name,
				Exists: false,
			})
			return
		}

		stats := filter.Stats()
		writeJSON(w, http.StatusOK, common.InfoResponse{
			Name:       req.Name,
			Exists:     true,
			UsedSlots:  stats.UsedSlots,
			TotalSlots: stats.TotalSlots,
			LoadRate:   stats.LoadRate,
		})
	})

	log.Printf("Cuckoo filter server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
