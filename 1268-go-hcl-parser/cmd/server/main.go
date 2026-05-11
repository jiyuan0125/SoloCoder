package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"hcl-parser/pkg/api"
	"hcl-parser/pkg/hclparser"
)

func getPort() string {
	if port := os.Getenv("HCL_SERVER_PORT"); port != "" {
		return port
	}
	port := flag.String("port", "8080", "Server port")
	flag.Parse()
	return *port
}

func main() {
	port := getPort()

	http.HandleFunc("/hcl/parse", parseHandler)
	http.HandleFunc("/hcl/validate", validateHandler)
	http.HandleFunc("/hcl/format", formatHandler)
	http.HandleFunc("/hcl/get", getHandler)

	fmt.Printf("HCL Parser Server listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func parseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.ParseResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.ParseResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to read request body: %v", err),
		})
		return
	}

	var req api.ParseRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ParseResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid JSON: %v", err),
		})
		return
	}

	parsed, err := hclparser.ParseString(req.Content)
	if err != nil {
		writeJSON(w, http.StatusOK, api.ParseResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.ParseResponse{
		Success: true,
		Body:    parsed,
	})
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.ValidateResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.ValidateResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to read request body: %v", err),
		})
		return
	}

	var req api.ValidateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ValidateResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid JSON: %v", err),
		})
		return
	}

	parsed, err := hclparser.ParseString(req.Content)
	if err != nil {
		writeJSON(w, http.StatusOK, api.ValidateResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	schema := schemaFromInterface(req.Schema)
	validationErrors, _ := hclparser.Validate(parsed, schema)

	apiErrors := make([]api.ValidationError, len(validationErrors))
	for i, e := range validationErrors {
		apiErrors[i] = api.ValidationError{
			Path:    e.Path,
			Message: e.Message,
		}
	}

	writeJSON(w, http.StatusOK, api.ValidateResponse{
		Success: true,
		Valid:   len(validationErrors) == 0,
		Errors:  apiErrors,
	})
}

func schemaFromInterface(s interface{}) *hclparser.Schema {
	schema := hclparser.NewSchema()

	schemaMap, ok := s.(map[string]interface{})
	if !ok {
		return schema
	}

	if blockTypes, ok := schemaMap["block_types"].(map[string]interface{}); ok {
		for name, bs := range blockTypes {
			schema.BlockTypes[name] = blockSchemaFromInterface(bs)
		}
	}

	return schema
}

func blockSchemaFromInterface(s interface{}) *hclparser.BlockSchema {
	bs := &hclparser.BlockSchema{
		Attributes:   make(map[string]*hclparser.AttributeSchema),
		NestedBlocks: make(map[string]*hclparser.BlockSchema),
	}

	m, ok := s.(map[string]interface{})
	if !ok {
		return bs
	}

	if labels, ok := m["labels"].(float64); ok {
		bs.Labels = int(labels)
	}
	if minLabels, ok := m["min_labels"].(float64); ok {
		bs.MinLabels = int(minLabels)
	}
	if maxLabels, ok := m["max_labels"].(float64); ok {
		bs.MaxLabels = int(maxLabels)
	}

	if attrs, ok := m["attributes"].(map[string]interface{}); ok {
		for name, as := range attrs {
			bs.Attributes[name] = attrSchemaFromInterface(as)
		}
	}

	if nested, ok := m["nested_blocks"].(map[string]interface{}); ok {
		for name, ns := range nested {
			bs.NestedBlocks[name] = blockSchemaFromInterface(ns)
		}
	}

	return bs
}

func attrSchemaFromInterface(s interface{}) *hclparser.AttributeSchema {
	as := &hclparser.AttributeSchema{}

	m, ok := s.(map[string]interface{})
	if !ok {
		return as
	}

	if required, ok := m["required"].(bool); ok {
		as.Required = required
	}
	if typ, ok := m["type"].(string); ok {
		as.Type = typ
	}

	return as
}

func formatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.FormatResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.FormatResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to read request body: %v", err),
		})
		return
	}

	var req api.FormatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.FormatResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid JSON: %v", err),
		})
		return
	}

	parsed, err := hclparser.ParseString(req.Content)
	if err != nil {
		writeJSON(w, http.StatusOK, api.FormatResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	formatted := hclparser.Format(parsed)
	writeJSON(w, http.StatusOK, api.FormatResponse{
		Success: true,
		Content: formatted,
	})
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.GetResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.GetResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to read request body: %v", err),
		})
		return
	}

	var req api.GetRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.GetResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid JSON: %v", err),
		})
		return
	}

	parsed, err := hclparser.ParseString(req.Content)
	if err != nil {
		writeJSON(w, http.StatusOK, api.GetResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	value, err := hclparser.GetValueByPath(parsed, req.Path)
	if err != nil {
		writeJSON(w, http.StatusOK, api.GetResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.GetResponse{
		Success: true,
		Value:   value,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
