package main

import (
	"encoding/json"
	"net/http"

	"multirate-limiter/common"
	"multirate-limiter/limiter"
)

type Server struct {
	manager *limiter.Manager
}

func NewServer() *Server {
	return &Server{
		manager: limiter.NewManager(),
	}
}

func (s *Server) HandleAllow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.WriteJSON(w, http.StatusMethodNotAllowed, common.ErrorResponse{Error: "method not allowed"})
		return
	}

	var req common.AllowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Path == "" {
		common.WriteJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "path is required"})
		return
	}

	allowed, results, found := s.manager.Allow(req.Path, req.ClientID, req.HasClientID)
	if !found {
		common.WriteJSON(w, http.StatusNotFound, common.ErrorResponse{Error: "path not configured"})
		return
	}

	resp := common.AllowResponse{
		Allowed:  allowed,
		AllRules: results,
	}
	if !allowed {
		rejected := make([]common.RuleResult, 0)
		for _, r := range results {
			if !r.Allowed {
				rejected = append(rejected, r)
			}
		}
		resp.RejectedRules = rejected
	}

	common.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) HandleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleConfigGet(w, r)
	case http.MethodPost, http.MethodPut:
		s.handleConfigSet(w, r)
	default:
		common.WriteJSON(w, http.StatusMethodNotAllowed, common.ErrorResponse{Error: "method not allowed"})
	}
}

func (s *Server) handleConfigGet(w http.ResponseWriter, r *http.Request) {
	var req common.ConfigGetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Path == "" {
		common.WriteJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "path is required"})
		return
	}

	rules, found := s.manager.GetRules(req.Path)
	if !found {
		common.WriteJSON(w, http.StatusNotFound, common.ErrorResponse{Error: "path not configured"})
		return
	}

	common.WriteJSON(w, http.StatusOK, common.ConfigGetResponse{
		Path:  req.Path,
		Rules: rules,
	})
}

func (s *Server) handleConfigSet(w http.ResponseWriter, r *http.Request) {
	var req common.ConfigSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Path == "" {
		common.WriteJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "path is required"})
		return
	}

	if len(req.Rules) == 0 {
		common.WriteJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "rules are required"})
		return
	}

	s.manager.SetRules(req.Path, req.Rules)
	common.WriteJSON(w, http.StatusOK, common.ConfigSetResponse{
		Success: true,
		Message: "configuration updated",
	})
}

func (s *Server) HandleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.WriteJSON(w, http.StatusMethodNotAllowed, common.ErrorResponse{Error: "method not allowed"})
		return
	}

	var req common.StatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Path == "" {
		common.WriteJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "path is required"})
		return
	}

	stats, found := s.manager.Stats(req.Path, req.ClientID, req.HasClientID)
	if !found {
		common.WriteJSON(w, http.StatusNotFound, common.ErrorResponse{Error: "path not configured"})
		return
	}

	resp := common.StatsResponse{
		Path:  req.Path,
		Stats: stats,
	}
	if req.HasClientID {
		resp.ClientID = req.ClientID
	}
	common.WriteJSON(w, http.StatusOK, resp)
}
