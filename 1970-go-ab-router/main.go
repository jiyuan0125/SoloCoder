package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Experiment struct {
	Name        string
	Description string
	Ratio       int
	CreatedAt   time.Time
	Active      bool
}

type UserAssignment struct {
	UserID      string
	GroupName   string
	AssignedAt  time.Time
}

type EventRecord struct {
	EventName  string
	Timestamp  time.Time
}

type GroupStats struct {
	GroupName     string
	UserCount     int
	ConversionCount int
	ConversionRate  float64
}

type TimeWindowStats struct {
	Window   string
	Control  GroupStats
	Experiment GroupStats
}

type StatsResponse struct {
	ExperimentName string
	Stats          []TimeWindowStats
}

type ExperimentStore struct {
	mu            sync.RWMutex
	experiments   map[string]*Experiment
	assignments   map[string]map[string]*UserAssignment
	events        map[string]map[string][]EventRecord
}

func NewExperimentStore() *ExperimentStore {
	return &ExperimentStore{
		experiments: make(map[string]*Experiment),
		assignments: make(map[string]map[string]*UserAssignment),
		events:      make(map[string]map[string][]EventRecord),
	}
}

func hashUserID(userID string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(userID))
	return h.Sum64()
}

func getBucket(hashVal uint64) int {
	return int(hashVal % 100)
}

func (s *ExperimentStore) CreateExperiment(name, description string, ratio int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ratio < 0 || ratio > 100 {
		return fmt.Errorf("ratio must be between 0 and 100")
	}

	if _, exists := s.experiments[name]; exists {
		return fmt.Errorf("experiment '%s' already exists", name)
	}

	s.experiments[name] = &Experiment{
		Name:        name,
		Description: description,
		Ratio:       ratio,
		CreatedAt:   time.Now(),
		Active:      true,
	}
	s.assignments[name] = make(map[string]*UserAssignment)
	s.events[name] = make(map[string][]EventRecord)
	return nil
}

func (s *ExperimentStore) UpdateRatio(name string, newRatio int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if newRatio < 0 || newRatio > 100 {
		return fmt.Errorf("ratio must be between 0 and 100")
	}

	exp, exists := s.experiments[name]
	if !exists {
		return fmt.Errorf("experiment '%s' not found", name)
	}
	if !exp.Active {
		return fmt.Errorf("experiment '%s' is not active", name)
	}

	exp.Ratio = newRatio
	return nil
}

func (s *ExperimentStore) DeleteExperiment(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.experiments[name]; !exists {
		return fmt.Errorf("experiment '%s' not found", name)
	}

	s.experiments[name].Active = false
	return nil
}

func (s *ExperimentStore) ListActiveExperiments() []*Experiment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	active := make([]*Experiment, 0)
	for _, exp := range s.experiments {
		if exp.Active {
			active = append(active, exp)
		}
	}
	return active
}

func (s *ExperimentStore) GetOrAssignUser(name, userID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	exp, exists := s.experiments[name]
	if !exists {
		return "", fmt.Errorf("experiment '%s' not found", name)
	}
	if !exp.Active {
		return "", fmt.Errorf("experiment '%s' is not active", name)
	}

	if assignments, ok := s.assignments[name]; ok {
		if assignment, ok := assignments[userID]; ok {
			return assignment.GroupName, nil
		}
	}

	hashVal := hashUserID(userID)
	bucket := getBucket(hashVal)
	var group string
	if bucket < exp.Ratio {
		group = "experiment"
	} else {
		group = "control"
	}

	s.assignments[name][userID] = &UserAssignment{
		UserID:     userID,
		GroupName:  group,
		AssignedAt: time.Now(),
	}
	return group, nil
}

func (s *ExperimentStore) RecordEvent(name, userID, eventName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.experiments[name]
	if !exists {
		return fmt.Errorf("experiment '%s' not found", name)
	}

	if _, ok := s.events[name]; !ok {
		s.events[name] = make(map[string][]EventRecord)
	}

	s.events[name][userID] = append(s.events[name][userID], EventRecord{
		EventName: eventName,
		Timestamp: time.Now(),
	})
	return nil
}

func (s *ExperimentStore) GetStats(name string) (*StatsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.experiments[name]
	if !exists {
		return nil, fmt.Errorf("experiment '%s' not found", name)
	}

	now := time.Now()
	windows := []struct {
		name     string
		duration time.Duration
	}{
		{"1h", 1 * time.Hour},
		{"24h", 24 * time.Hour},
		{"7d", 7 * 24 * time.Hour},
	}

	result := &StatsResponse{
		ExperimentName: name,
		Stats:          make([]TimeWindowStats, 0),
	}

	for _, window := range windows {
		cutoff := now.Add(-window.duration)

		controlUsers := make(map[string]bool)
		experimentUsers := make(map[string]bool)
		controlConversions := 0
		experimentConversions := 0

		if assignments, ok := s.assignments[name]; ok {
			for userID, assignment := range assignments {
				if assignment.AssignedAt.After(cutoff) {
					if assignment.GroupName == "control" {
						controlUsers[userID] = true
					} else {
						experimentUsers[userID] = true
					}
				}
			}
		}

		if events, ok := s.events[name]; ok {
			for userID, userEvents := range events {
				for _, event := range userEvents {
					if event.Timestamp.After(cutoff) {
						if controlUsers[userID] {
							controlConversions++
						}
						if experimentUsers[userID] {
							experimentConversions++
						}
						break
					}
				}
			}
		}

		controlCount := len(controlUsers)
		expCount := len(experimentUsers)

		var controlRate, expRate float64
		if controlCount > 0 {
			controlRate = float64(controlConversions) / float64(controlCount)
		}
		if expCount > 0 {
			expRate = float64(experimentConversions) / float64(expCount)
		}

		result.Stats = append(result.Stats, TimeWindowStats{
			Window: window.name,
			Control: GroupStats{
				GroupName:       "control",
				UserCount:       controlCount,
				ConversionCount: controlConversions,
				ConversionRate:  controlRate,
			},
			Experiment: GroupStats{
				GroupName:       "experiment",
				UserCount:       expCount,
				ConversionCount: experimentConversions,
				ConversionRate:  expRate,
			},
		})
	}

	return result, nil
}

type CreateExperimentRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Ratio       int    `json:"ratio"`
}

type UpdateRatioRequest struct {
	Ratio int `json:"ratio"`
}

type EventRequest struct {
	UserID    string `json:"user_id"`
	EventName string `json:"event_name"`
}

type ExperimentListResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Ratio       int    `json:"ratio"`
	CreatedAt   string `json:"created_at"`
}

type AssignmentResponse struct {
	UserID  string `json:"user_id"`
	Group   string `json:"group"`
}

func jsonError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func jsonResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

type Router struct {
	store *ExperimentStore
}

func NewRouter(store *ExperimentStore) *Router {
	return &Router{store: store}
}

var (
	experimentsRe    = regexp.MustCompile(`^/experiments$`)
	experimentNameRe = regexp.MustCompile(`^/experiments/([^/]+)$`)
	ratioRe          = regexp.MustCompile(`^/experiments/([^/]+)/ratio$`)
	assignRe         = regexp.MustCompile(`^/experiments/([^/]+)/assign$`)
	eventsRe         = regexp.MustCompile(`^/experiments/([^/]+)/events$`)
	statsRe          = regexp.MustCompile(`^/experiments/([^/]+)/stats$`)
)

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	if experimentsRe.MatchString(path) {
		switch req.Method {
		case http.MethodGet:
			r.listExperiments(w, req)
		case http.MethodPost:
			r.createExperiment(w, req)
		default:
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if match := ratioRe.FindStringSubmatch(path); match != nil {
		if req.Method == http.MethodPut {
			r.updateRatio(w, req, match[1])
		} else {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if match := assignRe.FindStringSubmatch(path); match != nil {
		if req.Method == http.MethodGet {
			r.getAssignment(w, req, match[1])
		} else {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if match := eventsRe.FindStringSubmatch(path); match != nil {
		if req.Method == http.MethodPost {
			r.recordEvent(w, req, match[1])
		} else {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if match := statsRe.FindStringSubmatch(path); match != nil {
		if req.Method == http.MethodGet {
			r.getStats(w, req, match[1])
		} else {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if match := experimentNameRe.FindStringSubmatch(path); match != nil {
		if req.Method == http.MethodDelete {
			r.deleteExperiment(w, req, match[1])
		} else {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	jsonError(w, http.StatusNotFound, "not found")
}

func (r *Router) listExperiments(w http.ResponseWriter, req *http.Request) {
	experiments := r.store.ListActiveExperiments()
	response := make([]ExperimentListResponse, 0, len(experiments))
	for _, exp := range experiments {
		response = append(response, ExperimentListResponse{
			Name:        exp.Name,
			Description: exp.Description,
			Ratio:       exp.Ratio,
			CreatedAt:   exp.CreatedAt.Format(time.RFC3339),
		})
	}
	jsonResponse(w, http.StatusOK, response)
}

func (r *Router) createExperiment(w http.ResponseWriter, req *http.Request) {
	var body CreateExperimentRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(body.Name) == "" {
		jsonError(w, http.StatusBadRequest, "name is required")
		return
	}

	if err := r.store.CreateExperiment(body.Name, body.Description, body.Ratio); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (r *Router) updateRatio(w http.ResponseWriter, req *http.Request, name string) {
	var body UpdateRatioRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := r.store.UpdateRatio(name, body.Ratio); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (r *Router) deleteExperiment(w http.ResponseWriter, req *http.Request, name string) {
	if err := r.store.DeleteExperiment(name); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (r *Router) getAssignment(w http.ResponseWriter, req *http.Request, name string) {
	userID := req.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" {
		jsonError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	group, err := r.store.GetOrAssignUser(name, userID)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, AssignmentResponse{
		UserID: userID,
		Group:  group,
	})
}

func (r *Router) recordEvent(w http.ResponseWriter, req *http.Request, name string) {
	var body EventRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(body.UserID) == "" {
		jsonError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	if strings.TrimSpace(body.EventName) == "" {
		jsonError(w, http.StatusBadRequest, "event_name is required")
		return
	}

	_, err := r.store.GetOrAssignUser(name, body.UserID)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := r.store.RecordEvent(name, body.UserID, body.EventName); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "recorded"})
}

func (r *Router) getStats(w http.ResponseWriter, req *http.Request, name string) {
	stats, err := r.store.GetStats(name)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, stats)
}

func main() {
	store := NewExperimentStore()
	router := NewRouter(store)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8632"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	fmt.Printf("Server starting on port %s...\n", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
