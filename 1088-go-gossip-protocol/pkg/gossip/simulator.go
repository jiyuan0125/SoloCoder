package gossip

import (
	"math/rand"
	"sync"
)

type Rumor struct {
	Key          string
	Value        string
	Version      int64
	HopCount     int
	ForwardProb  float64
}

type Simulator struct {
	nodes       map[string]*Node
	rumors      map[string]*Rumor
	mu          sync.RWMutex
	round       int
	rng         *rand.Rand
}

func NewSimulator() *Simulator {
	return &Simulator{
		nodes:  make(map[string]*Node),
		rumors: make(map[string]*Rumor),
		rng:    rand.New(rand.NewSource(42)),
	}
}

func (s *Simulator) AddNode(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodes[id] = NewNode(id)
}

func (s *Simulator) RemoveNode(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.nodes, id)
}

func (s *Simulator) InjectData(nodeID, key, value string) {
	s.mu.RLock()
	node, exists := s.nodes[nodeID]
	s.mu.RUnlock()
	if !exists {
		return
	}
	node.Put(key, value)
	
	s.mu.Lock()
	defer s.mu.Unlock()
	storage := node.GetStorage()
	if v, ok := storage[key]; ok {
		rumorKey := nodeID + ":" + key
		s.rumors[rumorKey] = &Rumor{
			Key:          key,
			Value:        v.Value,
			Version:      v.Version,
			HopCount:     0,
			ForwardProb:  1.0,
		}
	}
}

func (s *Simulator) GetNodeIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.nodes))
	for id := range s.nodes {
		ids = append(ids, id)
	}
	return ids
}

func (s *Simulator) GetNode(id string) (*Node, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	node, exists := s.nodes[id]
	return node, exists
}

func (s *Simulator) GetAliveNodes() []*Node {
	s.mu.RLock()
	defer s.mu.RUnlock()
	alive := make([]*Node, 0)
	for _, node := range s.nodes {
		if node.GetStatus() == StatusAlive {
			alive = append(alive, node)
		}
	}
	return alive
}

func (s *Simulator) GetRandomNode(excludeID string) (*Node, bool) {
	alive := s.GetAliveNodes()
	if len(alive) == 0 {
		return nil, false
	}
	
	filtered := make([]*Node, 0, len(alive))
	for _, node := range alive {
		if node.ID != excludeID {
			filtered = append(filtered, node)
		}
	}
	
	if len(filtered) == 0 {
		return nil, false
	}
	
	idx := s.rng.Intn(len(filtered))
	return filtered[idx], true
}

func (s *Simulator) Step() {
	s.mu.Lock()
	s.round++
	currentRound := s.round
	s.mu.Unlock()
	
	aliveNodes := s.GetAliveNodes()
	
	s.performFailureDetection(aliveNodes)
	s.performRumorMongering()
	s.performAntiEntropy(aliveNodes)
	s.updateSuspectNodes()
	
	_ = currentRound
}

func (s *Simulator) performFailureDetection(nodes []*Node) {
	for _, node := range nodes {
		target, exists := s.GetRandomNode(node.ID)
		if !exists {
			continue
		}
		
		success := s.pingNode(target)
		if success {
			target.RecordPingSuccess()
		} else {
			target.RecordPingFailure()
		}
	}
}

func (s *Simulator) pingNode(target *Node) bool {
	return s.rng.Float64() > 0.05
}

func (s *Simulator) performRumorMongering() {
	s.mu.Lock()
	rumorsCopy := make(map[string]*Rumor)
	for k, v := range s.rumors {
		rumorsCopy[k] = &Rumor{
			Key:         v.Key,
			Value:       v.Value,
			Version:     v.Version,
			HopCount:    v.HopCount,
			ForwardProb: v.ForwardProb,
		}
	}
	s.mu.Unlock()
	
	newRumors := make(map[string]*Rumor)
	
	for _, rumor := range rumorsCopy {
		if rumor.HopCount > 0 && s.rng.Float64() > rumor.ForwardProb {
			continue
		}
		
		propagator, exists := s.getRandomAliveNode()
		if !exists {
			continue
		}
		
		propagator.PutWithVersion(rumor.Key, rumor.Value, rumor.Version)
		
		target, exists := s.GetRandomNode(propagator.ID)
		if !exists {
			continue
		}
		
		target.PutWithVersion(rumor.Key, rumor.Value, rumor.Version)
		
		if rumor.ForwardProb > 0.1 {
			newRumor := &Rumor{
				Key:         rumor.Key,
				Value:       rumor.Value,
				Version:     rumor.Version,
				HopCount:    rumor.HopCount + 1,
				ForwardProb: rumor.ForwardProb / 2,
			}
			
			rumorKey := target.ID + ":" + rumor.Key
			newRumors[rumorKey] = newRumor
		}
	}
	
	s.mu.Lock()
	for k, v := range newRumors {
		s.rumors[k] = v
	}
	
	filteredRumors := make(map[string]*Rumor)
	for k, v := range s.rumors {
		if v.ForwardProb > 0.1 {
			filteredRumors[k] = v
		}
	}
	s.rumors = filteredRumors
	s.mu.Unlock()
}

func (s *Simulator) getRandomAliveNode() (*Node, bool) {
	alive := s.GetAliveNodes()
	if len(alive) == 0 {
		return nil, false
	}
	idx := s.rng.Intn(len(alive))
	return alive[idx], true
}

func (s *Simulator) performAntiEntropy(nodes []*Node) {
	for _, node := range nodes {
		if s.rng.Float64() > 0.3 {
			continue
		}
		
		partner, exists := s.GetRandomNode(node.ID)
		if !exists {
			continue
		}
		
		s.exchangeDigests(node, partner)
	}
}

func (s *Simulator) exchangeDigests(a, b *Node) {
	digestA := a.GetDigest()
	digestB := b.GetDigest()
	
	storageA := a.GetStorage()
	storageB := b.GetStorage()
	
	for key, versionB := range digestB {
		if versionA, exists := digestA[key]; !exists || versionB > versionA {
			if v, ok := storageB[key]; ok {
				a.PutWithVersion(key, v.Value, v.Version)
			}
		}
	}
	
	for key, versionA := range digestA {
		if versionB, exists := digestB[key]; !exists || versionA > versionB {
			if v, ok := storageA[key]; ok {
				b.PutWithVersion(key, v.Value, v.Version)
			}
		}
	}
}

func (s *Simulator) updateSuspectNodes() {
	s.mu.RLock()
	nodesCopy := make([]*Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodesCopy = append(nodesCopy, node)
	}
	s.mu.RUnlock()
	
	for _, node := range nodesCopy {
		if node.GetStatus() == StatusSuspect {
			node.IncrementSuspectRound()
		}
	}
}

func (s *Simulator) GetState() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	state := make(map[string]interface{})
	nodesState := make(map[string]interface{})
	
	for id, node := range s.nodes {
		storage := node.GetStorage()
		storageMap := make(map[string]interface{})
		for k, v := range storage {
			storageMap[k] = map[string]interface{}{
				"value":   v.Value,
				"version": v.Version,
			}
		}
		
		nodesState[id] = map[string]interface{}{
			"status":  string(node.GetStatus()),
			"storage": storageMap,
		}
	}
	
	state["round"] = s.round
	state["nodes"] = nodesState
	
	return state
}

func (s *Simulator) GetConsistency() float64 {
	aliveNodes := s.GetAliveNodes()
	if len(aliveNodes) < 2 {
		return 100.0
	}
	
	allKeys := make(map[string]bool)
	for _, node := range aliveNodes {
		for k := range node.GetStorage() {
			allKeys[k] = true
		}
	}
	
	if len(allKeys) == 0 {
		return 100.0
	}
	
	consistentKeys := 0
	for key := range allKeys {
		if s.isKeyConsistent(key, aliveNodes) {
			consistentKeys++
		}
	}
	
	return (float64(consistentKeys) / float64(len(allKeys))) * 100.0
}

func (s *Simulator) isKeyConsistent(key string, nodes []*Node) bool {
	var referenceValue string
	var referenceVersion int64
	first := true
	
	for _, node := range nodes {
		storage := node.GetStorage()
		v, exists := storage[key]
		
		if !exists {
			return false
		}
		
		if first {
			referenceValue = v.Value
			referenceVersion = v.Version
			first = false
		} else {
			if v.Value != referenceValue || v.Version != referenceVersion {
				return false
			}
		}
	}
	
	return true
}

func (s *Simulator) SimulateNodeFailure(nodeID string) {
	s.mu.RLock()
	node, exists := s.nodes[nodeID]
	s.mu.RUnlock()
	if exists {
		node.SimulateFailure()
	}
}

func (s *Simulator) SimulateNodeRecovery(nodeID string) {
	s.mu.RLock()
	node, exists := s.nodes[nodeID]
	s.mu.RUnlock()
	if exists {
		node.SimulateRecovery()
	}
}
