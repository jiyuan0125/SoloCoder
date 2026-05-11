package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"thompson-nfa/pkg/api"
	"thompson-nfa/pkg/regex"
)

func main() {
	http.HandleFunc("/api/regex/compile", handleCompile)
	http.HandleFunc("/api/regex/match", handleMatch)
	http.HandleFunc("/api/regex/findall", handleFindAll)
	http.HandleFunc("/api/regex/visualize", handleVisualize)
	http.HandleFunc("/api/regex/test", handleTest)

	fmt.Println("Server starting on :8202")
	if err := http.ListenAndServe(":8202", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func handleCompile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.CompileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.CompileResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	re, err := regex.NewRegex(req.Pattern)
	if err != nil {
		writeJSON(w, http.StatusOK, api.CompileResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.CompileResponse{
		Success:         true,
		StateCount:      re.StateCount(),
		TransitionCount: re.TransitionCount(),
	})
}

func handleMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.MatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.MatchResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	re, err := regex.NewRegex(req.Pattern)
	if err != nil {
		writeJSON(w, http.StatusOK, api.MatchResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.MatchResponse{
		Success: true,
		Match:   re.MatchString(req.Input),
	})
}

func handleFindAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.FindAllRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.FindAllResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	re, err := regex.NewRegex(req.Pattern)
	if err != nil {
		writeJSON(w, http.StatusOK, api.FindAllResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	matches := re.FindAllString(req.Input)
	indices := re.FindAllStringIndex(req.Input)

	var matchInfos []api.MatchInfo
	for i, match := range matches {
		matchInfos = append(matchInfos, api.MatchInfo{
			Match: match,
			Start: indices[i][0],
			End:   indices[i][1],
		})
	}

	writeJSON(w, http.StatusOK, api.FindAllResponse{
		Success: true,
		Matches: matchInfos,
	})
}

func handleVisualize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.VisualizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.VisualizeResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	re, err := regex.NewRegex(req.Pattern)
	if err != nil {
		writeJSON(w, http.StatusOK, api.VisualizeResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.VisualizeResponse{
		Success: true,
		Graph:   re.Visualize(),
	})
}

func handleTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.TestResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	re, err := regex.NewRegex(req.Pattern)
	if err != nil {
		writeJSON(w, http.StatusOK, api.TestResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var results []api.TestResult
	for _, input := range req.Inputs {
		results = append(results, api.TestResult{
			Input: input,
			Match: re.MatchString(input),
		})
	}

	writeJSON(w, http.StatusOK, api.TestResponse{
		Success: true,
		Results: results,
	})
}
