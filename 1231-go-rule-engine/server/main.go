package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"ruleengine/api"
	"ruleengine/engine"
)

type server struct {
	eng *engine.Engine
}

func newServer() *server {
	return &server{eng: engine.NewEngine(engine.NewStore())}
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func convertEnv(env map[string]interface{}) map[string]string {
	result := make(map[string]string, len(env))
	for k, v := range env {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, api.ErrorResponse{Error: msg})
}

func toAPIRule(r *engine.Rule) *api.Rule {
	return &api.Rule{
		ID:          r.ID,
		Priority:    r.Priority,
		Condition:   r.Condition,
		Actions:     append([]string(nil), r.Actions...),
		Description: r.Description,
	}
}

func fromAPIRule(r *api.Rule) *engine.Rule {
	return &engine.Rule{
		ID:          r.ID,
		Priority:    r.Priority,
		Condition:   r.Condition,
		Actions:     append([]string(nil), r.Actions...),
		Description: r.Description,
	}
}

func (s *server) handleRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.addRule(w, r)
	case http.MethodGet:
		s.listRules(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *server) addRule(w http.ResponseWriter, r *http.Request) {
	var req api.AddRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Rule == nil {
		writeError(w, http.StatusBadRequest, "rule is required")
		return
	}

	erule := fromAPIRule(req.Rule)
	if err := engine.ValidateRule(erule); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, exists := s.eng.GetRule(erule.ID); exists {
		writeError(w, http.StatusConflict, "rule id already exists")
		return
	}

	s.eng.AddRule(erule)
	writeJSON(w, http.StatusCreated, api.AddRuleResponse{ID: erule.ID})
}

func (s *server) listRules(w http.ResponseWriter, _ *http.Request) {
	rules := s.eng.ListRules()
	apiRules := make([]*api.Rule, len(rules))
	for i, r := range rules {
		apiRules[i] = toAPIRule(r)
	}
	writeJSON(w, http.StatusOK, api.ListRulesResponse{Rules: apiRules})
}

func (s *server) handleRule(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/rules/")
	if id == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getRule(w, r, id)
	case http.MethodPut:
		s.updateRule(w, r, id)
	case http.MethodDelete:
		s.deleteRule(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *server) getRule(w http.ResponseWriter, _ *http.Request, id string) {
	r, ok := s.eng.GetRule(id)
	if !ok {
		writeError(w, http.StatusNotFound, "rule not found")
		return
	}
	writeJSON(w, http.StatusOK, api.GetRuleResponse{Rule: toAPIRule(r)})
}

func (s *server) updateRule(w http.ResponseWriter, r *http.Request, id string) {
	var req api.UpdateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Rule == nil {
		writeError(w, http.StatusBadRequest, "rule is required")
		return
	}
	if req.Rule.ID != id {
		writeError(w, http.StatusBadRequest, "rule id mismatch")
		return
	}

	erule := fromAPIRule(req.Rule)
	if err := engine.ValidateRule(erule); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, exists := s.eng.GetRule(erule.ID); !exists {
		writeError(w, http.StatusNotFound, "rule not found")
		return
	}

	s.eng.UpdateRule(erule)
	writeJSON(w, http.StatusOK, api.UpdateRuleResponse{})
}

func (s *server) deleteRule(w http.ResponseWriter, _ *http.Request, id string) {
	ok := s.eng.DeleteRule(id)
	writeJSON(w, http.StatusOK, api.DeleteRuleResponse{OK: ok})
}

func (s *server) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.EvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	env := convertEnv(req.Environment)
	result, err := s.eng.Evaluate(env)
	if err != nil {
		writeJSON(w, http.StatusOK, api.EvaluateResponse{Error: err.Error()})
		return
	}

	matched := make([]api.Rule, len(result.MatchedRules))
	for i, m := range result.MatchedRules {
		matched[i] = api.Rule{
			ID:          m.ID,
			Priority:    m.Priority,
			Condition:   m.Condition,
			Actions:     append([]string(nil), m.Actions...),
			Description: m.Description,
		}
	}

	results := make([]api.ActionResult, len(result.Results))
	for i, r := range result.Results {
		results[i] = api.ActionResult{
			RuleID:  r.RuleID,
			RulePri: r.RulePri,
			Action:  r.Action,
			Status:  r.Status,
			Error:   r.Error,
		}
	}

	writeJSON(w, http.StatusOK, api.EvaluateResponse{
		MatchedRules: matched,
		Results:      results,
	})
}

func getPort() string {
	var port int
	flag.IntVar(&port, "port", 0, "server port")
	flag.Parse()

	if port > 0 {
		return strconv.Itoa(port)
	}

	if envPort := os.Getenv("RULE_ENGINE_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			return strconv.Itoa(p)
		}
	}

	return "8212"
}

func main() {
	srv := newServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/rules", srv.handleRules)
	mux.HandleFunc("/rules/", srv.handleRule)
	mux.HandleFunc("/evaluate", srv.handleEvaluate)

	port := getPort()
	addr := fmt.Sprintf(":%s", port)

	log.Printf("rule engine server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
