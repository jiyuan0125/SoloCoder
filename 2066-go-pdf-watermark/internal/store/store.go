package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"pdf-watermark/internal/types"
)

type Store struct {
	workDir string
	jobs    map[string]*types.Job
	mu      sync.RWMutex
}

func New(workDir string) (*Store, error) {
	s := &Store{
		workDir: workDir,
		jobs:    make(map[string]*types.Job),
	}
	s.load()
	return s, nil
}

func (s *Store) Close() error {
	return s.save()
}

func (s *Store) SaveJob(job *types.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	return s.save()
}

func (s *Store) GetJob(id string) (*types.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[id]
	if !ok {
		return nil, os.ErrNotExist
	}
	return job, nil
}

func (s *Store) ListJobs() ([]*types.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*types.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		list = append(list, job)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})
	if len(list) > 50 {
		list = list[:50]
	}
	return list, nil
}

func (s *Store) filePath() string {
	return filepath.Join(s.workDir, "jobs.json")
}

func (s *Store) load() {
	data, err := os.ReadFile(s.filePath())
	if err != nil {
		return
	}
	var jobs []*types.Job
	json.Unmarshal(data, &jobs)
	for _, job := range jobs {
		s.jobs[job.ID] = job
	}
}

func (s *Store) save() error {
	list := make([]*types.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		list = append(list, job)
	}
	data, _ := json.MarshalIndent(list, "", "  ")
	return os.WriteFile(s.filePath(), data, 0644)
}
