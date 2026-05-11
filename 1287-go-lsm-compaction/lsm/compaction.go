package lsm

import (
	"lsm-tree/common"
	"sort"
	"sync"
)

type CompactionStrategy interface {
	Strategy() common.CompactionStrategy
	ShouldCompact(levels map[int][]*SSTable) bool
	Compact(levels map[int][]*SSTable, dataDir string) (map[int][]*SSTable, int64, error)
	CanAddTable(level int, levels map[int][]*SSTable) bool
	MaxLevels() int
}

type SizeTieredCompaction struct {
	minTables   int
	maxLevel    int
}

func NewSizeTieredCompaction(minTables, maxLevel int) *SizeTieredCompaction {
	return &SizeTieredCompaction{
		minTables: minTables,
		maxLevel:  maxLevel,
	}
}

func (st *SizeTieredCompaction) Strategy() common.CompactionStrategy {
	return common.SizeTieredStrategy
}

func (st *SizeTieredCompaction) MaxLevels() int {
	return st.maxLevel
}

func (st *SizeTieredCompaction) CanAddTable(level int, levels map[int][]*SSTable) bool {
	return true
}

func (st *SizeTieredCompaction) ShouldCompact(levels map[int][]*SSTable) bool {
	for _, tables := range levels {
		if len(tables) >= st.minTables {
			return true
		}
	}
	return false
}

func (st *SizeTieredCompaction) Compact(levels map[int][]*SSTable, dataDir string) (map[int][]*SSTable, int64, error) {
	newLevels := make(map[int][]*SSTable)
	for k, v := range levels {
		newLevels[k] = append([]*SSTable{}, v...)
	}
	
	totalBytesWritten := int64(0)
	
	levelKeys := make([]int, 0, len(newLevels))
	for k := range newLevels {
		levelKeys = append(levelKeys, k)
	}
	sort.Ints(levelKeys)
	
	for _, level := range levelKeys {
		tables := newLevels[level]
		if len(tables) < st.minTables {
			continue
		}
		
		nextLevel := level + 1
		if nextLevel > st.maxLevel {
			continue
		}
		
		toMerge := tables[:st.minTables]
		remaining := tables[st.minTables:]
		
		newTable, err := MergeTables(toMerge, nextLevel, dataDir, func(oldTables []*SSTable) error {
			for _, t := range oldTables {
				if err := t.Delete(); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, 0, err
		}
		
		totalBytesWritten += newTable.Size()
		newLevels[level] = remaining
		newLevels[nextLevel] = append(newLevels[nextLevel], newTable)
	}
	
	return newLevels, totalBytesWritten, nil
}

type LeveledCompaction struct {
	maxFilesPerLevel int
	maxLevel         int
}

func NewLeveledCompaction(maxFilesPerLevel, maxLevel int) *LeveledCompaction {
	return &LeveledCompaction{
		maxFilesPerLevel: maxFilesPerLevel,
		maxLevel:         maxLevel,
	}
}

func (lc *LeveledCompaction) Strategy() common.CompactionStrategy {
	return common.LeveledStrategy
}

func (lc *LeveledCompaction) MaxLevels() int {
	return lc.maxLevel
}

func (lc *LeveledCompaction) CanAddTable(level int, levels map[int][]*SSTable) bool {
	return true
}

func (lc *LeveledCompaction) ShouldCompact(levels map[int][]*SSTable) bool {
	for level, tables := range levels {
		if level == 0 {
			if len(tables) >= lc.maxFilesPerLevel {
				return true
			}
		} else {
			maxFiles := lc.maxFilesPerLevel * (level + 1)
			if len(tables) > maxFiles {
				return true
			}
		}
	}
	return false
}

func (lc *LeveledCompaction) Compact(levels map[int][]*SSTable, dataDir string) (map[int][]*SSTable, int64, error) {
	newLevels := make(map[int][]*SSTable)
	for k, v := range levels {
		newLevels[k] = append([]*SSTable{}, v...)
	}
	
	totalBytesWritten := int64(0)
	
	levelKeys := make([]int, 0, len(newLevels))
	for k := range newLevels {
		levelKeys = append(levelKeys, k)
	}
	sort.Ints(levelKeys)
	
	for _, level := range levelKeys {
		tables := newLevels[level]
		maxFiles := lc.maxFilesPerLevel * (level + 1)
		
		if len(tables) <= maxFiles {
			continue
		}
		
		nextLevel := level + 1
		if nextLevel > lc.maxLevel {
			continue
		}
		
		var toMerge []*SSTable
		if level == 0 {
			toMerge = tables
			newLevels[level] = []*SSTable{}
		} else {
			toMerge = []*SSTable{tables[0]}
			newLevels[level] = tables[1:]
		}
		
		overlapping := lc.findOverlapping(toMerge, newLevels[nextLevel])
		allToMerge := append(toMerge, overlapping...)
		
		nextRemaining := make([]*SSTable, 0)
		for _, t := range newLevels[nextLevel] {
			isOverlapping := false
			for _, o := range overlapping {
				if t == o {
					isOverlapping = true
					break
				}
			}
			if !isOverlapping {
				nextRemaining = append(nextRemaining, t)
			}
		}
		newLevels[nextLevel] = nextRemaining
		
		newTable, err := MergeTables(allToMerge, nextLevel, dataDir, func(oldTables []*SSTable) error {
			for _, t := range oldTables {
				if err := t.Delete(); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, 0, err
		}
		
		totalBytesWritten += newTable.Size()
		newLevels[nextLevel] = append(newLevels[nextLevel], newTable)
	}
	
	newLevels = lc.organizeLeveled(newLevels)
	return newLevels, totalBytesWritten, nil
}

func (lc *LeveledCompaction) findOverlapping(mergeCandidates []*SSTable, levelTables []*SSTable) []*SSTable {
	overlapping := make([]*SSTable, 0)
	for _, table := range mergeCandidates {
		for _, other := range levelTables {
			if table.OverlapsWith(other) {
				alreadyAdded := false
				for _, o := range overlapping {
					if o == other {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					overlapping = append(overlapping, other)
				}
			}
		}
	}
	return overlapping
}

func (lc *LeveledCompaction) organizeLeveled(levels map[int][]*SSTable) map[int][]*SSTable {
	organized := make(map[int][]*SSTable)
	for level, tables := range levels {
		sorted := make([]*SSTable, len(tables))
		copy(sorted, tables)
		
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].MinKey() < sorted[j].MinKey()
		})
		
		organized[level] = sorted
	}
	return organized
}

type CompactionStats struct {
	mu           sync.RWMutex
	totalWrites  int64
	userWrites   int64
}

func NewCompactionStats() *CompactionStats {
	return &CompactionStats{}
}

func (s *CompactionStats) RecordUserWrite(bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userWrites += bytes
	s.totalWrites += bytes
}

func (s *CompactionStats) RecordCompactionWrite(bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalWrites += bytes
}

func (s *CompactionStats) WriteAmplification() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.userWrites == 0 {
		return 1.0
	}
	return float64(s.totalWrites) / float64(s.userWrites)
}

func (s *CompactionStats) GetStats() (total, user int64, wa float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totalWrites, s.userWrites, float64(s.totalWrites) / float64(max(s.userWrites, 1))
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
