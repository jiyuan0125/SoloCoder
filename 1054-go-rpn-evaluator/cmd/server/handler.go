package main

import (
	"encoding/json"
	"net/http"

	"rpn-evaluator/pkg/api"
	"rpn-evaluator/pkg/rpn"
)

type Server struct {
	vars *rpn.VariableStore
}

func NewServer() *Server {
	return &Server{
		vars: rpn.NewVariableStore(),
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.ConvertResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ConvertResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	terms, err := rpn.InfixToRPN(req.Expression)
	if err != nil {
		writeJSON(w, http.StatusOK, api.ConvertResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.ConvertResponse{
		Success: true,
		RPN:     rpn.RPNToString(terms),
	})
}

func (s *Server) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.EvaluateResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.EvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.EvaluateResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	terms, err := rpn.InfixToRPN(req.Expression)
	if err != nil {
		writeJSON(w, http.StatusOK, api.EvaluateResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	result, err := rpn.EvaluateRPN(terms, s.vars)
	if err != nil {
		writeJSON(w, http.StatusOK, api.EvaluateResponse{
			Success: false,
			RPN:     rpn.RPNToString(terms),
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.EvaluateResponse{
		Success:   true,
		RPN:       rpn.RPNToString(terms),
		Result:    result,
		ResultStr: rpn.FormatResult(result),
	})
}

func (s *Server) handleSetVariable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.SetVariableResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.SetVariableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.SetVariableResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	if req.Name == "" {
		writeJSON(w, http.StatusOK, api.SetVariableResponse{
			Success: false,
			Error:   "variable name cannot be empty",
		})
		return
	}

	s.vars.Set(req.Name, req.Value)
	writeJSON(w, http.StatusOK, api.SetVariableResponse{
		Success: true,
	})
}

func (s *Server) handleGetVariable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.GetVariableResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.GetVariableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.GetVariableResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	value, ok := s.vars.Get(req.Name)
	if !ok {
		writeJSON(w, http.StatusOK, api.GetVariableResponse{
			Success: false,
			Error:   "variable not found: " + req.Name,
		})
		return
	}

	writeJSON(w, http.StatusOK, api.GetVariableResponse{
		Success: true,
		Value:   value,
	})
}

func (s *Server) handleListVariables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, api.ListVariablesResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	variables := s.vars.GetAll()
	writeJSON(w, http.StatusOK, api.ListVariablesResponse{
		Success:   true,
		Variables: variables,
	})
}
