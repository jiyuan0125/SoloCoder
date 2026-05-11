package lsm

import (
	"errors"
	"fmt"
	"lsm-tree/common"
	"os"
	"sort"
	"sync"
)

type LSMTreeConfig struct {
	MemTableSize   int
	DataDir        string
	Strategy       common.CompactionStrategy
	MaxLevels      int
	SizeTieredMin  int
	LeveledMaxFiles int
}

var DefaultConfig = LSMTreeConfig{
	MemTableSize:    1024 * 1024,
	DataDir:         "./data",
	Strategy:        common.SizeTieredStrategy,
	MaxLevels:       5,
	SizeTieredMin:   4,
	LeveledMaxFiles: 10,
}

type LSMTree struct {
	config      LSMTreeConfig
	memTable    *MemTable
	immutable   *MemTable
	levels      map[int][]*SSTable
	strategy    CompactionStrategy
	stats       *CompactionStats
	seqNum      uint64
	mu          sync.RWMutex
}

func NewLSMTree(config LSMTreeConfig) (*LSMTree, error) {
	if config.MemTableSize <= 0 {
		config.MemTableSize = DefaultConfig.MemTableSize
	}
	if config.DataDir == "" {
		config.DataDir = DefaultConfig.DataDir
	}
	if config.MaxLevels <= 0 {
		config.MaxLevels = DefaultConfig.MaxLevels
	}
	if config.SizeTieredMin <= 0 {
		config.SizeTieredMin = DefaultConfig.SizeTieredMin
	}
	if config.LeveledMaxFiles <= 0 {
		config.LeveledMaxFiles = DefaultConfig.LeveledMaxFiles
	}
	
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return nil, err
	}
	
	levels, err := ListSSTables(config.DataDir)
	if err != nil {
		return nil, err
	}
	
	tree := &LSMTree{
		config:   config,
		memTable: NewMemTable(config.MemTableSize),
		levels:   levels,
		stats:    NewCompactionStats(),
	}
	
	if err := tree.setStrategy(config.Strategy); err != nil {
		return nil, err
	}
	
	return tree, nil
}

func (t *LSMTree) setStrategy(strategy common.CompactionStrategy) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	switch strategy {
	case common.SizeTieredStrategy:
		t.strategy = NewSizeTieredCompaction(t.config.SizeTieredMin, t.config.MaxLevels)
	case common.LeveledStrategy:
		t.strategy = NewLeveledCompaction(t.config.LeveledMaxFiles, t.config.MaxLevels)
	default:
		return fmt.Errorf("unknown compaction strategy: %s", strategy)
	}
	
	t.config.Strategy = strategy
	return nil
}

func (t *LSMTree) SwitchStrategy(strategy common.CompactionStrategy) error {
	return t.setStrategy(strategy)
}

func (t *LSMTree) GetStrategy() common.CompactionStrategy {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.config.Strategy
}

func (t *LSMTree) Put(key, value string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.seqNum++
	t.memTable.Put(key, value, t.seqNum)
	
	if t.memTable.IsFull() {
		if err := t.flushLocked(); err != nil {
			return err
		}
	}
	
	entry := &Entry{
		Key:     key,
		Value:   value,
		Deleted: false,
		SeqNum:  t.seqNum,
	}
	t.stats.RecordUserWrite(int64(entry.Size()))
	
	return nil
}

func (t *LSMTree) Delete(key string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.seqNum++
	t.memTable.Delete(key, t.seqNum)
	
	if t.memTable.IsFull() {
		if err := t.flushLocked(); err != nil {
			return err
		}
	}
	
	entry := &Entry{
		Key:     key,
		Deleted: true,
		SeqNum:  t.seqNum,
	}
	t.stats.RecordUserWrite(int64(entry.Size()))
	
	return nil
}

func (t *LSMTree) Get(key string) (string, bool, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	if entry, exists := t.memTable.Get(key); exists {
		if entry.Deleted {
			return "", false, nil
		}
		return entry.Value, true, nil
	}
	
	if t.immutable != nil {
		if entry, exists := t.immutable.Get(key); exists {
			if entry.Deleted {
				return "", false, nil
			}
			return entry.Value, true, nil
		}
	}
	
	levelKeys := make([]int, 0, len(t.levels))
	for k := range t.levels {
		levelKeys = append(levelKeys, k)
	}
	sort.Ints(levelKeys)
	
	for _, level := range levelKeys {
		tables := t.levels[level]
		
		for _, table := range tables {
			if !table.ContainsKeyRange(key) {
				continue
			}
			
			entry, found, err := table.Get(key)
			if err != nil {
				return "", false, err
			}
			
			if found {
				if entry.Deleted {
					return "", false, nil
				}
				return entry.Value, true, nil
			}
		}
	}
	
	return "", false, nil
}

func (t *LSMTree) Range(start, end string) ([]*Entry, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	allEntries := make([]*Entry, 0)
	
	memEntries := t.memTable.Range(start, end)
	allEntries = append(allEntries, memEntries...)
	
	if t.immutable != nil {
		immEntries := t.immutable.Range(start, end)
		allEntries = append(allEntries, immEntries...)
	}
	
	levelKeys := make([]int, 0, len(t.levels))
	for k := range t.levels {
		levelKeys = append(levelKeys, k)
	}
	sort.Ints(levelKeys)
	
	for _, level := range levelKeys {
		for _, table := range t.levels[level] {
			if (table.MaxKey() < start) || (end != "" && table.MinKey() > end) {
				continue
			}
			
			entries, err := table.Range(start, end)
			if err != nil {
				return nil, err
			}
			allEntries = append(allEntries, entries...)
		}
	}
	
	return MergeEntries(allEntries), nil
}

func (t *LSMTree) flushLocked() error {
	t.immutable = t.memTable
	t.memTable = NewMemTable(t.config.MemTableSize)
	
	go func() {
		t.flushImmutable()
	}()
	
	return nil
}

func (t *LSMTree) flushImmutable() {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if t.immutable == nil {
		return
	}
	
	entries := t.immutable.Entries()
	if len(entries) == 0 {
		t.immutable = nil
		return
	}
	
	table, err := NewSSTable(0, t.config.DataDir)
	if err != nil {
		return
	}
	
	if err := table.Write(entries); err != nil {
		return
	}
	
	t.stats.RecordCompactionWrite(table.Size())
	t.levels[0] = append(t.levels[0], table)
	t.immutable = nil
	
	if t.strategy.ShouldCompact(t.levels) {
		go t.BackgroundCompact()
	}
}

func (t *LSMTree) BackgroundCompact() {
	t.mu.Lock()
	if !t.strategy.ShouldCompact(t.levels) {
		t.mu.Unlock()
		return
	}
	t.mu.Unlock()
	
	for {
		t.mu.Lock()
		if !t.strategy.ShouldCompact(t.levels) {
			t.mu.Unlock()
			break
		}
		
		newLevels, written, err := t.strategy.Compact(t.levels, t.config.DataDir)
		if err != nil {
			t.mu.Unlock()
			return
		}
		
		t.levels = newLevels
		if written > 0 {
			t.stats.RecordCompactionWrite(written)
		}
		t.mu.Unlock()
	}
}

func (t *LSMTree) Compact() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	for t.strategy.ShouldCompact(t.levels) {
		newLevels, written, err := t.strategy.Compact(t.levels, t.config.DataDir)
		if err != nil {
			return err
		}
		
		t.levels = newLevels
		if written > 0 {
			t.stats.RecordCompactionWrite(written)
		}
	}
	
	return nil
}

func (t *LSMTree) GetStats() *CompactionStats {
	return t.stats
}

func (t *LSMTree) GetLevelStats() []common.LevelStats {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	stats := make([]common.LevelStats, 0)
	maxLevel := 0
	for level := range t.levels {
		if level > maxLevel {
			maxLevel = level
		}
	}
	
	for level := 0; level <= maxLevel; level++ {
		tables := t.levels[level]
		var totalSize int64 = 0
		for _, t := range tables {
			totalSize += t.Size()
		}
		
		stats = append(stats, common.LevelStats{
			Level:     level,
			FileCount: len(tables),
			TotalSize: totalSize,
		})
	}
	
	return stats
}

func (t *LSMTree) Config() LSMTreeConfig {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.config
}

func (t *LSMTree) Close() error {
	return nil
}

var (
	ErrNotFound = errors.New("key not found")
)
