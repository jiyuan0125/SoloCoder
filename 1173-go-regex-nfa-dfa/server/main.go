package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regex-engine/api"
	"regex-engine/regex"
	"strings"
)

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) compileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSON(w, http.StatusMethodNotAllowed, api.CompileResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}

	var req api.CompileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.CompileResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}

	re, err := regex.Compile(req.Pattern)
	if err != nil {
		s.writeJSON(w, http.StatusOK, api.CompileResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, api.CompileResponse{
		Success:      true,
		Pattern:      re.Pattern,
		NFAStates:    re.NFAStateCount,
		DFAStates:    re.DFAStateCount,
		MinDFAStates: re.MinimizedDFACount,
	})
}

func (s *Server) matchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSON(w, http.StatusMethodNotAllowed, api.MatchResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}

	var req api.MatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.MatchResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}

	matched, err := regex.Match(req.Pattern, req.Text)
	if err != nil {
		s.writeJSON(w, http.StatusOK, api.MatchResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, api.MatchResponse{
		Success: true,
		Match:   matched,
	})
}

func convertNFA(nfa *regex.NFARepresentation) api.NFAJSON {
	transitions := make([]api.NFATransitionJSON, len(nfa.Transitions))
	for i, t := range nfa.Transitions {
		transitions[i] = api.NFATransitionJSON{
			From:   t.From,
			Symbol: t.Symbol,
			To:     append([]int{}, t.To...),
		}
	}
	return api.NFAJSON{
		StartState:  nfa.StartState,
		FinalStates: append([]int{}, nfa.FinalStates...),
		States:      nfa.States,
		Transitions: transitions,
	}
}

func convertDFA(dfa *regex.DFARepresentation) api.DFAJSON {
	transitions := make([]api.DFATransitionJSON, len(dfa.Transitions))
	for i, t := range dfa.Transitions {
		transitions[i] = api.DFATransitionJSON{
			From:   t.From,
			Symbol: t.Symbol,
			To:     t.To,
		}
	}
	return api.DFAJSON{
		StartState:  dfa.StartState,
		FinalStates: append([]int{}, dfa.FinalStates...),
		States:      dfa.States,
		Transitions: transitions,
	}
}

func convertMinDFA(dfa *regex.MinimizedDFARepresentation) api.DFAJSON {
	transitions := make([]api.DFATransitionJSON, len(dfa.Transitions))
	for i, t := range dfa.Transitions {
		transitions[i] = api.DFATransitionJSON{
			From:   t.From,
			Symbol: t.Symbol,
			To:     t.To,
		}
	}
	return api.DFAJSON{
		StartState:  dfa.StartState,
		FinalStates: append([]int{}, dfa.FinalStates...),
		States:      dfa.States,
		Transitions: transitions,
	}
}

func (s *Server) vizHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSON(w, http.StatusMethodNotAllowed, api.VizResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}

	var req api.VizRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.VizResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}

	re, err := regex.Compile(req.Pattern)
	if err != nil {
		s.writeJSON(w, http.StatusOK, api.VizResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	viz := re.Visualize()

	s.writeJSON(w, http.StatusOK, api.VizResponse{
		Success:      true,
		Pattern:      viz.Pattern,
		AST:          viz.AST,
		NFA:          convertNFA(viz.NFA),
		DFA:          convertDFA(viz.DFA),
		MinimizedDFA: convertMinDFA(viz.MinimizedDFA),
		NFAStates:    viz.NFAStates,
		DFAStates:    viz.DFAStates,
		MinDFAStates: viz.MinDFAStates,
	})
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func getPort() string {
	port := "8200"

	if envPort := os.Getenv("REGEX_SERVER_PORT"); envPort != "" {
		port = envPort
	}

	flag.StringVar(&port, "port", port, "server port")
	flag.Parse()

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	return port
}

func main() {
	port := getPort()
	server := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/compile", server.compileHandler)
	mux.HandleFunc("/match", server.matchHandler)
	mux.HandleFunc("/viz", server.vizHandler)
	mux.HandleFunc("/health", server.healthHandler)

	fmt.Printf("Regex Engine Server listening on %s\n", port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  POST /compile - compile regex pattern\n")
	fmt.Printf("  POST /match   - test regex matching\n")
	fmt.Printf("  POST /viz     - get full visualization\n")
	fmt.Printf("  GET  /health  - health check\n")

	if err := http.ListenAndServe(port, mux); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
