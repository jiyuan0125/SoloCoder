package externalsort

type Stats struct {
	TotalRecords   int64
	ChunksCreated  int
	MergePasses    int
	DiskReads      int64
	DiskWrites     int64
	ChunkSizes     []int
	MergePassSizes []int
}

func newStats() *Stats {
	return &Stats{
		ChunkSizes:     make([]int, 0),
		MergePassSizes: make([]int, 0),
	}
}

func (s *Stats) addChunk(size int) {
	s.ChunksCreated++
	s.ChunkSizes = append(s.ChunkSizes, size)
}

func (s *Stats) addMergePass() {
	s.MergePasses++
}

func (s *Stats) addRead(count int64) {
	s.DiskReads += count
}

func (s *Stats) addWrite(count int64) {
	s.DiskWrites += count
}

func (s *Stats) Clone() *Stats {
	clone := &Stats{
		TotalRecords:   s.TotalRecords,
		ChunksCreated:  s.ChunksCreated,
		MergePasses:    s.MergePasses,
		DiskReads:      s.DiskReads,
		DiskWrites:     s.DiskWrites,
		ChunkSizes:     make([]int, len(s.ChunkSizes)),
		MergePassSizes: make([]int, len(s.MergePassSizes)),
	}
	copy(clone.ChunkSizes, s.ChunkSizes)
	copy(clone.MergePassSizes, s.MergePassSizes)
	return clone
}
