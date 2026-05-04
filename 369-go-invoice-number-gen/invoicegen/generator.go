package invoicegen

import (
	"errors"
	"fmt"
	"log"
	"time"
)

var (
	ErrSequenceExceeded = errors.New("sequence number exceeded maximum value of 99999999")
)

func NewGenerator(storagePath string) (*Generator, error) {
	g := &Generator{
		storagePath:   storagePath,
		pendingWrites: make(map[string]bool),
	}

	if err := g.loadState(); err != nil {
		return nil, err
	}

	return g, nil
}

func (g *Generator) Generate(prefix string) (string, error) {
	effectivePrefix := prefix
	if effectivePrefix == "" {
		effectivePrefix = DefaultPrefix
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	currentDate := getCurrentDate()
	currentYear := time.Now().Format("2006")

	state, exists := g.state.States[effectivePrefix]
	if !exists {
		state = &SequenceState{
			Prefix:   effectivePrefix,
			Sequence: 0,
			LastDate: currentDate,
		}
		g.state.States[effectivePrefix] = state
	}

	if state.LastDate != currentDate {
		state.Sequence = 0
		state.LastDate = currentDate
	}

	newSequence := state.Sequence + 1
	if newSequence > MaxSequenceNumber {
		return "", ErrSequenceExceeded
	}

	state.Sequence = newSequence
	state.LastUpdated = time.Now()

	invoiceNumber := fmt.Sprintf("%s-%s-"+SequenceFormat,
		effectivePrefix,
		currentYear,
		newSequence,
	)

	if err := g.saveState(); err != nil {
		log.Printf("WARNING: Failed to persist state for prefix '%s': %v. Incremented in memory, will retry later.", effectivePrefix, err)
		g.pendingWrites[effectivePrefix] = true
	} else {
		delete(g.pendingWrites, effectivePrefix)
	}

	return invoiceNumber, nil
}

func (g *Generator) GetState(prefix string) (*SequenceState, bool) {
	effectivePrefix := prefix
	if effectivePrefix == "" {
		effectivePrefix = DefaultPrefix
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	state, exists := g.state.States[effectivePrefix]
	if !exists {
		return nil, false
	}

	copyState := *state
	return &copyState, true
}

func (g *Generator) Sync() {
	g.syncPending()
}
