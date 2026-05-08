package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sensitive-word-filter/pkg/common"
	"sensitive-word-filter/pkg/filter"
)

var f *filter.Filter

func main() {
	f = filter.NewFilter()

	http.HandleFunc("/api/filter", handleFilter)
	http.HandleFunc("/api/sensitive-words", handleSensitiveWords)
	http.HandleFunc("/api/sensitive-words/add", handleAddSensitiveWord)
	http.HandleFunc("/api/sensitive-words/remove", handleRemoveSensitiveWord)
	http.HandleFunc("/api/sensitive-words/update", handleUpdateSensitiveWord)
	http.HandleFunc("/api/whitelist", handleWhitelist)
	http.HandleFunc("/api/whitelist/add", handleAddWhitelist)
	http.HandleFunc("/api/whitelist/remove", handleRemoveWhitelist)
	http.HandleFunc("/api/stats", handleStats)

	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func handleFilter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.FilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := f.Filter(req.Text)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleSensitiveWords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	words := f.ListSensitiveWords()
	response := common.ListSensitiveWordsResponse{
		Words: words,
		Total: len(words),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleAddSensitiveWord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.AddSensitiveWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := f.AddSensitiveWord(req.Word); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(common.GenericResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.GenericResponse{
		Success: true,
		Message: "Word added successfully",
	})
}

func handleRemoveSensitiveWord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.DeleteSensitiveWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := f.RemoveSensitiveWord(req.Word); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(common.GenericResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.GenericResponse{
		Success: true,
		Message: "Word removed successfully",
	})
}

func handleUpdateSensitiveWord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.UpdateSensitiveWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := f.UpdateSensitiveWord(req.OldWord, req.NewWord); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(common.GenericResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.GenericResponse{
		Success: true,
		Message: "Word updated successfully",
	})
}

func handleWhitelist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	words := f.ListWhitelistWords()
	response := common.ListWhitelistResponse{
		Words: words,
		Total: len(words),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleAddWhitelist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.AddWhitelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := f.AddWhitelistWord(req.Word); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(common.GenericResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.GenericResponse{
		Success: true,
		Message: "Word added to whitelist successfully",
	})
}

func handleRemoveWhitelist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.DeleteWhitelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := f.RemoveWhitelistWord(req.Word); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(common.GenericResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.GenericResponse{
		Success: true,
		Message: "Word removed from whitelist successfully",
	})
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := f.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
