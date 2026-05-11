package lsm

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type SSTableHeader struct {
	Magic    [4]byte
	Version  uint32
	EntryCount uint32
	MinKey   string
	MaxKey   string
}

type SSTableIndex struct {
	Key    string
	Offset int64
}

type SSTable struct {
	filename   string
	path       string
	level      int
	size       int64
	entries    uint32
	minKey     string
	maxKey     string
	index      []SSTableIndex
	keyOffsets map[string]int64
	mu         sync.RWMutex
	loaded     bool
}

func NewSSTable(level int, dataDir string) (*SSTable, error) {
	id := generateTableID()
	filename := fmt.Sprintf("table-%s.sst", id)
	path := filepath.Join(dataDir, fmt.Sprintf("level-%d", level), filename)
	
	levelDir := filepath.Join(dataDir, fmt.Sprintf("level-%d", level))
	if err := os.MkdirAll(levelDir, 0755); err != nil {
		return nil, err
	}
	
	return &SSTable{
		filename:   filename,
		path:       path,
		level:      level,
		keyOffsets: make(map[string]int64),
	}, nil
}

func generateTableID() string {
	h := sha256.New()
	timestamp := make([]byte, 8)
	binary.BigEndian.PutUint64(timestamp, uint64(os.Getpid()))
	h.Write(timestamp)
	return hex.EncodeToString(h.Sum(nil))[:12]
}

func (s *SSTable) Write(entries []*Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	f, err := os.Create(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	
	writer := bufio.NewWriter(f)
	
	var minKey, maxKey string
	if len(entries) > 0 {
		minKey = entries[0].Key
		maxKey = entries[len(entries)-1].Key
	}
	
	offset := int64(0)
	index := make([]SSTableIndex, 0, len(entries))
	
	for _, e := range entries {
		if e.Key < minKey {
			minKey = e.Key
		}
		if e.Key > maxKey {
			maxKey = e.Key
		}
		
		data := e.Encode()
		if _, err := writer.Write(data); err != nil {
			return err
		}
		
		index = append(index, SSTableIndex{Key: e.Key, Offset: offset})
		s.keyOffsets[e.Key] = offset
		offset += int64(len(data))
	}
	
	if err := writer.Flush(); err != nil {
		return err
	}
	
	info, err := f.Stat()
	if err != nil {
		return err
	}
	
	s.size = info.Size()
	s.entries = uint32(len(entries))
	s.minKey = minKey
	s.maxKey = maxKey
	s.index = index
	s.loaded = true
	
	return nil
}

func (s *SSTable) loadIndex() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.loaded {
		return nil
	}
	
	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	
	reader := bufio.NewReader(f)
	offset := int64(0)
	index := make([]SSTableIndex, 0)
	s.keyOffsets = make(map[string]int64)
	var minKey, maxKey string
	
	for {
		e, err := DecodeEntry(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		
		if minKey == "" || e.Key < minKey {
			minKey = e.Key
		}
		if maxKey == "" || e.Key > maxKey {
			maxKey = e.Key
		}
		
		size := e.Size()
		index = append(index, SSTableIndex{Key: e.Key, Offset: offset})
		s.keyOffsets[e.Key] = offset
		offset += int64(size)
	}
	
	info, err := f.Stat()
	if err != nil {
		return err
	}
	
	s.size = info.Size()
	s.entries = uint32(len(index))
	s.minKey = minKey
	s.maxKey = maxKey
	s.index = index
	s.loaded = true
	
	return nil
}

func (s *SSTable) Get(key string) (*Entry, bool, error) {
	if err := s.loadIndex(); err != nil {
		return nil, false, err
	}
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	offset, exists := s.keyOffsets[key]
	if !exists {
		return nil, false, nil
	}
	
	f, err := os.Open(s.path)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()
	
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return nil, false, err
	}
	
	reader := bufio.NewReader(f)
	entry, err := DecodeEntry(reader)
	if err != nil {
		return nil, false, err
	}
	
	return entry, true, nil
}

func (s *SSTable) Range(start, end string) ([]*Entry, error) {
	if err := s.loadIndex(); err != nil {
		return nil, err
	}
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	f, err := os.Open(s.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	reader := bufio.NewReader(f)
	results := make([]*Entry, 0)
	
	for {
		e, err := DecodeEntry(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		
		if e.Key >= start && (end == "" || e.Key <= end) {
			results = append(results, e)
		}
		
		if end != "" && e.Key > end {
			break
		}
	}
	
	return results, nil
}

func (s *SSTable) Entries() ([]*Entry, error) {
	if err := s.loadIndex(); err != nil {
		return nil, err
	}
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	f, err := os.Open(s.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	reader := bufio.NewReader(f)
	entries := make([]*Entry, 0)
	
	for {
		e, err := DecodeEntry(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	
	return entries, nil
}

func (s *SSTable) ContainsKeyRange(key string) bool {
	if err := s.loadIndex(); err != nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return key >= s.minKey && key <= s.maxKey
}

func (s *SSTable) OverlapsWith(other *SSTable) bool {
	if err := s.loadIndex(); err != nil {
		return false
	}
	if err := other.loadIndex(); err != nil {
		return false
	}
	
	s.mu.RLock()
	other.mu.RLock()
	defer s.mu.RUnlock()
	defer other.mu.RUnlock()
	
	return !(s.maxKey < other.minKey || other.maxKey < s.minKey)
}

func (s *SSTable) Filename() string {
	return s.filename
}

func (s *SSTable) Path() string {
	return s.path
}

func (s *SSTable) Level() int {
	return s.level
}

func (s *SSTable) Size() int64 {
	if err := s.loadIndex(); err != nil {
		return 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.size
}

func (s *SSTable) EntryCount() uint32 {
	if err := s.loadIndex(); err != nil {
		return 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.entries
}

func (s *SSTable) MinKey() string {
	if err := s.loadIndex(); err != nil {
		return ""
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.minKey
}

func (s *SSTable) MaxKey() string {
	if err := s.loadIndex(); err != nil {
		return ""
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.maxKey
}

func (s *SSTable) Delete() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.Remove(s.path)
}

func LoadSSTableFromPath(path string, level int) (*SSTable, error) {
	filename := filepath.Base(path)
	table := &SSTable{
		filename:   filename,
		path:       path,
		level:      level,
		keyOffsets: make(map[string]int64),
	}
	if err := table.loadIndex(); err != nil {
		return nil, err
	}
	return table, nil
}

func MergeTables(tables []*SSTable, outputLevel int, dataDir string, cleanup func([]*SSTable) error) (*SSTable, error) {
	allEntries := make([]*Entry, 0)
	
	for _, t := range tables {
		entries, err := t.Entries()
		if err != nil {
			return nil, err
		}
		allEntries = append(allEntries, entries...)
	}
	
	merged := MergeEntries(allEntries)
	
	newTable, err := NewSSTable(outputLevel, dataDir)
	if err != nil {
		return nil, err
	}
	
	if err := newTable.Write(merged); err != nil {
		return nil, err
	}
	
	if cleanup != nil {
		if err := cleanup(tables); err != nil {
			return newTable, err
		}
	}
	
	return newTable, nil
}

func MergeEntries(entries []*Entry) []*Entry {
	grouped := make(map[string][]*Entry)
	
	for _, e := range entries {
		grouped[e.Key] = append(grouped[e.Key], e)
	}
	
	merged := make([]*Entry, 0, len(grouped))
	
	for _, versions := range grouped {
		sort.Slice(versions, func(i, j int) bool {
			return versions[i].SeqNum > versions[j].SeqNum
		})
		latest := versions[0]
		merged = append(merged, latest)
	}
	
	sort.Sort(EntryList(merged))
	return merged
}

func ListSSTables(dataDir string) (map[int][]*SSTable, error) {
	levels := make(map[int][]*SSTable)
	
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return levels, nil
		}
		return nil, err
	}
	
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		
		name := e.Name()
		if !strings.HasPrefix(name, "level-") {
			continue
		}
		
		var level int
		if _, err := fmt.Sscanf(name, "level-%d", &level); err != nil {
			continue
		}
		
		levelDir := filepath.Join(dataDir, name)
		files, err := os.ReadDir(levelDir)
		if err != nil {
			return nil, err
		}
		
		tables := make([]*SSTable, 0)
		for _, f := range files {
			if !strings.HasSuffix(f.Name(), ".sst") {
				continue
			}
			path := filepath.Join(levelDir, f.Name())
			table, err := LoadSSTableFromPath(path, level)
			if err != nil {
				return nil, err
			}
			tables = append(tables, table)
		}
		
		levels[level] = tables
	}
	
	return levels, nil
}
