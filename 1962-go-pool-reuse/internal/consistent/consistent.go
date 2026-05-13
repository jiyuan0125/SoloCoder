package consistent

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
	"strconv"
	"sync"
)

type HashFunc func(data []byte) uint64

func SHA256(data []byte) uint64 {
	h := sha256.Sum256(data)
	return binary.BigEndian.Uint64(h[:8])
}

type Consistent struct {
	hash        HashFunc
	virtualNode int
	replicas    map[uint64]int
	sortedKeys  []uint64
	nodes       []int
	mu          sync.RWMutex
}

func New(virtualNode int, hash HashFunc) *Consistent {
	if hash == nil {
		hash = SHA256
	}
	if virtualNode <= 0 {
		virtualNode = 100
	}
	c := &Consistent{
		hash:        hash,
		virtualNode: virtualNode,
		replicas:    make(map[uint64]int),
		nodes:       []int{},
	}
	return c
}

func (c *Consistent) AddNodes(nodes []int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, node := range nodes {
		c.addNode(node)
	}
	c.sortKeys()
}

func (c *Consistent) AddNode(node int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.addNode(node)
	c.sortKeys()
}

func (c *Consistent) addNode(node int) {
	for i := 0; i < c.virtualNode; i++ {
		key := fmt.Sprintf("%d:%d", node, i)
		hash := c.hash([]byte(key))
		c.replicas[hash] = node
	}
	c.nodes = append(c.nodes, node)
}

func (c *Consistent) RemoveNode(node int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < c.virtualNode; i++ {
		key := fmt.Sprintf("%d:%d", node, i)
		hash := c.hash([]byte(key))
		delete(c.replicas, hash)
	}

	for i, n := range c.nodes {
		if n == node {
			c.nodes = append(c.nodes[:i], c.nodes[i+1:]...)
			break
		}
	}
	c.sortKeys()
}

func (c *Consistent) sortKeys() {
	c.sortedKeys = make([]uint64, 0, len(c.replicas))
	for k := range c.replicas {
		c.sortedKeys = append(c.sortedKeys, k)
	}
	sort.Slice(c.sortedKeys, func(i, j int) bool {
		return c.sortedKeys[i] < c.sortedKeys[j]
	})
}

func (c *Consistent) Get(key string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.sortedKeys) == 0 {
		return -1
	}

	hash := c.hash([]byte(key))

	idx := sort.Search(len(c.sortedKeys), func(i int) bool {
		return c.sortedKeys[i] >= hash
	})

	if idx == len(c.sortedKeys) {
		idx = 0
	}

	return c.replicas[c.sortedKeys[idx]]
}

func (c *Consistent) Nodes() []int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	nodes := make([]int, len(c.nodes))
	copy(nodes, c.nodes)
	return nodes
}

func (c *Consistent) NodeCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.nodes)
}

func hashKey(data []byte) string {
	return strconv.FormatUint(SHA256(data), 10)
}
