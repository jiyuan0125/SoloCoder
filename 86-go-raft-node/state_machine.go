package main

import (
	"encoding/json"
	"fmt"
	"sync"
)

type StateMachine struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewStateMachine() *StateMachine {
	return &StateMachine{
		data: make(map[string]string),
	}
}

func (sm *StateMachine) Apply(entry *LogEntry) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if entry.Command == "SET" {
		sm.data[entry.Key] = entry.Value
	}
}

func (sm *StateMachine) Get(key string) (string, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	value, exists := sm.data[key]
	return value, exists
}

func (sm *StateMachine) CreateSnapshot() ([]byte, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	return json.Marshal(sm.data)
}

func (sm *StateMachine) RestoreSnapshot(snapshot []byte) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	var data map[string]string
	if err := json.Unmarshal(snapshot, &data); err != nil {
		return err
	}
	
	sm.data = data
	return nil
}

func (sm *StateMachine) String() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	return fmt.Sprintf("%v", sm.data)
}
