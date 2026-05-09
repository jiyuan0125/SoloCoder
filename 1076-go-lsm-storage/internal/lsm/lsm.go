package lsm

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"lsm-storage/internal/memtable"
	"lsm-storage/internal/sstable"
	"lsm-storage/internal/wal"
)

type Config struct {
	Dir               string
	MemTableThreshold int
	MaxLevel           int
	LevelSizeThreshold int
}

type DB struct {
	cfg              Config
	mu               sync.Mutex
	mem              *memtable.MemTable
	memWAL           *wal.WAL
	immu             *memtable.MemTable
	immuWAL          *wal.WAL
	levels           [][]*sstable.Reader
	nextSeq          int64
	compactNotify    chan struct{}
	compactStop      chan struct{}
	compactWg        sync.WaitGroup
	closed           int32
}

func DefaultConfig() Config {
	return Config{
		Dir:               "./data",
		MemTableThreshold: 4 * 1024 * 1024,
		MaxLevel:           7,
		LevelSizeThreshold: 4,
	}
}

func Open(cfg Config) (*DB, error) {
	if cfg.Dir == "" {
		cfg = DefaultConfig()
	}
	if cfg.MemTableThreshold <= 0 {
		cfg.MemTableThreshold = DefaultConfig().MemTableThreshold
	}
	if cfg.MaxLevel <= 0 {
		cfg.MaxLevel = DefaultConfig().MaxLevel
	}
	if cfg.LevelSizeThreshold <= 0 {
		cfg.LevelSizeThreshold = DefaultConfig().LevelSizeThreshold
	}

	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, err
	}
	walDir := filepath.Join(cfg.Dir, "wal")
	sstDir := filepath.Join(cfg.Dir, "sst")
	if err := os.MkdirAll(walDir, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(sstDir, 0755); err != nil {
		return nil, err
	}

	db := &DB{
		cfg:           cfg,
		mem:           memtable.New(cfg.MemTableThreshold),
		levels:        make([][]*sstable.Reader, cfg.MaxLevel),
		compactNotify: make(chan struct{}, 1),
		compactStop:   make(chan struct{}),
	}

	nextSeq, err := db.recover()
	if err != nil {
		return nil, err
	}
	db.nextSeq = nextSeq

	walPath := wal.GenPath(walDir, nextSeq)
	w, err := wal.New(walPath)
	if err != nil {
		return nil, err
	}
	db.memWAL = w

	db.compactWg.Add(1)
	go db.compactLoop()

	return db, nil
}

func (db *DB) recover() (int64, error) {
	walDir := filepath.Join(db.cfg.Dir, "wal")
	sstDir := filepath.Join(db.cfg.Dir, "sst")

	sstFiles, err := sstable.List(sstDir)
	if err != nil {
		return 0, err
	}

	var maxSeq int64 = 0
	for _, path := range sstFiles {
		seq := parseSeq(path)
		if seq > maxSeq {
			maxSeq = seq
		}
		level := parseLevel(path)
		if level >= 0 && level < db.cfg.MaxLevel {
			r, err := sstable.NewReader(path)
			if err != nil {
				return 0, err
			}
			db.levels[level] = append(db.levels[level], r)
		}
	}

	walFiles, err := wal.List(walDir)
	if err != nil {
		return 0, err
	}

	for _, path := range walFiles {
		seq := parseWALSeq(path)
		if seq > maxSeq {
			maxSeq = seq
		}
		if seq > db.nextSeq {
			entries, err := wal.Load(path)
			if err != nil {
				return 0, err
			}
			for _, e := range entries {
				if e.Tombstone {
					db.mem.Delete(e.Key)
				} else {
					db.mem.Put(e.Key, e.Value)
				}
			}
		}
	}

	return maxSeq + 1, nil
}

func parseSeq(path string) int64 {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	parts := strings.Split(name, "_")
	if len(parts) != 2 {
		return 0
	}
	seq, _ := strconv.ParseInt(parts[1], 10, 64)
	return seq
}

func parseLevel(path string) int {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	parts := strings.Split(name, "_")
	if len(parts) != 2 || len(parts[0]) < 2 {
		return -1
	}
	level, _ := strconv.Atoi(parts[0][1:])
	return level
}

func parseWALSeq(path string) int64 {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	seq, _ := strconv.ParseInt(name, 10, 64)
	return seq
}

func (db *DB) Put(key, value []byte) error {
	if atomic.LoadInt32(&db.closed) != 0 {
		return fmt.Errorf("db closed")
	}
	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.memWAL.Append(&wal.Entry{Key: key, Value: value, Tombstone: false}); err != nil {
		return err
	}
	db.mem.Put(key, value)

	if db.mem.ShouldFlush() {
		if err := db.switchMem(); err != nil {
			return err
		}
		db.notifyCompact()
	}
	return nil
}

func (db *DB) Delete(key []byte) error {
	if atomic.LoadInt32(&db.closed) != 0 {
		return fmt.Errorf("db closed")
	}
	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.memWAL.Append(&wal.Entry{Key: key, Value: nil, Tombstone: true}); err != nil {
		return err
	}
	db.mem.Delete(key)

	if db.mem.ShouldFlush() {
		if err := db.switchMem(); err != nil {
			return err
		}
		db.notifyCompact()
	}
	return nil
}

func (db *DB) Get(key []byte) ([]byte, bool, error) {
	if atomic.LoadInt32(&db.closed) != 0 {
		return nil, false, fmt.Errorf("db closed")
	}
	db.mu.Lock()
	mem := db.mem
	immu := db.immu
	db.mu.Unlock()

	val, tomb, ok := mem.Get(key)
	if ok {
		return val, !tomb, nil
	}

	if immu != nil {
		val, tomb, ok = immu.Get(key)
		if ok {
			return val, !tomb, nil
		}
	}

	for level := 0; level < db.cfg.MaxLevel; level++ {
		db.mu.Lock()
		readers := db.levels[level]
		db.mu.Unlock()
		for i := len(readers) - 1; i >= 0; i-- {
			r := readers[i]
			if !r.MayContain(key) {
				continue
			}
			val, tomb, ok = r.Get(key)
			if ok {
				return val, !tomb, nil
			}
		}
	}

	return nil, false, nil
}

func (db *DB) Scan(start, end []byte) []*KVPair {
	if atomic.LoadInt32(&db.closed) != 0 {
		return nil
	}
	seen := make(map[string]struct{})
	var result []*KVPair

	db.mu.Lock()
	mem := db.mem
	immu := db.immu
	db.mu.Unlock()

	it := mem.Iterator()
	for it.Valid() {
		key := it.Key()
		if inRange(key, start, end) {
			if !it.Tombstone() {
				k := string(key)
				if _, ok := seen[k]; !ok {
					seen[k] = struct{}{}
					result = append(result, &KVPair{Key: key, Value: it.Value()})
				}
			}
		}
		it.Next()
	}

	if immu != nil {
		it = immu.Iterator()
		for it.Valid() {
			key := it.Key()
			if inRange(key, start, end) {
				k := string(key)
				if _, ok := seen[k]; !ok {
					seen[k] = struct{}{}
					if !it.Tombstone() {
						result = append(result, &KVPair{Key: key, Value: it.Value()})
					}
				}
			}
			it.Next()
		}
	}

	for level := 0; level < db.cfg.MaxLevel; level++ {
		db.mu.Lock()
		readers := db.levels[level]
		db.mu.Unlock()
		for i := len(readers) - 1; i >= 0; i-- {
			records := readers[i].Scan(start, end)
			for _, rec := range records {
				k := string(rec.Key)
				if _, ok := seen[k]; !ok {
					seen[k] = struct{}{}
					if !rec.Tombstone {
						result = append(result, &KVPair{Key: rec.Key, Value: rec.Value})
					}
				}
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return compare(result[i].Key, result[j].Key) < 0
	})
	return result
}

type KVPair struct {
	Key   []byte
	Value []byte
}

func inRange(key, start, end []byte) bool {
	if start != nil && compare(key, start) < 0 {
		return false
	}
	if end != nil && compare(key, end) >= 0 {
		return false
	}
	return true
}

func compare(a, b []byte) int {
	la, lb := len(a), len(b)
	minLen := la
	if lb < minLen {
		minLen = lb
	}
	for i := 0; i < minLen; i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	if la < lb {
		return -1
	} else if la > lb {
		return 1
	}
	return 0
}

func (db *DB) switchMem() error {
	oldMem := db.mem
	oldWAL := db.memWAL

	if err := oldWAL.Close(); err != nil {
		return err
	}

	db.immu = oldMem
	db.immuWAL = oldWAL

	db.mem = memtable.New(db.cfg.MemTableThreshold)

	walDir := filepath.Join(db.cfg.Dir, "wal")
	newSeq := atomic.AddInt64(&db.nextSeq, 1) - 1
	walPath := wal.GenPath(walDir, newSeq)
	w, err := wal.New(walPath)
	if err != nil {
		return err
	}
	db.memWAL = w
	return nil
}

func (db *DB) notifyCompact() {
	select {
	case db.compactNotify <- struct{}{}:
	default:
	}
}

func (db *DB) flushImmutable() {
	db.mu.Lock()
	immu := db.immu
	immuWAL := db.immuWAL
	db.immu = nil
	db.immuWAL = nil
	db.mu.Unlock()

	if immu == nil {
		return
	}

	sstDir := filepath.Join(db.cfg.Dir, "sst")
	seq := atomic.AddInt64(&db.nextSeq, 1) - 1
	path := sstable.GenPath(sstDir, 0, seq)

	w, err := sstable.NewWriter(path)
	if err != nil {
		return
	}

	it := immu.Iterator()
	for it.Valid() {
		w.Append(it.Key(), it.Value(), it.Tombstone())
		it.Next()
	}
	if err := w.Finish(); err != nil {
		return
	}

	r, err := sstable.NewReader(path)
	if err != nil {
		return
	}

	db.mu.Lock()
	db.levels[0] = append(db.levels[0], r)
	db.mu.Unlock()

	_ = immuWAL.Delete()
}

func (db *DB) compactLoop() {
	defer db.compactWg.Done()
	for {
		select {
		case <-db.compactNotify:
			db.flushImmutable()
			db.runCompaction()
		case <-db.compactStop:
			return
		}
	}
}

func (db *DB) runCompaction() {
	for level := 0; level < db.cfg.MaxLevel-1; level++ {
		db.mu.Lock()
		count := len(db.levels[level])
		db.mu.Unlock()
		if count >= db.cfg.LevelSizeThreshold {
			db.compactLevel(level)
		}
	}
}

func (db *DB) compactLevel(level int) {
	db.mu.Lock()
	srcReaders := db.levels[level]
	db.levels[level] = nil
	db.mu.Unlock()

	if len(srcReaders) == 0 {
		return
	}

	iters := make([]*sstable.Iterator, len(srcReaders))
	for i, r := range srcReaders {
		iters[i] = r.Iterator()
	}

	merged := mergeIterators(iters, level == db.cfg.MaxLevel-1)

	sstDir := filepath.Join(db.cfg.Dir, "sst")
	seq := atomic.AddInt64(&db.nextSeq, 1) - 1
	path := sstable.GenPath(sstDir, level+1, seq)

	w, err := sstable.NewWriter(path)
	if err != nil {
		return
	}
	for _, rec := range merged {
		w.Append(rec.Key, rec.Value, rec.Tombstone)
	}
	if err := w.Finish(); err != nil {
		return
	}

	r, err := sstable.NewReader(path)
	if err != nil {
		return
	}

	db.mu.Lock()
	db.levels[level+1] = append(db.levels[level+1], r)
	db.mu.Unlock()

	for _, sr := range srcReaders {
		sr.Close()
		os.Remove(sr.Path())
	}
}

func mergeIterators(iters []*sstable.Iterator, isBottom bool) []sstable.Record {
	var records []sstable.Record
	seen := make(map[string]sstable.Record)
	for _, it := range iters {
		for it.Valid() {
			key := string(it.Key())
			if _, ok := seen[key]; !ok {
				seen[key] = sstable.Record{
					Key:       it.Key(),
					Value:     it.Value(),
					Tombstone: it.Tombstone(),
				}
			}
			it.Next()
		}
	}
	for _, rec := range seen {
		if isBottom && rec.Tombstone {
			continue
		}
		records = append(records, rec)
	}
	sort.Slice(records, func(i, j int) bool {
		return compare(records[i].Key, records[j].Key) < 0
	})
	return records
}

func (db *DB) Close() error {
	if !atomic.CompareAndSwapInt32(&db.closed, 0, 1) {
		return nil
	}

	close(db.compactStop)
	db.compactWg.Wait()

	db.mu.Lock()
	if db.memWAL != nil {
		db.memWAL.Close()
	}
	if db.immuWAL != nil {
		db.immuWAL.Close()
	}
	for _, level := range db.levels {
		for _, r := range level {
			r.Close()
		}
	}
	db.mu.Unlock()

	return nil
}
