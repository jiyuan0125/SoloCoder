package invoicegen

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func (g *Generator) loadState() error {
	data, err := os.ReadFile(g.storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("WARNING: Storage file not found at %s, starting fresh", g.storagePath)
			g.state = &GeneratorState{
				States: make(map[string]*SequenceState),
			}
			return nil
		}
		log.Printf("WARNING: Failed to read storage file: %v, starting fresh", err)
		g.state = &GeneratorState{
			States: make(map[string]*SequenceState),
		}
		return nil
	}

	var state GeneratorState
	if err := json.Unmarshal(data, &state); err != nil {
		log.Printf("WARNING: Storage file corrupted at %s: %v, starting fresh", g.storagePath, err)
		g.state = &GeneratorState{
			States: make(map[string]*SequenceState),
		}
		return nil
	}

	if state.States == nil {
		state.States = make(map[string]*SequenceState)
	}
	g.state = &state
	return nil
}

func (g *Generator) saveState() error {
	tempPath := g.storagePath + ".tmp"

	data, err := json.MarshalIndent(g.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	dir := filepath.Dir(g.storagePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tempPath, g.storagePath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func (g *Generator) syncPending() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.pendingWrites) == 0 {
		return
	}

	if err := g.saveState(); err != nil {
		log.Printf("WARNING: Failed to sync pending writes: %v", err)
		return
	}

	g.pendingWrites = make(map[string]bool)
}

func (g *Generator) markPending(prefix string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.pendingWrites[prefix] = true
}

func getCurrentDate() string {
	return time.Now().Format(TimestampFormat)
}
