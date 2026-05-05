package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"textstat/pkg/common"
	"textstat/pkg/textstat"
)

func getStopWords(custom map[string]bool) *textstat.StopWords {
	if custom != nil && len(custom) > 0 {
		return textstat.NewStopWordsWithCustom(custom)
	}
	return textstat.NewStopWords()
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(common.ErrorResponse{Error: "Method not allowed, use POST"})
		return
	}

	var req common.TextStatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ErrorResponse{Error: "Invalid request body"})
		return
	}

	stopWords := getStopWords(req.CustomStopWords)

	charCountWithSpaces := textstat.CountChars(req.Text, true)
	charCountWithoutSpaces := textstat.CountChars(req.Text, false)
	wordCount := textstat.CountWords(req.Text)
	lineCount := textstat.CountLines(req.Text)
	paragraphCount := textstat.CountParagraphs(req.Text)

	freqMap := textstat.CountWordFrequencies(req.Text, stopWords)
	topN := 10
	if req.TopN > 0 {
		topN = req.TopN
	}
	topWords := textstat.GetTopNWords(freqMap, topN)

	topWordItems := make([]common.TopWordItem, len(topWords))
	for i, wf := range topWords {
		topWordItems[i] = common.TopWordItem{
			Word:  wf.Word,
			Count: wf.Count,
		}
	}

	resp := common.TextStatsResponse{
		CharCountWithSpaces:    charCountWithSpaces,
		CharCountWithoutSpaces: charCountWithoutSpaces,
		WordCount:              wordCount,
		LineCount:              lineCount,
		ParagraphCount:         paragraphCount,
		TopWords:               topWordItems,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func similarityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(common.ErrorResponse{Error: "Method not allowed, use POST"})
		return
	}

	var req common.SimilarityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ErrorResponse{Error: "Invalid request body"})
		return
	}

	stopWords := getStopWords(req.CustomStopWords)
	similarity := textstat.TextSimilarity(req.Text1, req.Text2, stopWords)

	resp := common.SimilarityResponse{
		Similarity: similarity,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	http.HandleFunc("/api/stats", statsHandler)
	http.HandleFunc("/api/similarity", similarityHandler)
	http.HandleFunc("/health", healthHandler)

	port := ":8080"
	fmt.Printf("Text Statistics Server starting on %s...\n", port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  POST /api/stats      - Text statistics\n")
	fmt.Printf("  POST /api/similarity - Text similarity\n")
	fmt.Printf("  GET  /health         - Health check\n")
	fmt.Printf("\nPress Ctrl+C to stop\n")

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
