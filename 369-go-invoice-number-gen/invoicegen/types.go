package invoicegen

import (
	"sync"
	"time"
)

const (
	DefaultPrefix     = "INV"
	MaxSequenceNumber = 99999999
	SequenceFormat    = "%08d"
	TimestampFormat   = "2006-01-02"
)

type SequenceState struct {
	Prefix      string    `json:"prefix"`
	Sequence    uint64    `json:"sequence"`
	LastDate    string    `json:"last_date"`
	LastUpdated time.Time `json:"last_updated"`
}

type GeneratorState struct {
	States map[string]*SequenceState `json:"states"`
}

type Generator struct {
	mu            sync.RWMutex
	state         *GeneratorState
	storagePath   string
	pendingWrites map[string]bool
}
