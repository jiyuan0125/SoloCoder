package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"go-protocol-adapter/common"
	"go-protocol-adapter/core/adapter"
	"go-protocol-adapter/core/business"
)

type Server struct {
	registry *adapter.AdapterRegistry
	store    *business.UserStore
}

func NewServer() *Server {
	registry := adapter.NewRegistry()
	registry.Register(adapter.NewJSONAdapter())
	registry.Register(adapter.NewXMLAdapter())
	registry.Register(adapter.NewCSVAdapter())

	return &Server{
		registry: registry,
		store:    business.NewUserStore(),
	}
}

func (s *Server) listAdapters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, common.CodeInvalidRequest, "method not allowed")
		return
	}

	adapters := s.registry.List()
	writeSuccess(w, adapters)
}

func (s *Server) registerAdapter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, common.CodeInvalidRequest, "method not allowed")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, common.CodeInvalidRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	var config adapter.DynamicAdapterConfig
	if err := json.Unmarshal(body, &config); err != nil {
		writeError(w, common.CodeInvalidRequest, fmt.Sprintf("invalid JSON: %v", err))
		return
	}

	newAdapter, err := adapter.NewDynamicAdapter(config)
	if err != nil {
		writeError(w, common.CodeInvalidRequest, err.Error())
		return
	}

	if err := s.registry.Register(newAdapter); err != nil {
		writeError(w, common.CodeAdapterExists, err.Error())
		return
	}

	writeSuccess(w, map[string]string{"status": "registered", "adapter": config.Name})
}

func (s *Server) encode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, common.CodeInvalidRequest, "method not allowed")
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		writeError(w, common.CodeFormatNotSupport, "format parameter is required")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, common.CodeInvalidRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		writeError(w, common.CodeDecodeFailed, fmt.Sprintf("invalid JSON: %v", err))
		return
	}

	enc, ok := s.registry.Find(format)
	if !ok {
		writeError(w, common.CodeFormatNotSupport, fmt.Sprintf("no adapter found for format: %s", format))
		return
	}

	encoded, err := enc.Encode(data)
	if err != nil {
		writeError(w, common.CodeEncodeFailed, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.Response{
		Code:    int(common.CodeSuccess),
		Message: common.CodeSuccess.String(),
		Data:    string(encoded),
	})
}

func (s *Server) decode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, common.CodeInvalidRequest, "method not allowed")
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		writeError(w, common.CodeFormatNotSupport, "format parameter is required")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, common.CodeInvalidRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	dec, ok := s.registry.Find(format)
	if !ok {
		writeError(w, common.CodeFormatNotSupport, fmt.Sprintf("no adapter found for format: %s", format))
		return
	}

	var result map[string]interface{}
	if err := dec.Decode(body, &result); err != nil {
		writeError(w, common.CodeDecodeFailed, err.Error())
		return
	}

	writeSuccess(w, result)
}

type ExecuteRequest struct {
	InputFormat  string          `json:"input_format"`
	OutputFormat string          `json:"output_format"`
	Operation    string          `json:"operation"`
	Data         json.RawMessage `json:"data"`
}

func (s *Server) execute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, common.CodeInvalidRequest, "method not allowed")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, common.CodeInvalidRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	var execReq ExecuteRequest
	if err := json.Unmarshal(body, &execReq); err != nil {
		writeError(w, common.CodeInvalidRequest, fmt.Sprintf("invalid JSON: %v", err))
		return
	}

	if execReq.InputFormat == "" {
		writeError(w, common.CodeFormatNotSupport, "input_format is required")
		return
	}
	if execReq.OutputFormat == "" {
		writeError(w, common.CodeFormatNotSupport, "output_format is required")
		return
	}
	if execReq.Operation == "" {
		writeError(w, common.CodeInvalidRequest, "operation is required")
		return
	}

	inputAdapter, ok := s.registry.Find(execReq.InputFormat)
	if !ok {
		writeError(w, common.CodeFormatNotSupport, fmt.Sprintf("no adapter found for input format: %s", execReq.InputFormat))
		return
	}

	outputAdapter, ok := s.registry.Find(execReq.OutputFormat)
	if !ok {
		writeError(w, common.CodeFormatNotSupport, fmt.Sprintf("no adapter found for output format: %s", execReq.OutputFormat))
		return
	}

	var result interface{}
	switch strings.ToLower(execReq.Operation) {
	case "create":
		var user business.User
		if err := inputAdapter.Decode(execReq.Data, &user); err != nil {
			writeError(w, common.CodeDecodeFailed, err.Error())
			return
		}
		created, err := s.store.Create(&user)
		if err != nil {
			writeError(w, common.CodeBusinessError, err.Error())
			return
		}
		result = created

	case "get":
		idStr := strings.TrimSpace(string(execReq.Data))
		idStr = strings.Trim(idStr, "\"")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, common.CodeInvalidRequest, fmt.Sprintf("invalid user ID: %s", idStr))
			return
		}
		user, err := s.store.Get(id)
		if err != nil {
			writeError(w, common.CodeBusinessError, err.Error())
			return
		}
		result = user

	case "update":
		var user business.User
		if err := inputAdapter.Decode(execReq.Data, &user); err != nil {
			writeError(w, common.CodeDecodeFailed, err.Error())
			return
		}
		updated, err := s.store.Update(&user)
		if err != nil {
			writeError(w, common.CodeBusinessError, err.Error())
			return
		}
		result = updated

	case "delete":
		idStr := strings.TrimSpace(string(execReq.Data))
		idStr = strings.Trim(idStr, "\"")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, common.CodeInvalidRequest, fmt.Sprintf("invalid user ID: %s", idStr))
			return
		}
		if err := s.store.Delete(id); err != nil {
			writeError(w, common.CodeBusinessError, err.Error())
			return
		}
		result = map[string]string{"status": "deleted"}

	case "list":
		result = s.store.List()

	default:
		writeError(w, common.CodeInvalidRequest, fmt.Sprintf("unknown operation: %s", execReq.Operation))
		return
	}

	encoded, err := outputAdapter.Encode(result)
	if err != nil {
		writeError(w, common.CodeEncodeFailed, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.Response{
		Code:    int(common.CodeSuccess),
		Message: common.CodeSuccess.String(),
		Data:    string(encoded),
	})
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.Success(data))
}

func writeError(w http.ResponseWriter, code common.ErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.Error(code, message))
}

func main() {
	server := NewServer()

	http.HandleFunc("/adapters", server.listAdapters)
	http.HandleFunc("/adapters/register", server.registerAdapter)
	http.HandleFunc("/encode", server.encode)
	http.HandleFunc("/decode", server.decode)
	http.HandleFunc("/execute", server.execute)

	fmt.Println("Server starting on :8404...")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /adapters           - List all registered adapters")
	fmt.Println("  POST /adapters/register  - Register a new adapter")
	fmt.Println("  POST /encode?format=xxx  - Encode data to specified format")
	fmt.Println("  POST /decode?format=xxx  - Decode data from specified format")
	fmt.Println("  POST /execute            - Execute decode + business + encode")

	if err := http.ListenAndServe(":8404", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
