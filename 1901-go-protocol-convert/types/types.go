package types

import (
	"sync"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type FieldMapping struct {
	JSONField  string            `json:"json_field"`
	ProtoField string            `json:"proto_field"`
	Nested     map[string]string `json:"nested,omitempty"`
}

type SchemaRegistration struct {
	ProtoDefinition string            `json:"proto_definition"`
	MessageName     string            `json:"message_name"`
	FieldMappings   map[string]string `json:"field_mappings"`
	BackendURL      string            `json:"backend_url,omitempty"`
}

type Schema struct {
	MessageName    string
	FieldMappings  map[string]string
	ReverseMappings map[string]string
	MessageDesc    protoreflect.MessageDescriptor
	BackendURL     string
}

type SchemaStore struct {
	mu      sync.RWMutex
	schemas map[string]*Schema
}

func NewSchemaStore() *SchemaStore {
	return &SchemaStore{
		schemas: make(map[string]*Schema),
	}
}

func (s *SchemaStore) Get(messageType string) (*Schema, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	schema, ok := s.schemas[messageType]
	return schema, ok
}

func (s *SchemaStore) Set(messageType string, schema *Schema) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schemas[messageType] = schema
}

func (s *SchemaStore) Delete(messageType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.schemas, messageType)
}
