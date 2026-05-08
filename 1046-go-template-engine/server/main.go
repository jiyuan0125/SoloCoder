package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"template-engine/common"
	"template-engine/template"
)

type TemplateStore struct {
	mu        sync.RWMutex
	templates map[string]*template.ParsedTemplate
	source    map[string]string
}

func NewTemplateStore() *TemplateStore {
	return &TemplateStore{
		templates: make(map[string]*template.ParsedTemplate),
		source:    make(map[string]string),
	}
}

func (s *TemplateStore) Register(name string, t *template.ParsedTemplate, source string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates[name] = t
	s.source[name] = source
}

func (s *TemplateStore) Get(name string) (*template.ParsedTemplate, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.templates[name]
	return t, ok
}

func (s *TemplateStore) Delete(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.templates[name]; ok {
		delete(s.templates, name)
		delete(s.source, name)
		return true
	}
	return false
}

func (s *TemplateStore) List() []common.TemplateInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]common.TemplateInfo, 0, len(s.templates))
	for name := range s.templates {
		list = append(list, common.TemplateInfo{Name: name})
	}
	return list
}

type Server struct {
	engine *template.Engine
	store  *TemplateStore
}

func NewServer() *Server {
	return &Server{
		engine: template.New(nil),
		store:  NewTemplateStore(),
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) handleRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.RenderResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req common.RenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.RenderResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}

	config := template.DefaultConfig()
	if req.Options != nil {
		config.StrictMissing = req.Options.StrictMissing
	}
	engine := template.New(config)

	result, err := engine.Render(req.Template, req.Data)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.RenderResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.RenderResponse{
		Success: true,
		Result:  result,
	})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.TemplateRegisterResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req common.TemplateRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.TemplateRegisterResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}

	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, common.TemplateRegisterResponse{
			Success: false,
			Error:   "template name is required",
		})
		return
	}

	parsed, err := s.engine.Parse(req.Template)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.TemplateRegisterResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	s.store.Register(req.Name, parsed, req.Template)
	writeJSON(w, http.StatusOK, common.TemplateRegisterResponse{
		Success: true,
	})
}

func (s *Server) handleRenderByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.RenderResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req common.TemplateRenderByNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.RenderResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}

	tpl, ok := s.store.Get(req.Name)
	if !ok {
		writeJSON(w, http.StatusNotFound, common.RenderResponse{
			Success: false,
			Error:   fmt.Sprintf("template '%s' not found", req.Name),
		})
		return
	}

	config := template.DefaultConfig()
	if req.Options != nil {
		config.StrictMissing = req.Options.StrictMissing
	}

	var engine *template.Engine
	if req.Options != nil {
		engine = template.New(config)
		source, _ := s.store.source[req.Name]
		var err error
		tpl, err = engine.Parse(source)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, common.RenderResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
	}

	result, err := tpl.Render(req.Data)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.RenderResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.RenderResponse{
		Success: true,
		Result:  result,
	})
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.TemplateListResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	templates := s.store.List()
	writeJSON(w, http.StatusOK, common.TemplateListResponse{
		Success:   true,
		Templates: templates,
	})
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, common.TemplateRegisterResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		writeJSON(w, http.StatusBadRequest, common.TemplateRegisterResponse{
			Success: false,
			Error:   "name parameter is required",
		})
		return
	}

	if !s.store.Delete(name) {
		writeJSON(w, http.StatusNotFound, common.TemplateRegisterResponse{
			Success: false,
			Error:   fmt.Sprintf("template '%s' not found", name),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.TemplateRegisterResponse{
		Success: true,
	})
}

func main() {
	server := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/render", server.handleRender)
	mux.HandleFunc("/templates/register", server.handleRegister)
	mux.HandleFunc("/templates/render", server.handleRenderByName)
	mux.HandleFunc("/templates/list", server.handleList)
	mux.HandleFunc("/templates/delete", server.handleDelete)

	addr := ":8080"
	log.Printf("template server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
