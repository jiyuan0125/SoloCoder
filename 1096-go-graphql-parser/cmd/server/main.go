package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"

	"graphql-parser/internal/common"
	"graphql-parser/internal/graphql/executor"
	"graphql-parser/internal/graphql/registry"
)

type Server struct {
	registry      *registry.Registry
	execContext   *executor.ExecutionContext
	presetVars    map[string]interface{}
	varsMu        sync.RWMutex
}

func NewServer() *Server {
	reg := registry.New()
	setupDemoSchema(reg)

	if err := reg.CheckCyclicReferences(); err != nil {
		fmt.Fprintf(os.Stderr, "Schema error: %v\n", err)
		os.Exit(1)
	}

	ec := executor.NewExecutionContext(reg)

	return &Server{
		registry:    reg,
		execContext: ec,
		presetVars:  make(map[string]interface{}),
	}
}

func setupDemoSchema(reg *registry.Registry) {
	queryType := reg.RegisterType("Query")
	userType := reg.RegisterType("User")
	postType := reg.RegisterType("Post")
	commentType := reg.RegisterType("Comment")

	usersData := map[string]interface{}{
		"1": map[string]interface{}{
			"id":    "1",
			"name":  "Alice",
			"email": "alice@example.com",
			"age":   30,
		},
		"2": map[string]interface{}{
			"id":    "2",
			"name":  "Bob",
			"email": "bob@example.com",
			"age":   25,
		},
	}

	postsData := map[string]interface{}{
		"p1": map[string]interface{}{
			"id":     "p1",
			"title":  "First Post",
			"body":   "Hello World",
			"userId": "1",
		},
		"p2": map[string]interface{}{
			"id":     "p2",
			"title":  "Second Post",
			"body":   "Goodbye World",
			"userId": "2",
		},
		"p3": map[string]interface{}{
			"id":     "p3",
			"title":  "Third Post",
			"body":   "Third time's the charm",
			"userId": "1",
		},
	}

	commentsData := []interface{}{
		map[string]interface{}{"id": "c1", "text": "Great post!", "postId": "p1"},
		map[string]interface{}{"id": "c2", "text": "Nice!", "postId": "p1"},
		map[string]interface{}{"id": "c3", "text": "Interesting", "postId": "p2"},
	}

	queryType.RegisterField("user", "User", map[string]string{
		"id": "ID!",
	}, func(ctx context.Context, source interface{}, args map[string]interface{}) (interface{}, error) {
		id := fmt.Sprintf("%v", args["id"])
		if user, ok := usersData[id]; ok {
			return user, nil
		}
		return nil, nil
	})

	queryType.RegisterField("users", "[User]", nil, func(ctx context.Context, source interface{}, args map[string]interface{}) (interface{}, error) {
		result := []interface{}{}
		for _, u := range usersData {
			result = append(result, u)
		}
		return result, nil
	})

	queryType.RegisterField("post", "Post", map[string]string{
		"id": "ID!",
	}, func(ctx context.Context, source interface{}, args map[string]interface{}) (interface{}, error) {
		id := fmt.Sprintf("%v", args["id"])
		if post, ok := postsData[id]; ok {
			return post, nil
		}
		return nil, nil
	})

	queryType.RegisterField("posts", "[Post]", nil, func(ctx context.Context, source interface{}, args map[string]interface{}) (interface{}, error) {
		result := []interface{}{}
		for _, p := range postsData {
			result = append(result, p)
		}
		return result, nil
	})

	userType.RegisterField("id", "ID", nil, nil)
	userType.RegisterField("name", "String", nil, nil)
	userType.RegisterField("email", "String", nil, nil)
	userType.RegisterField("age", "Int", nil, nil)
	userType.RegisterField("posts", "[Post]", nil, func(ctx context.Context, source interface{}, args map[string]interface{}) (interface{}, error) {
		user := source.(map[string]interface{})
		userId := fmt.Sprintf("%v", user["id"])
		result := []interface{}{}
		for _, p := range postsData {
			post := p.(map[string]interface{})
			if fmt.Sprintf("%v", post["userId"]) == userId {
				result = append(result, p)
			}
		}
		return result, nil
	})

	postType.RegisterField("id", "ID", nil, nil)
	postType.RegisterField("title", "String", nil, nil)
	postType.RegisterField("body", "String", nil, nil)
	postType.RegisterField("comments", "[Comment]", nil, func(ctx context.Context, source interface{}, args map[string]interface{}) (interface{}, error) {
		post := source.(map[string]interface{})
		postId := fmt.Sprintf("%v", post["id"])
		result := []interface{}{}
		for _, c := range commentsData {
			comment := c.(map[string]interface{})
			if fmt.Sprintf("%v", comment["postId"]) == postId {
				result = append(result, c)
			}
		}
		return result, nil
	})
	postType.RegisterField("author", "User", nil, func(ctx context.Context, source interface{}, args map[string]interface{}) (interface{}, error) {
		post := source.(map[string]interface{})
		userId := fmt.Sprintf("%v", post["userId"])
		if user, ok := usersData[userId]; ok {
			return user, nil
		}
		return nil, nil
	})

	commentType.RegisterField("id", "ID", nil, nil)
	commentType.RegisterField("text", "String", nil, nil)
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		query := r.URL.Query().Get("query")
		if query == "" {
			s.writeJSON(w, http.StatusBadRequest, common.ValidateResponse{
				Valid:  false,
				Errors: []string{"query parameter required"},
			})
			return
		}
		_, err := executor.Validate(query)
		if err != nil {
			s.writeJSON(w, http.StatusOK, common.ValidateResponse{
				Valid:  false,
				Errors: []string{err.Error()},
			})
			return
		}
		formatted, formatErr := executor.Format(query)
		if formatErr != nil {
			formatted = query
		}
		s.writeJSON(w, http.StatusOK, common.ValidateResponse{
			Valid:    true,
			Warnings: []string{fmt.Sprintf("Formatted:\n%s", formatted)},
		})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.writeJSON(w, http.StatusBadRequest, common.GraphQLResponse{
			Errors: []string{err.Error()},
		})
		return
	}
	defer r.Body.Close()

	var req common.GraphQLRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, common.GraphQLResponse{
			Errors: []string{err.Error()},
		})
		return
	}

	if req.Query == "" {
		s.writeJSON(w, http.StatusBadRequest, common.GraphQLResponse{
			Errors: []string{"query is required"},
		})
		return
	}

	ctx := r.Context()

	s.varsMu.RLock()
	presetVars := make(map[string]interface{})
	for k, v := range s.presetVars {
		presetVars[k] = v
	}
	s.varsMu.RUnlock()

	variables := make(map[string]interface{})
	for k, v := range presetVars {
		variables[k] = v
	}
	for k, v := range req.Variables {
		variables[k] = v
	}

	result, execErr := executor.Execute(ctx, s.execContext, req.Query, variables)
	if execErr != nil {
		if result != nil {
			s.writeJSON(w, http.StatusOK, result)
			return
		}
		s.writeJSON(w, http.StatusInternalServerError, common.GraphQLResponse{
			Errors: []string{execErr.Error()},
		})
		return
	}

	s.writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleSchema(w http.ResponseWriter, r *http.Request) {
	types := s.registry.Types()
	response := common.SchemaResponse{}

	for _, td := range types {
		ti := common.TypeInfo{Name: td.Name}
		for _, fd := range td.Fields {
			args := make(map[string]string)
			for k, v := range fd.Args {
				args[k] = v
			}
			ti.Fields = append(ti.Fields, common.FieldInfo{
				Name: fd.Name,
				Type: fd.Type,
				Args: args,
			})
		}
		response.Types = append(response.Types, ti)
	}

	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleVariables(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			s.writeJSON(w, http.StatusBadRequest, common.VariablesResponse{})
			return
		}
		defer r.Body.Close()

		var req common.VariablesSetRequest
		if err := json.Unmarshal(body, &req); err != nil {
			s.writeJSON(w, http.StatusBadRequest, common.VariablesResponse{})
			return
		}

		s.varsMu.Lock()
		for k, v := range req.Variables {
			s.presetVars[k] = v
		}
		presetCopy := make(map[string]interface{})
		for k, v := range s.presetVars {
			presetCopy[k] = v
		}
		s.varsMu.Unlock()

		s.writeJSON(w, http.StatusOK, common.VariablesResponse{Variables: presetCopy})
		return
	}

	if r.Method == http.MethodGet {
		s.varsMu.RLock()
		presetCopy := make(map[string]interface{})
		for k, v := range s.presetVars {
			presetCopy[k] = v
		}
		s.varsMu.RUnlock()
		s.writeJSON(w, http.StatusOK, common.VariablesResponse{Variables: presetCopy})
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			s.writeJSON(w, http.StatusBadRequest, common.ValidateResponse{
				Valid:  false,
				Errors: []string{err.Error()},
			})
			return
		}
		defer r.Body.Close()

		var req common.GraphQLRequest
		if err := json.Unmarshal(body, &req); err != nil {
			s.writeJSON(w, http.StatusBadRequest, common.ValidateResponse{
				Valid:  false,
				Errors: []string{err.Error()},
			})
			return
		}

		_, parseErr := executor.Validate(req.Query)
		formatted, formatErr := executor.Format(req.Query)
		if parseErr != nil {
			s.writeJSON(w, http.StatusOK, common.ValidateResponse{
				Valid:  false,
				Errors: []string{parseErr.Error()},
			})
			return
		}

		warnings := []string{}
		if formatErr == nil {
			warnings = append(warnings, "Formatted query:\n"+formatted)
		}

		s.writeJSON(w, http.StatusOK, common.ValidateResponse{
			Valid:    true,
			Warnings: warnings,
		})
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleFormat(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			s.writeJSON(w, http.StatusBadRequest, common.FormatResponse{
				Errors: []string{err.Error()},
			})
			return
		}
		defer r.Body.Close()

		var req common.GraphQLRequest
		if err := json.Unmarshal(body, &req); err != nil {
			s.writeJSON(w, http.StatusBadRequest, common.FormatResponse{
				Errors: []string{err.Error()},
			})
			return
		}

		formatted, err := executor.Format(req.Query)
		if err != nil {
			s.writeJSON(w, http.StatusOK, common.FormatResponse{
				Errors: []string{err.Error()},
			})
			return
		}

		s.writeJSON(w, http.StatusOK, common.FormatResponse{Formatted: formatted})
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func main() {
	server := NewServer()

	port := 8080
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/graphql", server.handleGraphQL)
	mux.HandleFunc("/api/schema", server.handleSchema)
	mux.HandleFunc("/api/variables", server.handleVariables)
	mux.HandleFunc("/api/validate", server.handleValidate)
	mux.HandleFunc("/api/format", server.handleFormat)

	fmt.Printf("GraphQL server starting on port %d...\n", port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  POST /api/graphql - Execute GraphQL query\n")
	fmt.Printf("  GET  /api/graphql?query=... - Validate syntax only\n")
	fmt.Printf("  GET  /api/schema - Get schema info\n")
	fmt.Printf("  GET  /api/variables - Get preset variables\n")
	fmt.Printf("  POST /api/variables - Set preset variables\n")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
