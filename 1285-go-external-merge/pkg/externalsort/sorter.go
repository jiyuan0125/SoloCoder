package externalsort

import (
	"sort"
	"sync"
)

const (
	defaultMemoryLimit  = 1000
	defaultMergeWays    = 2
	minMemoryLimit      = 1
	minMergeWays        = 2
)

type ExternalSorter struct {
	mu           sync.Mutex
	memoryLimit  int
	mergeWays    int
	data         []int64
	chunks       []*chunk
	finalResult  []int64
	stats        *Stats
	sorted       bool
	inProgress   bool
}

func New() *ExternalSorter {
	return &ExternalSorter{
		memoryLimit: defaultMemoryLimit,
		mergeWays:   defaultMergeWays,
		data:        make([]int64, 0),
		chunks:      make([]*chunk, 0),
		stats:       newStats(),
	}
}

func (s *ExternalSorter) SetMemoryLimit(limit int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if limit <= 0 {
		limit = defaultMemoryLimit
	}
	s.memoryLimit = limit
}

func (s *ExternalSorter) GetMemoryLimit() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.memoryLimit
}

func (s *ExternalSorter) SetMergeWays(ways int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ways <= 1 {
		ways = defaultMergeWays
	}
	s.mergeWays = ways
}

func (s *ExternalSorter) GetMergeWays() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mergeWays
}

func (s *ExternalSorter) AddData(values []int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = append(s.data, values...)
	s.stats.TotalRecords += int64(len(values))
	s.sorted = false
	s.chunks = nil
	s.finalResult = nil
	s.stats = newStats()
	s.stats.TotalRecords = int64(len(s.data))
}

func (s *ExternalSorter) GetData() []int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]int64(nil), s.data...)
}

func (s *ExternalSorter) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make([]int64, 0)
	s.chunks = make([]*chunk, 0)
	s.finalResult = nil
	s.stats = newStats()
	s.sorted = false
	s.inProgress = false
}

func (s *ExternalSorter) GetStats() *Stats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stats.Clone()
}

func (s *ExternalSorter) GetFinalResult() []int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.finalResult == nil {
		return nil
	}
	return append([]int64(nil), s.finalResult...)
}

func (s *ExternalSorter) IsSorted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sorted
}

func (s *ExternalSorter) Sort() error {
	s.mu.Lock()
	if s.inProgress {
		s.mu.Unlock()
		return nil
	}
	s.inProgress = true
	s.chunks = make([]*chunk, 0)
	s.stats = newStats()
	s.stats.TotalRecords = int64(len(s.data))
	s.sorted = false
	s.finalResult = nil
	memoryLimit := s.memoryLimit
	mergeWays := s.mergeWays
	data := s.data
	s.mu.Unlock()

	var result []int64
	var err error

	if len(data) <= memoryLimit {
		result, err = s.sortInMemory(data)
	} else {
		result, err = s.sortExternally(data, memoryLimit, mergeWays)
	}

	if err != nil {
		s.mu.Lock()
		s.inProgress = false
		s.mu.Unlock()
		return err
	}

	s.mu.Lock()
	s.finalResult = result
	s.sorted = true
	s.inProgress = false
	s.mu.Unlock()

	return nil
}

func (s *ExternalSorter) sortInMemory(data []int64) ([]int64, error) {
	if len(data) == 0 {
		return []int64{}, nil
	}
	s.mu.Lock()
	s.stats.addWrite(int64(len(data)))
	s.mu.Unlock()

	result := make([]int64, len(data))
	copy(result, data)
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	s.mu.Lock()
	s.stats.addRead(int64(len(data)))
	s.mu.Unlock()

	return result, nil
}

func (s *ExternalSorter) sortExternally(data []int64, memoryLimit int, mergeWays int) ([]int64, error) {
	if len(data) == 0 {
		return []int64{}, nil
	}

	chunks, err := s.createChunks(data, memoryLimit)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.chunks = chunks
	s.mu.Unlock()

	result, err := s.mergeChunks(chunks, mergeWays)
	return result, err
}

func (s *ExternalSorter) createChunks(data []int64, memoryLimit int) ([]*chunk, error) {
	chunks := make([]*chunk, 0)

	for i := 0; i < len(data); i += memoryLimit {
		end := i + memoryLimit
		if end > len(data) {
			end = len(data)
		}

		chunkData := make([]int64, end-i)
		copy(chunkData, data[i:end])

		sort.Slice(chunkData, func(a, b int) bool {
			return chunkData[a] < chunkData[b]
		})

		chunks = append(chunks, newChunk(chunkData))

		s.mu.Lock()
		s.stats.addChunk(len(chunkData))
		s.stats.addWrite(int64(len(chunkData)))
		s.stats.addRead(int64(len(chunkData)))
		s.mu.Unlock()
	}

	return chunks, nil
}

func (s *ExternalSorter) mergeChunks(chunks []*chunk, ways int) ([]int64, error) {
	if len(chunks) == 0 {
		return []int64{}, nil
	}

	currentChunks := make([]*chunk, len(chunks))
	for i, c := range chunks {
		currentChunks[i] = newChunk(c.records)
	}

	for len(currentChunks) > 1 {
		nextChunks := make([]*chunk, 0)

		for i := 0; i < len(currentChunks); i += ways {
			end := i + ways
			if end > len(currentChunks) {
				end = len(currentChunks)
			}

			group := currentChunks[i:end]
			merged := s.mergeGroup(group)

			nextChunks = append(nextChunks, merged)

			s.mu.Lock()
			s.stats.addRead(int64(merged.size()))
			s.stats.addWrite(int64(merged.size()))
			s.mu.Unlock()
		}

		currentChunks = nextChunks

		s.mu.Lock()
		s.stats.addMergePass()
		s.stats.MergePassSizes = append(s.stats.MergePassSizes, len(nextChunks))
		s.mu.Unlock()
	}

	if len(currentChunks) == 0 {
		return []int64{}, nil
	}

	return currentChunks[0].records, nil
}

func (s *ExternalSorter) mergeGroup(chunks []*chunk) *chunk {
	if len(chunks) == 0 {
		return newChunk([]int64{})
	}
	if len(chunks) == 1 {
		return newChunk(chunks[0].records)
	}

	h := newMinHeap()
	activeChunks := make([]*chunk, len(chunks))
	copy(activeChunks, chunks)

	for i, c := range activeChunks {
		if val, ok := c.peek(); ok {
			h.PushItem(val, i)
		}
	}

	result := make([]int64, 0)

	for h.Len() > 0 {
		val, chunkID, ok := h.PopItem()
		if !ok {
			break
		}

		result = append(result, val)
		activeChunks[chunkID].next()

		if nextVal, ok := activeChunks[chunkID].peek(); ok {
			h.PushItem(nextVal, chunkID)
		}
	}

	return newChunk(result)
}

func (s *ExternalSorter) GetMergePassDetails() [][]int {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.chunks) == 0 {
		return nil
	}

	passDetails := make([][]int, 0)
	passDetails = append(passDetails, make([]int, len(s.chunks)))
	for i, c := range s.chunks {
		passDetails[0][i] = c.size()
	}

	ways := s.mergeWays
	currentSize := len(s.chunks)

	for currentSize > 1 {
		nextPass := make([]int, 0)
		for i := 0; i < currentSize; i += ways {
			end := i + ways
			if end > currentSize {
				end = currentSize
			}
			groupSize := 0
			for j := i; j < end; j++ {
				groupSize += s.chunks[j].size()
			}
			nextPass = append(nextPass, groupSize)
		}
		passDetails = append(passDetails, nextPass)
		currentSize = len(nextPass)
	}

	return passDetails
}
