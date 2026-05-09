package fnv

import (
	"fmt"
	"sort"
	"strconv"
)

type ConsistentHash struct {
	circle    map[uint32]string
	keys      []uint32
	replicas  int
	nodeSet   map[string]bool
}

func NewConsistentHash(replicas int) *ConsistentHash {
	return &ConsistentHash{
		circle:   make(map[uint32]string),
		keys:     []uint32{},
		replicas: replicas,
		nodeSet:  make(map[string]bool),
	}
}

func (ch *ConsistentHash) hashNode(node string, replica int) (uint32, error) {
	key := fmt.Sprintf("%s%d", node, replica)
	return Hash32([]byte(key))
}

func (ch *ConsistentHash) Add(node string) error {
	if ch.nodeSet[node] {
		return nil
	}
	for i := 0; i < ch.replicas; i++ {
		h, err := ch.hashNode(node, i)
		if err != nil {
			return err
		}
		ch.circle[h] = node
		ch.keys = append(ch.keys, h)
	}
	sort.Slice(ch.keys, func(i, j int) bool {
		return ch.keys[i] < ch.keys[j]
	})
	ch.nodeSet[node] = true
	return nil
}

func (ch *ConsistentHash) Remove(node string) {
	if !ch.nodeSet[node] {
		return
	}
	newKeys := []uint32{}
	for i := 0; i < ch.replicas; i++ {
		h, err := ch.hashNode(node, i)
		if err != nil {
			continue
		}
		delete(ch.circle, h)
	}
	for _, k := range ch.keys {
		if _, exists := ch.circle[k]; exists {
			newKeys = append(newKeys, k)
		}
	}
	ch.keys = newKeys
	delete(ch.nodeSet, node)
}

func (ch *ConsistentHash) Get(key string) (string, error) {
	if len(ch.nodeSet) == 0 {
		return "", fmt.Errorf("no nodes available")
	}
	h, err := Hash32([]byte(key))
	if err != nil {
		return "", err
	}
	idx := sort.Search(len(ch.keys), func(i int) bool {
		return ch.keys[i] >= h
	})
	if idx == len(ch.keys) {
		idx = 0
	}
	return ch.circle[ch.keys[idx]], nil
}

func (ch *ConsistentHash) Nodes() []string {
	nodes := make([]string, 0, len(ch.nodeSet))
	for n := range ch.nodeSet {
		nodes = append(nodes, n)
	}
	return nodes
}

func (ch *ConsistentHash) Has(node string) bool {
	return ch.nodeSet[node]
}

func ParseIntToUint32(s string) (uint32, error) {
	val, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(val), nil
}
