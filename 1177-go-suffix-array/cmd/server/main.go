package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"suffixarray/pkg/api"
	"suffixarray/pkg/suffixarray"
)

const defaultPort = "8401"

func main() {
	var portFlag string
	flag.StringVar(&portFlag, "port", "", "服务端监听端口")
	flag.Parse()

	port := getPort(portFlag)

	http.HandleFunc("/build", handleBuild)
	http.HandleFunc("/lrs", handleLRS)
	http.HandleFunc("/count", handleCount)
	http.HandleFunc("/demo", handleDemo)

	log.Printf("后缀数组服务已启动，监听端口: %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func getPort(portFlag string) string {
	if portFlag != "" {
		return portFlag
	}
	if envPort := os.Getenv("SUFFIXARRAY_PORT"); envPort != "" {
		return envPort
	}
	return defaultPort
}

func handleBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "只支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	var req api.BuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, fmt.Sprintf("请求解析失败: %v", err), http.StatusBadRequest)
		return
	}

	sa := suffixarray.Build(req.Text)

	resp := api.BuildResponse{
		Success: true,
		Text:    req.Text,
		SA:      sa.SA(),
		Rank:    sa.Rank(),
		LCP:     sa.LCP(),
	}

	sendJSON(w, resp)
}

func handleLRS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "只支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	var req api.LRSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, fmt.Sprintf("请求解析失败: %v", err), http.StatusBadRequest)
		return
	}

	sa := suffixarray.Build(req.Text)
	substrings := sa.LongestRepeatedSubstrings()

	maxLen := 0
	for _, s := range substrings {
		if len(s.Substring) > maxLen {
			maxLen = len(s.Substring)
		}
	}

	apiSubstrings := make([]api.RepeatedSubstring, len(substrings))
	for i, s := range substrings {
		apiSubstrings[i] = api.RepeatedSubstring{
			Substring: s.Substring,
			Positions: s.Positions,
		}
	}

	resp := api.LRSResponse{
		Success:    true,
		Text:       req.Text,
		MaxLength:  maxLen,
		Substrings: apiSubstrings,
	}

	sendJSON(w, resp)
}

func handleCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "只支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	var req api.CountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, fmt.Sprintf("请求解析失败: %v", err), http.StatusBadRequest)
		return
	}

	sa := suffixarray.Build(req.Text)
	occurrences := sa.CountOccurrences(req.Pattern)

	resp := api.CountResponse{
		Success:   true,
		Text:      req.Text,
		Pattern:   req.Pattern,
		Count:     occurrences.Count,
		Positions: occurrences.Positions,
	}

	sendJSON(w, resp)
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "只支持GET请求", http.StatusMethodNotAllowed)
		return
	}

	text := "To be or not to be, that is the question. Whether 'tis nobler in the mind to suffer the slings and arrows of outrageous fortune."

	sa := suffixarray.Build(text)
	lrs := sa.LongestRepeatedSubstrings()

	patterns := []string{"to", "be", "the", "question"}
	queries := make([]api.QueryResult, len(patterns))
	for i, p := range patterns {
		occ := sa.CountOccurrences(p)
		queries[i] = api.QueryResult{
			Pattern:   p,
			Count:     occ.Count,
			Positions: occ.Positions,
		}
	}

	apiSubstrings := make([]api.RepeatedSubstring, len(lrs))
	for i, s := range lrs {
		apiSubstrings[i] = api.RepeatedSubstring{
			Substring: s.Substring,
			Positions: s.Positions,
		}
	}

	resp := api.DemoResponse{
		Success:         true,
		Text:            text,
		SA:              sa.SA(),
		LCP:             sa.LCP(),
		LongestRepeated: apiSubstrings,
		ExampleQueries:  queries,
	}

	sendJSON(w, resp)
}

func sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
