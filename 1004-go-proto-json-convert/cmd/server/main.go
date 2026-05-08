package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"protojson.example.com/protojson/internal/common"
	"protojson.example.com/protojson/internal/protobufjson"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

type Server struct {
	types *protoregistry.Types
	mu    sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		types: new(protoregistry.Types),
	}
}

func (s *Server) RegisterMessageType(md protoreflect.MessageDescriptor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	mt := dynamicpb.NewMessageType(md)
	s.types.RegisterMessage(mt)
}

func (s *Server) findMessage(name string) (protoreflect.MessageType, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.types.FindMessageByName(protoreflect.FullName(name))
}

func (s *Server) handleProtoToJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ProtoToJSONRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	mt, err := s.findMessage(req.ProtoName)
	if err != nil {
		resp := common.ProtoToJSONResponse{
			Error: fmt.Sprintf("unknown message type: %s", req.ProtoName),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	msg := mt.New().Interface()

	var msgMap map[string]interface{}
	if err := json.Unmarshal(req.ProtoData, &msgMap); err != nil {
		resp := common.ProtoToJSONResponse{
			Error: fmt.Sprintf("invalid proto data: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	protoBytes, _ := json.Marshal(msgMap)
	if err := protobufjson.Unmarshal(protoBytes, msg); err != nil {
		resp := common.ProtoToJSONResponse{
			Error: fmt.Sprintf("failed to parse proto: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	jsonBytes, err := protobufjson.Marshal(msg)
	if err != nil {
		resp := common.ProtoToJSONResponse{
			Error: fmt.Sprintf("failed to marshal to JSON: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.ProtoToJSONResponse{
		JSONData: json.RawMessage(jsonBytes),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleJSONToProto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.JSONToProtoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	mt, err := s.findMessage(req.ProtoName)
	if err != nil {
		resp := common.JSONToProtoResponse{
			Error: fmt.Sprintf("unknown message type: %s", req.ProtoName),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	msg := mt.New().Interface()

	if err := protobufjson.Unmarshal(req.JSONData, msg); err != nil {
		resp := common.JSONToProtoResponse{
			Error: fmt.Sprintf("failed to unmarshal JSON: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	jsonBytes, err := protobufjson.Marshal(msg)
	if err != nil {
		resp := common.JSONToProtoResponse{
			Error: fmt.Sprintf("failed to marshal proto: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.JSONToProtoResponse{
		ProtoData: json.RawMessage(jsonBytes),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	server := NewServer()

	http.HandleFunc("/health", server.handleHealth)
	http.HandleFunc("/api/v1/proto-to-json", server.handleProtoToJSON)
	http.HandleFunc("/api/v1/json-to-proto", server.handleJSONToProto)

	fmt.Println("Server starting on :8080")
	fmt.Println("Endpoints:")
	fmt.Println("  GET  /health")
	fmt.Println("  POST /api/v1/proto-to-json")
	fmt.Println("  POST /api/v1/json-to-proto")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

var _ = proto.Marshal
