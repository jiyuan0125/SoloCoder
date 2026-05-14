package mapping

import (
	"sync"
)

type FieldMapping struct {
	XMLPath  string `json:"xml_path"`
	JSONPath string `json:"json_path"`
}

type Mapping struct {
	MessageType string         `json:"message_type"`
	XML2JSON    []FieldMapping `json:"xml2json"`
	JSON2XML    []FieldMapping `json:"json2xml"`
}

type Store struct {
	mappings map[string]Mapping
	mu       sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		mappings: make(map[string]Mapping),
	}
}

func (s *Store) Register(mapping Mapping) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := mapping.MessageType
	newKey := string([]byte(key))
	s.mappings[newKey] = mapping
}

func (s *Store) Get(messageType string) (Mapping, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.mappings[messageType]
	return m, ok
}

func (s *Store) Update(messageType string, mapping Mapping) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.mappings[messageType]
	if !exists {
		return false
	}
	mapping.MessageType = messageType
	s.mappings[messageType] = mapping
	return true
}

func (s *Store) Exists(messageType string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.mappings[messageType]
	return ok
}
