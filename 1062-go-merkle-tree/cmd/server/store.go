package main

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"merkle-tree/pkg/merkle"
)

type TreeStore struct {
	mu    sync.RWMutex
	trees map[string]*merkle.MerkleTree
}

func NewTreeStore() *TreeStore {
	return &TreeStore{
		trees: make(map[string]*merkle.MerkleTree),
	}
}

func (s *TreeStore) Store(tree *merkle.MerkleTree) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateID()
	s.trees[id] = tree
	return id
}

func (s *TreeStore) Get(id string) (*merkle.MerkleTree, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tree, ok := s.trees[id]
	return tree, ok
}

func generateID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
