package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"yaml-anchor-resolver/pkg/anchor"
	"yaml-anchor-resolver/pkg/api"
)

func main() {
	http.HandleFunc("/resolve", handleResolve)
	http.HandleFunc("/anchors", handleListAnchors)

	fmt.Println("Server starting on :8201")
	if err := http.ListenAndServe(":8201", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func handleResolve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	yamlText, err := extractYAML(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resolver := anchor.NewResolver()
	result, err := resolver.Resolve(yamlText)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := api.ResolveResponse{
		Success:   true,
		Data:      result.ResolvedData,
		Anchors:   result.Anchors,
		Refs:      result.References,
		Relations: result.Relations,
		Warnings:  result.Warnings,
	}

	writeJSON(w, http.StatusOK, response)
}

func handleListAnchors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	yamlText, err := extractYAML(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resolver := anchor.NewResolver()
	result, err := resolver.Resolve(yamlText)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := api.ListAnchorsResponse{
		Success:   true,
		Anchors:   result.Anchors,
		Refs:      result.References,
		Relations: result.Relations,
	}

	writeJSON(w, http.StatusOK, response)
}

func extractYAML(r *http.Request) (string, error) {
	contentType := r.Header.Get("Content-Type")

	if contentType == "multipart/form-data" {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			return "", fmt.Errorf("failed to parse multipart form: %v", err)
		}

		file, _, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			content, err := io.ReadAll(file)
			if err != nil {
				return "", fmt.Errorf("failed to read file: %v", err)
			}
			return string(content), nil
		}

		yamlText := r.FormValue("yaml_text")
		if yamlText != "" {
			return yamlText, nil
		}

		return "", fmt.Errorf("no file or yaml_text provided")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read request body: %v", err)
	}
	defer r.Body.Close()

	if len(body) == 0 {
		return "", fmt.Errorf("empty request body")
	}

	var req api.ResolveRequest
	if err := json.Unmarshal(body, &req); err == nil {
		if req.YAMLText != "" {
			return req.YAMLText, nil
		}
	}

	return string(body), nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
