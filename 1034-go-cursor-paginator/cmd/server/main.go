package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"cursor-paginator/pkg/api"
	"cursor-paginator/pkg/cursorpaginator"
)

type Server struct {
	users []*api.User
	mu    sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		users: cursorpaginator.GenerateSampleUsers(50),
	}
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req := &api.PageRequest{
		Cursor:    r.URL.Query().Get("cursor"),
		Previous:  r.URL.Query().Get("previous"),
		SortField: r.URL.Query().Get("sort_field"),
		SortOrder: api.SortOrder(r.URL.Query().Get("sort_order")),
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil {
			req.Limit = limit
		}
	}

	s.mu.RLock()
	users := make([]*api.User, len(s.users))
	copy(users, s.users)
	s.mu.RUnlock()

	response, err := cursorpaginator.PaginateUsers(users, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleAddUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user api.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	user.ID = len(s.users) + 1
	s.users = append(s.users, &user)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "id parameter required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	newUsers := make([]*api.User, 0, len(s.users))
	for _, user := range s.users {
		if user.ID != id {
			newUsers = append(newUsers, user)
		}
	}
	s.users = newUsers
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	server := NewServer()

	http.HandleFunc("/users", server.handleUsers)
	http.HandleFunc("/users/add", server.handleAddUser)
	http.HandleFunc("/users/delete", server.handleDeleteUser)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
