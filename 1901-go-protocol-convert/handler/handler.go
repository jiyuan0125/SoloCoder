package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"

	"json-protobuf-gateway/converter"
	"json-protobuf-gateway/schema"
	"json-protobuf-gateway/types"
)

type Handler struct {
	store     *types.SchemaStore
	parser    *schema.SimpleProtoParser
	converter *converter.Converter
}

func NewHandler(store *types.SchemaStore) *Handler {
	return &Handler{
		store:     store,
		parser:    schema.NewSimpleProtoParser(),
		converter: converter.NewConverter(),
	}
}

type RegisterRequest struct {
	ProtoDefinition string            `json:"proto_definition"`
	MessageName     string            `json:"message_name"`
	FieldMappings   map[string]string `json:"field_mappings"`
	BackendURL      string            `json:"backend_url,omitempty"`
}

func (h *Handler) RegisterSchema(c *gin.Context) {
	messageType := c.Param("message_type")
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}
	if req.ProtoDefinition == "" || req.MessageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "proto_definition and message_name are required"})
		return
	}

	desc, err := h.parser.Parse(req.ProtoDefinition, req.MessageName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse proto definition: " + err.Error()})
		return
	}

	reverseMappings := make(map[string]string)
	for jsonField, protoField := range req.FieldMappings {
		reverseMappings[protoField] = jsonField
	}

	s := &types.Schema{
		MessageName:     req.MessageName,
		FieldMappings:   req.FieldMappings,
		ReverseMappings: reverseMappings,
		MessageDesc:     desc,
		BackendURL:      req.BackendURL,
	}

	h.store.Set(messageType, s)
	c.JSON(http.StatusCreated, gin.H{
		"status":       "registered",
		"message_type": messageType,
		"message_name": req.MessageName,
	})
}

func (h *Handler) UpdateSchema(c *gin.Context) {
	messageType := c.Param("message_type")
	if _, exists := h.store.Get(messageType); !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "message type not registered"})
		return
	}
	h.RegisterSchema(c)
}

func (h *Handler) ConvertForward(c *gin.Context) {
	messageType := c.Param("message_type")
	s, ok := h.store.Get(messageType)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "message type not registered"})
		return
	}

	var jsonData map[string]interface{}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
		return
	}

	msg, err := h.converter.JSONToProto(jsonData, s.FieldMappings, s.MessageDesc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to convert JSON to proto: " + err.Error()})
		return
	}

	protoBytes, err := proto.Marshal(msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal proto: " + err.Error()})
		return
	}

	if s.BackendURL == "" {
		c.Data(http.StatusOK, "application/octet-stream", protoBytes)
		return
	}

	respBytes, err := h.forwardToBackend(s.BackendURL, protoBytes)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "backend request failed: " + err.Error()})
		return
	}

	respMsg := dynamicpb.NewMessage(s.MessageDesc)
	if err := proto.Unmarshal(respBytes, respMsg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unmarshal backend response: " + err.Error()})
		return
	}

	jsonResp, err := h.converter.ProtoToJSON(respMsg, s.ReverseMappings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to convert proto to JSON: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonResp)
}

func (h *Handler) ConvertReverse(c *gin.Context) {
	messageType := c.Param("message_type")
	s, ok := h.store.Get(messageType)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "message type not registered"})
		return
	}

	protoBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	msg := dynamicpb.NewMessage(s.MessageDesc)
	if err := proto.Unmarshal(protoBytes, msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid protobuf: " + err.Error()})
		return
	}

	jsonData, err := h.converter.ProtoToJSON(msg, s.ReverseMappings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to convert proto to JSON: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonData)
}

func (h *Handler) forwardToBackend(url string, data []byte) ([]byte, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("backend returned status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
