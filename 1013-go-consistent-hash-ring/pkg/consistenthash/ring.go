package consistenthash

import (
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"sync"
)

var (
	ErrEmptyRing = errors.New("consistent hash ring is empty")
)

type HashFunc func(data []byte) uint32

type Config struct {
	BaseVirtualNodes int
	HashFunc         HashFunc
}

type Node struct {
	Name   string
	Weight int
}

type HashRing struct {
	mu               sync.RWMutex
	nodes            map[string]*Node
	virtualNodes     []uint32
	virtualNodeMap   map[uint32][]string
	baseVirtualNodes int
	hashFunc         HashFunc
}

func DefaultHashFunc(data []byte) uint32 {
	h := fnv.New32a()
	h.Write(data)
	return h.Sum32()
}

func NewHashRing(config *Config) *HashRing {
	hr := &HashRing{
		nodes:            make(map[string]*Node),
		virtualNodeMap:   make(map[uint32][]string),
		baseVirtualNodes: 150,
		hashFunc:         DefaultHashFunc,
	}
	if config != nil {
		if config.BaseVirtualNodes > 0 {
			hr.baseVirtualNodes = config.BaseVirtualNodes
		}
		if config.HashFunc != nil {
			hr.hashFunc = config.HashFunc
		}
	}
	return hr
}

func (hr *HashRing) AddNode(name string, weight int) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if weight < 1 {
		weight = 1
	}

	if existing, ok := hr.nodes[name]; ok {
		hr.removeNodeInternal(name, existing.Weight)
	}

	hr.nodes[name] = &Node{
		Name:   name,
		Weight: weight,
	}

	virtualCount := hr.baseVirtualNodes * weight
	for i := 0; i < virtualCount; i++ {
		virtualKey := fmt.Sprintf("%s#%d", name, i)
		hash := hr.hashFunc([]byte(virtualKey))
		hr.virtualNodeMap[hash] = append(hr.virtualNodeMap[hash], name)
	}

	hr.rebuildRing()
}

func (hr *HashRing) RemoveNode(name string) bool {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	node, ok := hr.nodes[name]
	if !ok {
		return false
	}

	hr.removeNodeInternal(name, node.Weight)
	hr.rebuildRing()
	return true
}

func (hr *HashRing) removeNodeInternal(name string, weight int) {
	delete(hr.nodes, name)

	virtualCount := hr.baseVirtualNodes * weight
	for i := 0; i < virtualCount; i++ {
		virtualKey := fmt.Sprintf("%s#%d", name, i)
		hash := hr.hashFunc([]byte(virtualKey))

		nodes := hr.virtualNodeMap[hash]
		for j := 0; j < len(nodes); j++ {
			if nodes[j] == name {
				nodes = append(nodes[:j], nodes[j+1:]...)
				j--
			}
		}
		if len(nodes) == 0 {
			delete(hr.virtualNodeMap, hash)
		} else {
			hr.virtualNodeMap[hash] = nodes
		}
	}
}

func (hr *HashRing) rebuildRing() {
	hashes := make([]uint32, 0, len(hr.virtualNodeMap))
	for h := range hr.virtualNodeMap {
		hashes = append(hashes, h)
	}
	sort.Slice(hashes, func(i, j int) bool {
		return hashes[i] < hashes[j]
	})
	hr.virtualNodes = hashes
}

func (hr *HashRing) GetNode(key string) (string, error) {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.virtualNodes) == 0 {
		return "", ErrEmptyRing
	}

	hash := hr.hashFunc([]byte(key))

	idx := sort.Search(len(hr.virtualNodes), func(i int) bool {
		return hr.virtualNodes[i] >= hash
	})

	if idx == len(hr.virtualNodes) {
		idx = 0
	}

	virtualHash := hr.virtualNodes[idx]
	nodeNames := hr.virtualNodeMap[virtualHash]
	return nodeNames[0], nil
}

func (hr *HashRing) GetNodeRanges(name string) ([]HashRange, error) {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if _, ok := hr.nodes[name]; !ok {
		return nil, fmt.Errorf("node %s not found", name)
	}

	ranges := make([]HashRange, 0)
	numVirtual := len(hr.virtualNodes)

	if numVirtual == 0 {
		return ranges, nil
	}

	for i, vh := range hr.virtualNodes {
		nodeNames := hr.virtualNodeMap[vh]
		found := false
		for _, nn := range nodeNames {
			if nn == name {
				found = true
				break
			}
		}
		if !found {
			continue
		}

		prevIdx := i - 1
		if prevIdx < 0 {
			prevIdx = numVirtual - 1
		}
		prevHash := hr.virtualNodes[prevIdx]

		start := prevHash + 1
		end := vh

		if prevHash >= vh {
			ranges = append(ranges, HashRange{Start: start, End: maxUint32})
			ranges = append(ranges, HashRange{Start: 0, End: end})
		} else {
			ranges = append(ranges, HashRange{Start: start, End: end})
		}
	}

	return mergeRanges(ranges), nil
}

const maxUint32 = 1<<32 - 1

type HashRange struct {
	Start uint32
	End   uint32
}

func mergeRanges(ranges []HashRange) []HashRange {
	if len(ranges) <= 1 {
		return ranges
	}

	hasZeroStart := false
	hasMaxEnd := false
	for _, r := range ranges {
		if r.Start == 0 {
			hasZeroStart = true
		}
		if r.End == maxUint32 {
			hasMaxEnd = true
		}
	}

	if hasZeroStart && hasMaxEnd {
		wrapEnd := uint32(0)
		normalStart := uint32(0)
		result := make([]HashRange, 0, len(ranges)-1)

		for _, r := range ranges {
			if r.End == maxUint32 {
				wrapEnd = r.Start
			} else if r.Start == 0 {
				normalStart = r.Start
				result = append(result, HashRange{Start: wrapEnd, End: r.End})
			} else {
				result = append(result, r)
			}
		}

		_ = normalStart
		return result
	}

	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].Start < ranges[j].Start
	})

	result := []HashRange{ranges[0]}
	for i := 1; i < len(ranges); i++ {
		last := &result[len(result)-1]
		if ranges[i].Start <= last.End+1 {
			if ranges[i].End > last.End {
				last.End = ranges[i].End
			}
		} else {
			result = append(result, ranges[i])
		}
	}
	return result
}

func (hr *HashRing) Stats() *RingStats {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	stats := &RingStats{
		TotalVirtualNodes: len(hr.virtualNodes),
		NodeStats:         make(map[string]*NodeStats),
	}

	for name, node := range hr.nodes {
		virtualCount := hr.baseVirtualNodes * node.Weight
		stats.NodeStats[name] = &NodeStats{
			Weight:           node.Weight,
			VirtualNodeCount: virtualCount,
			Percentage:       float64(virtualCount) / float64(len(hr.virtualNodes)) * 100,
		}
	}

	return stats
}

type RingStats struct {
	TotalVirtualNodes int
	NodeStats         map[string]*NodeStats
}

type NodeStats struct {
	Weight           int
	VirtualNodeCount int
	Percentage       float64
}

func (hr *HashRing) hashKey(key string) uint32 {
	return hr.hashFunc([]byte(key))
}

func (hr *HashRing) virtualNodeKey(name string, index int) string {
	return name + "#" + strconv.Itoa(index)
}
