package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sync"
    "time"

    "batch-rename/protocol"
)

type HistoryManager struct {
    records   []protocol.HistoryRecord
    nextID    int64
    mu        sync.RWMutex
    storagePath string
}

var globalHistory *HistoryManager

func init() {
    globalHistory = NewHistoryManager("")
}

func NewHistoryManager(storagePath string) *HistoryManager {
    hm := &HistoryManager{
        records:     make([]protocol.HistoryRecord, 0),
        nextID:      1,
        storagePath: storagePath,
    }

    if storagePath != "" {
        hm.loadFromFile()
    }

    return hm
}

func (hm *HistoryManager) AddRecord(items []protocol.RenameItem, success, failed, skipped int) {
    hm.mu.Lock()
    defer hm.mu.Unlock()

    record := protocol.HistoryRecord{
        ID:        hm.nextID,
        Timestamp: time.Now().Unix(),
        Items:     items,
        Success:   success,
        Failed:    failed,
        Skipped:   skipped,
    }

    hm.records = append(hm.records, record)
    hm.nextID++

    if hm.storagePath != "" {
        hm.saveToFile()
    }
}

func (hm *HistoryManager) GetRecords(limit int) []protocol.HistoryRecord {
    hm.mu.RLock()
    defer hm.mu.RUnlock()

    if limit <= 0 || limit > len(hm.records) {
        limit = len(hm.records)
    }

    start := len(hm.records) - limit
    result := make([]protocol.HistoryRecord, limit)
    copy(result, hm.records[start:])

    return result
}

func (hm *HistoryManager) GetAllRecords() []protocol.HistoryRecord {
    hm.mu.RLock()
    defer hm.mu.RUnlock()

    result := make([]protocol.HistoryRecord, len(hm.records))
    copy(result, hm.records)

    return result
}

func (hm *HistoryManager) saveToFile() {
    if hm.storagePath == "" {
        return
    }

    dir := filepath.Dir(hm.storagePath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return
    }

    data, err := json.MarshalIndent(hm.records, "", "  ")
    if err != nil {
        return
    }

    os.WriteFile(hm.storagePath, data, 0644)
}

func (hm *HistoryManager) loadFromFile() {
    if hm.storagePath == "" {
        return
    }

    data, err := os.ReadFile(hm.storagePath)
    if err != nil {
        return
    }

    var records []protocol.HistoryRecord
    if err := json.Unmarshal(data, &records); err != nil {
        return
    }

    hm.records = records
    if len(records) > 0 {
        hm.nextID = records[len(records)-1].ID + 1
    }
}

func AddHistoryRecord(items []protocol.RenameItem, success, failed, skipped int) {
    globalHistory.AddRecord(items, success, failed, skipped)
}

func GetHistoryRecords(limit int) []protocol.HistoryRecord {
    return globalHistory.GetRecords(limit)
}

func GetAllHistoryRecords() []protocol.HistoryRecord {
    return globalHistory.GetAllRecords()
}
