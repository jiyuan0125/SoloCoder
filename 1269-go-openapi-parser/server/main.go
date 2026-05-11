package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/example/openapi-parser/api"
	"github.com/example/openapi-parser/openapi"
)

func parseSpecFromRequest(r *http.Request) (*openapi.OpenAPI, error) {
	var req api.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	filename := req.Filename
	if filename == "" {
		filename = "openapi.yaml"
	}

	return openapi.Parse([]byte(req.Content), filename)
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	spec, err := parseSpecFromRequest(r)
	if err != nil {
		json.NewEncoder(w).Encode(api.ParseResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(api.ParseResponse{
		Success: true,
		Data:    spec.Raw,
	})
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	spec, err := parseSpecFromRequest(r)
	if err != nil {
		json.NewEncoder(w).Encode(api.ValidateResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	result := openapi.Validate(spec)

	errors := make([]api.ValidateError, len(result.Errors))
	for i, e := range result.Errors {
		errors[i] = api.ValidateError{
			Field:   e.Field,
			Message: e.Message,
		}
	}

	json.NewEncoder(w).Encode(api.ValidateResponse{
		Success: true,
		Valid:   result.Valid,
		Errors:  errors,
	})
}

func handleListPaths(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	spec, err := parseSpecFromRequest(r)
	if err != nil {
		json.NewEncoder(w).Encode(api.ListPathsResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	paths := openapi.GetAllPaths(spec)

	result := make([]api.PathInfo, len(paths))
	for i, p := range paths {
		result[i] = api.PathInfo{
			Path:    p.Path,
			Methods: p.Methods,
		}
	}

	json.NewEncoder(w).Encode(api.ListPathsResponse{
		Success: true,
		Paths:   result,
	})
}

func handleResolve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req api.ResolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(api.ResolveResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
		return
	}

	filename := req.Filename
	if filename == "" {
		filename = "openapi.yaml"
	}

	spec, err := openapi.Parse([]byte(req.Content), filename)
	if err != nil {
		json.NewEncoder(w).Encode(api.ResolveResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var resolved interface{}
	if req.All {
		resolved = openapi.ResolveAll(spec)
	} else {
		resolved, err = openapi.ResolveReference(spec, req.Ref)
		if err != nil {
			json.NewEncoder(w).Encode(api.ResolveResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
	}

	json.NewEncoder(w).Encode(api.ResolveResponse{
		Success: true,
		Data:    resolved,
	})
}

func getPort() string {
	port := "8080"

	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	flagPort := flag.String("port", "", "HTTP server port (default: 8080 or $PORT)")
	flag.Parse()

	if *flagPort != "" {
		port = *flagPort
	}

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	return port
}

func main() {
	port := getPort()

	http.HandleFunc("/openapi/parse", handleParse)
	http.HandleFunc("/openapi/validate", handleValidate)
	http.HandleFunc("/openapi/list-paths", handleListPaths)
	http.HandleFunc("/openapi/resolve", handleResolve)

	fmt.Printf("OpenAPI Parser Server starting on port %s\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /openapi/parse     - Parse OpenAPI file")
	fmt.Println("  POST /openapi/validate  - Validate OpenAPI file")
	fmt.Println("  POST /openapi/list-paths - List all paths")
	fmt.Println("  POST /openapi/resolve   - Resolve $ref references")

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}
