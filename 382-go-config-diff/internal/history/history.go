package history

import (
    "sync"
    "time"

    "github.com/config-diff/internal/protocol"
)

type HistoryRecord struct {
    ID         string
    File1Path  string
    File2Path  string
    IgnoreKeys []string
    Identical  bool
    Diffs      []protocol.DiffResult
    Timestamp  time.Time
}

type History struct {
    records []HistoryRecord
    maxSize int
    mu      sync.RWMutex
}

func NewHistory(maxSize int) *History {
    if maxSize <= 0 {
        maxSize = 100
    }
    return &History{
        records: make([]HistoryRecord, 0, maxSize),
        maxSize: maxSize,
    }
}

func (h *History) AddRecord(record HistoryRecord) {
    h.mu.Lock()
    defer h.mu.Unlock()

    if len(h.records) >= h.maxSize {
        h.records = h.records[1:]
    }

    h.records = append(h.records, record)
}

func (h *History) GetAllRecords() []HistoryRecord {
    h.mu.RLock()
    defer h.mu.RUnlock()

    records := make([]HistoryRecord, len(h.records))
    copy(records, h.records)
    return records
}

func (h *History) GetLastRecord() *HistoryRecord {
    h.mu.RLock()
    defer h.mu.RUnlock()

    if len(h.records) == 0 {
        return nil
    }

    return &h.records[len(h.records)-1]
}

func (h *History) Clear() {
    h.mu.Lock()
    defer h.mu.Unlock()

    h.records = make([]HistoryRecord, 0, h.maxSize)
}
