package unionfind

import (
	"fmt"
)

type UnionFind struct {
	parent map[string]string
	rank   map[string]int
	stack  []operation
	step   int
}

type operation struct {
	step     int
	parent   map[string]string
	rank     map[string]int
}

func New() *UnionFind {
	return &UnionFind{
		parent: make(map[string]string),
		rank:   make(map[string]int),
		stack:  []operation{},
		step:   0,
	}
}

func (uf *UnionFind) AddNode(node string) error {
	if _, exists := uf.parent[node]; exists {
		return fmt.Errorf("node %s already exists", node)
	}
	uf.parent[node] = node
	uf.rank[node] = 1
	uf.step++
	return nil
}

func (uf *UnionFind) Find(a string) (string, error) {
	if _, exists := uf.parent[a]; !exists {
		return "", fmt.Errorf("node %s does not exist", a)
	}
	for uf.parent[a] != a {
		a = uf.parent[a]
	}
	return a, nil
}

func (uf *UnionFind) Union(a, b string) (bool, error) {
	rootA, err := uf.Find(a)
	if err != nil {
		return false, err
	}
	rootB, err := uf.Find(b)
	if err != nil {
		return false, err
	}
	
	uf.step++
	
	if rootA == rootB {
		return false, nil
	}
	
	parentCopy := make(map[string]string)
	for k, v := range uf.parent {
		parentCopy[k] = v
	}
	rankCopy := make(map[string]int)
	for k, v := range uf.rank {
		rankCopy[k] = v
	}
	
	if uf.rank[rootA] < uf.rank[rootB] {
		uf.parent[rootA] = rootB
	} else if uf.rank[rootB] < uf.rank[rootA] {
		uf.parent[rootB] = rootA
	} else {
		uf.parent[rootB] = rootA
		uf.rank[rootA]++
	}
	
	uf.stack = append(uf.stack, operation{
		step: uf.step,
		parent: parentCopy,
		rank: rankCopy,
	})
	
	return true, nil
}

func (uf *UnionFind) Undo() (bool, error) {
	if len(uf.stack) == 0 {
		return false, fmt.Errorf("no union operations to undo")
	}
	
	lastOp := uf.stack[len(uf.stack)-1]
	uf.parent = lastOp.parent
	uf.rank = lastOp.rank
	uf.step = lastOp.step
	uf.stack = uf.stack[:len(uf.stack)-1]
	return true, nil
}

func (uf *UnionFind) UndoTo(step int) (bool, error) {
	if step < 0 {
		return false, fmt.Errorf("step must be non-negative")
	}
	if step >= uf.step {
		return false, nil
	}
	
	for len(uf.stack) > 0 && uf.stack[len(uf.stack)-1].step > step {
		lastOp := uf.stack[len(uf.stack)-1]
		uf.parent = lastOp.parent
		uf.rank = lastOp.rank
		uf.step = lastOp.step
		uf.stack = uf.stack[:len(uf.stack)-1]
	}
	
	if step < uf.step {
		uf.step = step
	}
	
	return true, nil
}

func (uf *UnionFind) Connected(a, b string) (bool, error) {
	rootA, err := uf.Find(a)
	if err != nil {
		return false, err
	}
	rootB, err := uf.Find(b)
	if err != nil {
		return false, err
	}
	return rootA == rootB, nil
}

func (uf *UnionFind) Step() int {
	return uf.step
}
